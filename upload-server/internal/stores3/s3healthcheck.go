package stores3

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

type S3HealthCheck struct {
	Client *s3.Client
	BucketName string
}

func (c *S3HealthCheck) Health(ctx context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.TUS_STORAGE_HEALTH_PREFIX

	if c.Client == nil {
		shr.Service = models.STATUS_DOWN
		shr.HealthIssue = "S3 client not available"
		return shr
	}

	_, err := c.Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: &c.BucketName,
	})

	if err != nil {
		shr.Service = models.STATUS_DOWN
		shr.HealthIssue = err.Error()
		return shr
	}

	shr.Service += " S3 Bucket " + c.BucketName
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}
