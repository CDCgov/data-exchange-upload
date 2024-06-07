package redislockerhealth

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/redis/go-redis/v9"
)

type RedisLockerHealth struct {
	client *redis.Client
}

func New(uri string) (*RedisLockerHealth, error) {

	connection, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(connection)
	if res := client.Ping(context.Background()); res.Err() != nil {
		return nil, res.Err()
	}

	return &RedisLockerHealth{
		client: client,
	}, nil

} // .New

func (redisLockerHealth RedisLockerHealth) Health(ctx context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.REDIS_LOCKER

	// Ping redis service
	client := redisLockerHealth.client
	if res := client.Ping(context.Background()); res.Err() != nil {
		return redisLockerDown(res.Err())
	}

	// all good
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}

func redisLockerDown(err error) models.ServiceHealthResp {
	return models.ServiceHealthResp{
		Service:     models.REDIS_LOCKER,
		Status:      models.STATUS_DOWN,
		HealthIssue: err.Error(),
	}
}
