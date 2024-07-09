package reports

import (
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
	SetContentBuilder(ContentBuilder[T]) Builder[T]
	Build() *Report
}

func NewBuilder[T any](version string, stage string, uploadId string, manifest map[string]string, dispType string, contentBuilder ContentBuilder[T]) Builder[T] {
	return &ReportBuilder[T]{
		Version:         version,
		Stage:           stage,
		UploadId:        uploadId,
		Manifest:        manifest,
		DispositionType: dispType,
		ContentBuilder:  contentBuilder,
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
	ContentBuilder  ContentBuilder[T]
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

func (b *ReportBuilder[T]) SetContentBuilder(cb ContentBuilder[T]) Builder[T] {
	b.ContentBuilder = cb
	return b
}

func (b *ReportBuilder[T]) Build() *Report {
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

func NewReportContentBuilder[T any]() *ReportContentBuilder[T] {
	return &ReportContentBuilder[T]{}
}

type ContentBuilder[T any] interface {
	SetContent(T) ContentBuilder[T]
	Build() T
}

type ReportContentBuilder[T any] struct {
	Content T
}

func (b *ReportContentBuilder[T]) SetContent(c T) ContentBuilder[T] {
	b.Content = c
	return b
}

func (b *ReportContentBuilder[T]) Build() T {
	return b.Content
}
