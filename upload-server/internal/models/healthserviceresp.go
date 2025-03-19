package models

type ServiceHealthResp struct {
	Service     string `json:"service"`
	Status      string `json:"status"`
	HealthIssue string `json:"health_issue"`
} // .ServiceHealthResp

func (shr ServiceHealthResp) BuildErrorResponse(err error) ServiceHealthResp {
	shr.Status = STATUS_DOWN
	shr.HealthIssue = err.Error()
	return shr
}
