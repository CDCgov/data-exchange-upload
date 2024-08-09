package stores3

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

type S3HealthCheck struct {
	Client *s3.Client
}

func (c *S3HealthCheck) Health(ctx context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.TUS_STORAGE_HEALTH_PREFIX

	if c.Client == nil {
		shr.Service = models.STATUS_DOWN
		shr.HealthIssue = "S3 client not available"
		return shr
	}
	testBucket := "tests3bucket"
	_, err := c.Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &testBucket,
	})
	var respError *types.BucketAlreadyExists
	if errors.As(err, &respError) {
		shr.Status = models.STATUS_UP
		shr.HealthIssue = models.HEALTH_ISSUE_NONE
		return shr
	}
	if err != nil {
		shr.Service = models.STATUS_DOWN
		shr.HealthIssue = err.Error()
		return shr
	}

	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}
