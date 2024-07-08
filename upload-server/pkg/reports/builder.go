package reports

import (
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"time"
)

type Type string
type Report struct {
	ReportSchemaVersion string          `json:"report_schema_version"`
	UploadID            string          `json:"upload_id"`
	DataStreamID        string          `json:"data_stream_id"`
	DataStreamRoute     string          `json:"data_stream_route"`
	Jurisdiction        string          `json:"jurisdiction"`
	DexIngestDatetime   string          `json:"dex_ingest_datetime"`
	ContentType         string          `json:"content_type"`
	DispositionType     string          `json:"disposition_type"`
	StageInfo           ReportStageInfo `json:"stage_info"`
	Content             any             `json:"content"` // TODO: Can we limit this to a specific type (i.e. ReportContent or UploadStatusTYpe type?
}

type ReportStageInfo struct {
	Service          string   `json:"service"`
	Stage            string   `json:"stage"`
	Version          string   `json:"version"`
	Status           string   `json:"status"`
	Issues           []string `json:"issues"`
	StartProcessTime string   `json:"start_process_time"`
	EndProcessTime   string   `json:"end_process_time"`
}

func (r *Report) Identifier() string {
	return r.UploadID
}

type ReportContent struct {
	SchemaName    string `json:"schema_name"`
	SchemaVersion string `json:"schema_version"`
}

type UploadLifecycleContent struct {
	ReportContent
	Status string `json:"status"`
}

type MetaDataVerifyContent struct {
	ReportContent
	Filename string `json:"filename"`
	Metadata any    `json:"metadata"`
}
type Builder interface {
	SetUploadId(string)
	SetManifest(map[string]string)
	AppendIssue(string)
	SetStatus(string)
	SetStartTime(time.Time)
	SetEndTime(time.Time)
	SetDispositionType(string)
	Build(stage string) (*Report, error)
}

func NewBuilder(version string) Builder {
	switch version {
	default:
		return &ReportBuilder{}
	}
}

type ReportBuilder struct {
	Version         string
	UploadId        string
	Manifest        map[string]string
	Issues          []string
	Status          string
	StartTime       time.Time
	EndTime         time.Time
	DispositionType string
}

func (b *ReportBuilder) SetUploadId(id string) {
	b.UploadId = id
}

func (b *ReportBuilder) SetManifest(m map[string]string) {
	b.Manifest = m
}

func (b *ReportBuilder) SetStatus(s string) {
	b.Status = s
}

func (b *ReportBuilder) AppendIssue(i string) {
	b.Issues = append(b.Issues, i)
}

func (b *ReportBuilder) SetStartTime(t time.Time) {
	b.StartTime = t
}

func (b *ReportBuilder) SetEndTime(t time.Time) {
	b.EndTime = t
}

func (b *ReportBuilder) SetDispositionType(d string) {
	b.DispositionType = d
}

func (b *ReportBuilder) Build(stage string) (*Report, error) {
	r := &Report{
		UploadID:          b.UploadId,
		DataStreamID:      metadata.GetDataStreamID(b.Manifest),
		DataStreamRoute:   metadata.GetDataStreamRoute(b.Manifest),
		Jurisdiction:      metadata.GetJurisdiction(b.Manifest),
		DexIngestDatetime: metadata.GetDexIngestDatetime(b.Manifest),
		ContentType:       "application/json",
		DispositionType:   b.DispositionType,
		StageInfo: ReportStageInfo{
			Issues:           b.Issues,
			Stage:            stage,
			Service:          "", // TODO get from version package
			Version:          "", // TODO get from version package
			Status:           b.Status,
			StartProcessTime: b.StartTime.String(),
			EndProcessTime:   b.EndTime.String(),
		},
	}

	switch stage {
	case "dex-metadata-verify":
		c := createMetadataVerifyContent(*b)
		r.Content = c

		return r, nil
	}

	return nil, fmt.Errorf("could not build report for stage %s", stage)
}

func createMetadataVerifyContent(b ReportBuilder) MetaDataVerifyContent {
	return MetaDataVerifyContent{
		ReportContent: ReportContent{
			SchemaVersion: b.Version,
			SchemaName:    "dex-metadata-verify",
		},
		Filename: metadata.GetFilename(b.Manifest),
		Metadata: b.Manifest,
	}
}

// TODO load stage info internally
