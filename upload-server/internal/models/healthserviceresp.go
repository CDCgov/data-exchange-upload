package models

type ServiceHealthResp struct {
	Service     string `json:"service"`
	Status      string `json:"status"`
	HealthIssue string `json:"health_issue"`
} // .ServiceHealthResp
