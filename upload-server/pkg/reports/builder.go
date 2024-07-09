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

// TODO make sure this matches schema
type UploadStatusContent struct {
	ReportContent
	Filename string `json:"filename"`
	Tguid    string `json:"tguid"`
	Offset   string `json:"offset"`
	Size     string `json:"size"`
}

// TODO return builder interface for chaining
type Builder interface {
	SetStage(string)
	SetUploadId(string)
	SetManifest(map[string]string)
	AppendIssue(string)
	SetStatus(string)
	SetStartTime(time.Time)
	SetEndTime(time.Time)
	SetDispositionType(string)
	SetContentBuilder(ContentBuilder)
	Build() *Report
}

func NewBuilder(version string, stage string, uploadId string, manifest map[string]string, dispType string, contentBuilder ContentBuilder) Builder {
	return &ReportBuilder{
		Version:         version,
		Stage:           stage,
		UploadId:        uploadId,
		Manifest:        manifest,
		DispositionType: dispType,
		ContentBuilder:  contentBuilder,
	}
}

type ReportBuilder struct {
	Stage           string // TODO maybe init in constructor only
	Version         string
	UploadId        string
	Manifest        map[string]string
	Issues          []string
	Status          string
	StartTime       time.Time
	EndTime         time.Time
	DispositionType string
	ContentBuilder  ContentBuilder
}

func (b *ReportBuilder) SetStage(s string) {
	b.Stage = s
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

func (b *ReportBuilder) SetContentBuilder(cb ContentBuilder) {
	b.ContentBuilder = cb
}

func (b *ReportBuilder) Build() *Report {
	switch b.Version {
	default:
		return &Report{
			UploadID:          b.UploadId,
			DataStreamID:      metadata.GetDataStreamID(b.Manifest),
			DataStreamRoute:   metadata.GetDataStreamRoute(b.Manifest),
			Jurisdiction:      metadata.GetJurisdiction(b.Manifest),
			DexIngestDatetime: metadata.GetDexIngestDatetime(b.Manifest),
			ContentType:       "application/json",
			DispositionType:   b.DispositionType,
			StageInfo: ReportStageInfo{
				Issues:           b.Issues,
				Stage:            b.Stage,
				Service:          "", // TODO get from version package
				Version:          "", // TODO get from version package
				Status:           b.Status,
				StartProcessTime: b.StartTime.String(),
				EndProcessTime:   b.EndTime.String(),
			},
			Content: b.ContentBuilder.Build(),
		}
	}
}

func NewMetadataVerifyContentBuilder(version string) *MetadataVerifyContentBuilder {
	return &MetadataVerifyContentBuilder{Version: version}
}

func NewUploadStatusContentBuilder(version string) *UploadStatusContentBuilder {
	return &UploadStatusContentBuilder{Version: version}
}

type ContentBuilder interface {
	SetVersion(string)
	SetContent(any) error
	Build() any
}

type MetadataVerifyContentBuilder struct {
	Version string
	Content MetaDataVerifyContent
}

func (b *MetadataVerifyContentBuilder) SetVersion(v string) {
	b.Version = v
}

func (b *MetadataVerifyContentBuilder) SetContent(c any) error {
	mvc, ok := c.(MetaDataVerifyContent)
	if !ok {
		return fmt.Errorf("bad content")
	}

	mvc.SchemaName = "dex-metadata-verify"
	mvc.SchemaVersion = b.Version
	b.Content = mvc
	return nil
}

func (b *MetadataVerifyContentBuilder) Build() any {
	switch b.Version {
	default:
		return b.Content
	}
}

type UploadStatusContentBuilder struct {
	Version string
	Content UploadStatusContent
}

func (b *UploadStatusContentBuilder) SetVersion(v string) {
	b.Version = v
}

func (b *UploadStatusContentBuilder) SetContent(c any) error {
	usc, ok := c.(UploadStatusContent)
	if !ok {
		// TODO make error var
		return fmt.Errorf("bad content")
	}

	usc.SchemaName = "upload"
	usc.SchemaVersion = b.Version
	b.Content = usc
	return nil
}

func (b *UploadStatusContentBuilder) Build() any {
	return b.Content
}
