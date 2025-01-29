package redislocker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/tus/tusd/v2/pkg/handler"
)

var (
	LockExchangeChannel = "tusd_lock_release_request_%s"
	LockReleaseChannel  = "tusd_lock_released_%s"
	LockExpiry          = 8 * time.Second
)

type LockerOption func(l *RedisLocker)

func WithLogger(logger *slog.Logger) LockerOption {
	return func(l *RedisLocker) {
		l.logger = logger
	}
}

func WithMutexCreator(uri string) (LockerOption, error) {
	connection, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	rsm := &RedSyncMutex{
		conn: connection,
	}

	return func(l *RedisLocker) {
		l.CreateMutex = rsm.CreateMutex
	}, nil
}

func New(uri string, lockerOptions ...LockerOption) (*RedisLocker, error) {
	defaultCreator, err := WithMutexCreator(uri)
	if err != nil {
		return nil, err
	}
	defaultOpts := []LockerOption{
		WithLogger(slog.Default()), defaultCreator,
	}

	connection, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(connection)

	if res := client.Ping(context.Background()); res.Err() != nil {
		return nil, res.Err()
	}

	locker := &RedisLocker{
		redis: client,
	}

	for _, option := range defaultOpts {
		option(locker)
	}

	for _, option := range lockerOptions {
		option(locker)
	}

	if locker.CreateMutex == nil || locker.logger == nil {
		return nil, fmt.Errorf("missing required properties for locker %v", *locker)
	}

	return locker, nil
}

type RedSyncMutex struct {
	rs   *redsync.Redsync
	conn *redis.Options
}

func (rsm *RedSyncMutex) CreateMutex(id string) MutexLock {
	// Recreate client if the pool lost connection.
	if rsm.rs == nil {
		client := redis.NewClient(rsm.conn)
		rsm.rs = redsync.New(goredis.NewPool(client))
	}
	return rsm.rs.NewMutex(id, redsync.WithExpiry(LockExpiry))
}

type LockExchange interface {
	Listen(ctx context.Context, id string, callback func())
	Request(ctx context.Context, id string) error
}

type BidirectionalLockExchange interface {
	LockExchange
	ReleaseChannel(ctx context.Context, id string) <-chan *redis.Message
	Release(ctx context.Context, id string) error
}

type RedisLockExchange struct {
	client *redis.Client
}

func (e *RedisLockExchange) Listen(ctx context.Context, id string, callback func()) {
	psub := e.client.PSubscribe(ctx, fmt.Sprintf(LockExchangeChannel, id))
	defer psub.Close()
	c := psub.Channel()
	select {
	case <-c:
		callback()
		return
	case <-ctx.Done():
		return
	}
}

func (e *RedisLockExchange) ReleaseChannel(ctx context.Context, id string) <-chan *redis.Message {
	psub := e.client.PSubscribe(ctx, fmt.Sprintf(LockReleaseChannel, id))
	releaseMessages := make(chan *redis.Message)
	c := psub.Channel()
	go func() {
		defer psub.Close()
		<-c
		close(releaseMessages)
	}()
	return releaseMessages
}

func (e *RedisLockExchange) Request(ctx context.Context, id string) error {
	res := e.client.Publish(ctx, fmt.Sprintf(LockExchangeChannel, id), id)
	return res.Err()
}

func (e *RedisLockExchange) Release(ctx context.Context, id string) error {
	res := e.client.Publish(ctx, fmt.Sprintf(LockReleaseChannel, id), id)
	return res.Err()
}

type MutexLock interface {
	TryLockContext(context.Context) error
	ExtendContext(context.Context) (bool, error)
	UnlockContext(context.Context) (bool, error)
	Until() time.Time
}

type RedisLocker struct {
	//rs          *redsync.Redsync
	CreateMutex func(id string) MutexLock
	redis       *redis.Client
	logger      *slog.Logger
}

func (locker *RedisLocker) UseIn(composer *handler.StoreComposer) {
	composer.UseLocker(locker)
}

func (locker *RedisLocker) NewLock(id string) (handler.Lock, error) {
	mutex := locker.CreateMutex(id)
	return &redisLock{
		id:    id,
		mutex: mutex,
		exchange: &RedisLockExchange{
			client: locker.redis,
		},
		logger: locker.logger.With("uploadId", id),
	}, nil
}

func (locker *RedisLocker) Health(_ context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.REDIS_LOCKER

	// Ping redis service
	client := locker.redis
	if res := client.Ping(context.Background()); res.Err() != nil {
		return models.ServiceHealthResp{
			Service:     models.REDIS_LOCKER,
			Status:      models.STATUS_DOWN,
			HealthIssue: res.Err().Error(),
		}
	}

	// all good
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}

type redisLock struct {
	id       string
	mutex    MutexLock
	ctx      context.Context
	cancel   func()
	exchange BidirectionalLockExchange
	logger   *slog.Logger
}

func (l *redisLock) Lock(ctx context.Context, releaseRequested func()) error {
	l.logger.Debug("locking upload")
	if err := l.requestLock(ctx); err != nil {
		return err
	}
	go l.exchange.Listen(l.ctx, l.id, releaseRequested)
	go func() {
		if err := l.keepAlive(l.ctx); err != nil {
			l.cancel()
			if releaseRequested != nil {
				releaseRequested()
			}
		}
	}()
	l.logger.Debug("locked upload")
	return nil
}

func (l *redisLock) aquireLock(ctx context.Context) error {
	if err := l.mutex.TryLockContext(ctx); err != nil {
		// Currently there aren't any errors
		// defined by redsync we don't want to retry.
		// If there are any return just that error without
		// handler.ErrFileLocked to show it's non-recoverable.
		return errors.Join(err, handler.ErrFileLocked)
	}

	l.ctx, l.cancel = context.WithCancel(context.Background())

	return nil
}

func (l *redisLock) requestLock(ctx context.Context) error {
	err := l.aquireLock(ctx)
	if err == nil {
		return nil
	}
	var errs error
	c := l.exchange.ReleaseChannel(ctx, l.id)
	if err := l.exchange.Request(ctx, l.id); err != nil {
		return err
	}
	if !errors.Is(err, handler.ErrFileLocked) {
		return err
	}
	errs = errors.Join(errs, err)
	select {
	case <-c:
		l.logger.Debug("notified of lock release")
		return l.aquireLock(ctx)
	case <-ctx.Done():
		return errors.Join(errs, handler.ErrLockTimeout)
	}
}

func (l *redisLock) keepAlive(ctx context.Context) error {
	//insures that an extend will be canceled if it's unlocked in the middle of an attempt
	for {
		select {
		case <-time.After(time.Until(l.mutex.Until()) / 2):
			l.logger.Debug("extend lock attempt started", "time", time.Now())
			_, err := l.mutex.ExtendContext(ctx)
			if err != nil {
				l.logger.Error("failed to extend lock", "time", time.Now(), "error", err)
				return err
			}
			l.logger.Debug("lock extended", "time", time.Now())
		case <-ctx.Done():
			l.logger.Debug("lock was closed")
			return nil
		}
	}
}

func (l *redisLock) Unlock() error {
	l.logger.Debug("unlocking upload")
	if l.cancel != nil {
		defer l.cancel()
	}
	b, err := l.mutex.UnlockContext(l.ctx)
	if !b {
		l.logger.Error("failed to release lock", "err", err)
	}
	l.logger.Debug("notifying of lock release")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if e := l.exchange.Release(ctx, l.id); e != nil {
		err = errors.Join(err, e)
	}
	if err != nil {
		l.logger.Error("errors while unlocking", "err", err)
	}
	return err
}
