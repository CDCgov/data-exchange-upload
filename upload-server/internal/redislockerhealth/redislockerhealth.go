package redislocker

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"

	//"github.com/cdcgov/data-exchange-upload/upload-server/pkg/redislocker"
	//"github.com/tus/tusd/v2/pkg/memorylocker"
	//"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"net/url"
)

type RedislockerHealth struct {
	EndpointHealth    string
	EndpointHealthURI *url.URL
} // .PsHealth

func New(appConfig appconfig.AppConfig) (*RedislockerHealth, error) {
	// initialize locker
	//var locker handlertusd.Locker = memorylocker.New()

	endpointURI, err := url.ParseRequestURI(appConfig.TusRedisLockURI)
	if err != nil {
		return nil, err
	} // .if

	return &RedislockerHealth{
		EndpointHealth:    appConfig.TusRedisLockURI,
		EndpointHealthURI: endpointURI,
	}, nil // .return

} // .New

func (rdlHealth RedislockerHealth) Health(ctx context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.REDIS_LOCKER

	// TODO: Get the redis locker instance.

	// TODO: Check if redis server starts.

	// all good
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}

func redisLockerDown(err error) models.ServiceHealthResp {
	return models.ServiceHealthResp{
		Service:     models.SERVICE_BUS,
		Status:      models.STATUS_DOWN,
		HealthIssue: err.Error(),
	}
}
