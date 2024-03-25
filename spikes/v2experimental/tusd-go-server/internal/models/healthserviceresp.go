package models

type HealthServiceResp struct {
	Service     string `json:"service"`
	Status      string `json:"status"`
	HealthIssue string `json:"health_issue"`
} // .HealthServiceResp
