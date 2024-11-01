package reports

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/version"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
)

const StageMetadataVerify = "metadata-verify"
const StageMetadataTransform = "metadata-transform"
const StageFileCopy = "blob-file-copy"
const StageUploadStatus = "upload-status"
const StageUploadStarted = "upload-started"
const StageUploadCompleted = "upload-completed"
const DispositionTypeAdd = "add"
const DispositionTypeReplace = "replace"
const StatusSuccess = "SUCCESS"
const StatusFailed = "FAILURE"
const IssueLevelWarning = "WARNING"
const IssueLevelError = "ERROR"

type Report struct {
	Event               event.Event     `json:"-"`
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

func (r *Report) RetryCount() int {
	return r.Event.RetryCount
}

func (r *Report) IncrementRetryCount() {
	r.Event.RetryCount++
}

func (r *Report) Type() string {
	return r.Event.Type
}

func (r *Report) OrigMessage() *azservicebus.ReceivedMessage {
	return nil
}

func (r *Report) SetIdentifier(id string) {
	r.Event.ID = id
}

func (r *Report) SetType(t string) {
	r.Event.Type = t
}

func (r *Report) SetOrigMessage(_ *azservicebus.ReceivedMessage) {
	// no-op
}

type ReportStageInfo struct {
	Service          string        `json:"service"`
	Action           string        `json:"action"`
	Version          string        `json:"version"`
	Status           string        `json:"status"`
	Issues           []ReportIssue `json:"issues"`
	StartProcessTime string        `json:"start_processing_time"`
	EndProcessTime   string        `json:"end_processing_time"`
}

type ReportIssue struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (r *ReportIssue) String() string {
	return fmt.Sprintf("%s: %s", r.Level, r.Message)
}

func (r *Report) Identifier() string {
	return r.UploadID
}

type ReportContent struct {
	ContentSchemaName    string `json:"content_schema_name"`
	ContentSchemaVersion string `json:"content_schema_version"`
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
	DestinationName        string `json:"destination_name"`
}

type UploadStatusContent struct {
	ReportContent
	Filename string `json:"filename"`
	Tguid    string `json:"tguid"`
	Offset   int64  `json:"offset"`
	Size     int64  `json:"size"`
}

type Builder[T any] interface {
	SetAction(string) Builder[T]
	SetUploadId(string) Builder[T]
	SetManifest(map[string]string) Builder[T]
	AppendIssue(ReportIssue) Builder[T]
	SetStatus(string) Builder[T]
	SetStartTime(time.Time) Builder[T]
	SetEndTime(time.Time) Builder[T]
	SetDispositionType(string) Builder[T]
	SetContent(T) Builder[T]
	Build() *Report
}

func NewBuilder[T any](version string, action string, uploadId string, dispType string) Builder[T] {
	return &ReportBuilder[T]{
		Version:         version,
		Action:          action,
		UploadId:        uploadId,
		DispositionType: dispType,
		Status:          StatusSuccess,
		StartTime:       time.Now().UTC(),
		EndTime:         time.Now().UTC(),
	}
}

func NewBuilderWithManifest[T any](version string, action string, uploadId string, manifest map[string]string, dispType string) Builder[T] {
	return &ReportBuilder[T]{
		Version:         version,
		Action:          action,
		UploadId:        uploadId,
		Manifest:        manifest,
		DispositionType: dispType,
		Status:          StatusSuccess,
		StartTime:       time.Now().UTC(),
		EndTime:         time.Now().UTC(),
	}
}

type ReportBuilder[T any] struct {
	Action          string
	Version         string
	UploadId        string
	Manifest        map[string]string
	Issues          []ReportIssue
	Status          string
	StartTime       time.Time
	EndTime         time.Time
	DispositionType string
	Content         T
}

func (b *ReportBuilder[T]) SetAction(s string) Builder[T] {
	b.Action = s
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

func (b *ReportBuilder[T]) AppendIssue(i ReportIssue) Builder[T] {
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
			Event: event.Event{
				Type: b.Action,
				ID:   b.UploadId,
			},
			ReportSchemaVersion: b.Version,
			UploadID:            b.UploadId,
			SenderID:            metadata.GetSenderId(b.Manifest),
			DataStreamID:        b.Manifest["data_stream_id"],
			DataStreamRoute:     b.Manifest["data_stream_route"],
			Jurisdiction:        metadata.GetJurisdiction(b.Manifest),
			DexIngestDatetime:   metadata.GetDexIngestDatetime(b.Manifest),
			ContentType:         "application/json",
			DispositionType:     b.DispositionType,
			StageInfo: ReportStageInfo{
				Issues:           b.Issues,
				Action:           b.Action,
				Service:          "UPLOAD API",
				Version:          fmt.Sprintf("%s_%s", version.LatestReleaseVersion, version.GitShortSha),
				Status:           b.Status,
				StartProcessTime: b.StartTime.Format(time.RFC3339Nano),
				EndProcessTime:   b.EndTime.Format(time.RFC3339Nano),
			},
			Content: b.Content,
		}
	}
}
