package reports

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/version"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"time"
)

const StageMetadataVerify = "dex-metadata-verify"
const StageMetadataTransform = "dex-metadata-transform"
const StageFileCopy = "dex-file-copy"
const StageUploadStatus = "dex-upload-status"
const StageUploadStarted = "dex-upload-started"
const StageUploadCompleted = "dex-upload-completed"
const DispositionTypeAdd = "add"
const DispositionTypeReplace = "replace"
const StatusSuccess = "success"
const StatusFailed = "failed"

type Report struct {
	ReportSchemaVersion string          `json:"report_schema_version"`
	UploadID            string          `json:"upload_id"`
	SenderID            string          `json:"sender_id"`
	DataStreamID        string          `json:"data_stream_id"`
	DataStreamRoute     string          `json:"data_stream_route"`
	Jurisdiction        string          `json:"jurisdiction"`
	DexIngestDatetime   string          `json:"dex_ingest_datetime"`
	ContentType         string          `json:"content_type"`
	DispositionType     string          `json:"disposition_type"`
	StageInfo           ReportStageInfo `json:"stage_info"`
	Content             any             `json:"content"` // TODO: Can we limit this to a specific type (i.e. ReportContent or UploadStatusTYpe type?
}

func (r *Report) Type() string {
	return "Report"
}

func (r *Report) OrigMessage() *azservicebus.ReceivedMessage {
	return nil
}

func (r *Report) SetIdentifier(id string) {
	r.UploadID = id
}

func (r *Report) SetType(t string) {
	r.StageInfo.Stage = t
}

func (r *Report) SetOrigMessage(_ *azservicebus.ReceivedMessage) {
	// no-op
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

type BulkMetadataTransformReportContent struct {
	ReportContent
	Transforms []MetadataTransformContent `json:"transforms"`
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

type MetadataTransformContent struct {
	Action string `json:"action"` // append, update, remove
	Field  string `json:"field"`  // Name of the field the action was performed on
	Value  string `json:"value"`  // Optional; Value given to the appended or updated field.
}

type FileCopyContent struct {
	ReportContent
	FileSourceBlobUrl      string `json:"file_source_blob_url"`
	FileDestinationBlobUrl string `json:"file_destination_blob_url"`
	Timestamp              string `json:"timestamp"`
}

type UploadStatusContent struct {
	ReportContent
	Filename string `json:"filename"`
	Tguid    string `json:"tguid"`
	Offset   string `json:"offset"`
	Size     string `json:"size"`
}

type Builder[T any] interface {
	SetStage(string) Builder[T]
	SetUploadId(string) Builder[T]
	SetManifest(map[string]string) Builder[T]
	AppendIssue(string) Builder[T]
	SetStatus(string) Builder[T]
	SetStartTime(time.Time) Builder[T]
	SetEndTime(time.Time) Builder[T]
	SetDispositionType(string) Builder[T]
	SetContent(T) Builder[T]
	Build() *Report
}

func NewBuilder[T any](version string, stage string, uploadId string, dispType string) Builder[T] {
	return &ReportBuilder[T]{
		Version:         version,
		Stage:           stage,
		UploadId:        uploadId,
		DispositionType: dispType,
		Status:          StatusSuccess,
		StartTime:       time.Now().UTC(),
		EndTime:         time.Now().UTC(),
	}
}

func NewBuilderWithManifest[T any](version string, stage string, uploadId string, manifest map[string]string, dispType string) Builder[T] {
	return &ReportBuilder[T]{
		Version:         version,
		Stage:           stage,
		UploadId:        uploadId,
		Manifest:        manifest,
		DispositionType: dispType,
		Status:          StatusSuccess,
		StartTime:       time.Now().UTC(),
		EndTime:         time.Now().UTC(),
	}
}

type ReportBuilder[T any] struct {
	Stage           string
	Version         string
	UploadId        string
	Manifest        map[string]string
	Issues          []string
	Status          string
	StartTime       time.Time
	EndTime         time.Time
	DispositionType string
	Content         T
}

func (b *ReportBuilder[T]) SetStage(s string) Builder[T] {
	b.Stage = s
	return b
}

func (b *ReportBuilder[T]) SetUploadId(id string) Builder[T] {
	b.UploadId = id
	return b
}

func (b *ReportBuilder[T]) SetManifest(m map[string]string) Builder[T] {
	b.Manifest = m
	return b
}

func (b *ReportBuilder[T]) SetStatus(s string) Builder[T] {
	b.Status = s
	return b
}

func (b *ReportBuilder[T]) AppendIssue(i string) Builder[T] {
	b.Issues = append(b.Issues, i)
	return b
}

func (b *ReportBuilder[T]) SetStartTime(t time.Time) Builder[T] {
	b.StartTime = t
	return b
}

func (b *ReportBuilder[T]) SetEndTime(t time.Time) Builder[T] {
	b.EndTime = t
	return b
}

func (b *ReportBuilder[T]) SetDispositionType(d string) Builder[T] {
	b.DispositionType = d
	return b
}

func (b *ReportBuilder[T]) SetContent(c T) Builder[T] {
	b.Content = c
	return b
}

func (b *ReportBuilder[T]) Build() *Report {
	switch b.Version {
	default:
		return &Report{
			ReportSchemaVersion: b.Version,
			UploadID:            b.UploadId,
			SenderID:            metadata.GetSenderId(b.Manifest),
			DataStreamID:        metadata.GetDataStreamID(b.Manifest),
			DataStreamRoute:     metadata.GetDataStreamRoute(b.Manifest),
			Jurisdiction:        metadata.GetJurisdiction(b.Manifest),
			DexIngestDatetime:   metadata.GetDexIngestDatetime(b.Manifest),
			ContentType:         "application/json",
			DispositionType:     b.DispositionType,
			StageInfo: ReportStageInfo{
				Issues:           b.Issues,
				Stage:            b.Stage,
				Service:          "UPLOAD API",
				Version:          fmt.Sprintf("%s_%s", version.LatestReleaseVersion, version.GitShortSha),
				Status:           b.Status,
				StartProcessTime: b.StartTime.Format(time.RFC3339),
				EndProcessTime:   b.EndTime.Format(time.RFC3339),
			},
			Content: b.Content,
		}
	}
}
