package models

type Report struct {
	ReportSchemaVersion string `json:"report_schema_version"`
	UploadID            string `json:"upload_id"`
	DataStreamID        string `json:"data_stream_id"`
	DataStreamRoute     string `json:"data_stream_route"`
	Jurisdiction        string `json:"jurisdiction"`
	DexIngestDatetime   string `json:"dex_ingest_datetime"`
	ContentType         string `json:"content_type"`
	DispositionType     string `json:"disposition_type"`
	Content             any    `json:"content"` // TODO: Can we limit this to a specific type (i.e. ReportContent or UploadStatusTYpe type?
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
	Filename  string `json:"filename"`
	Metadata  any    `json:"metadata"`
	Timestamp string `json:"timestamp"`
}

type MetaDataTransformContent struct {
	ReportContent
	Action string `json:"action"` // append, update, remove
	Field  string `json:"field"`  // Name of the field the action was performed on
	Value  string `json:"value"`  // Optional; Value given to the appended or updated field.
}

type BulkMetaDataTransformContent struct {
	ReportContent
	Transforms []MetaDataTransformContent `json:"transforms"`
}

type UploadStatusContent struct {
	ReportContent
	Filename string `json:"filename"`
	// Additional postReceive values:
	Tguid  string `json:"tguid"`
	Offset string `json:"offset"`
	Size   string `json:"size"`
}

type FileCopyContent struct {
	ReportContent
	FileSourceBlobUrl      string `json:"file_source_blob_url"`
	FileDestinationBlobUrl string `json:"file_destination_blob_url"`
	Timestamp              string `json:"timestamp"`
}

func (r *Report) Identifier() string {
	return r.UploadID
}
