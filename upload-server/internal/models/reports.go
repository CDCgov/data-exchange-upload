package models

type Report struct {
	UploadID        string `json:"upload_id"`
	StageName       string `json:"stage_name"`
	DataStreamID    string `json:"data_stream_id"`
	DataStreamRoute string `json:"data_stream_route"`
	ContentType     string `json:"content_type"`
	DispositionType string `json:"disposition_type"`
	Content         any    `json:"content"` // TODO: Can we limit this to a specific type (i.e. ReportContent or UploadStatusTYpe type?
}

type ReportContent struct {
	SchemaVersion string `json:"schema_version"`
	SchemaName    string `json:"schema_name"`
}

type UploadLifecycleContent struct {
	ReportContent
	Status string `json:"status"`
}

type MetaDataVerifyContent struct {
	ReportContent
	Filename string `json:"filename"`
	Metadata any    `json:"metadata"`
	Issues   error  `json:"issues"`
}

type MetaDataTransformContent struct {
	ReportContent
	Action string `json:"action"` // append, update, remove
	Field  string `json:"field"`  // Name of the field the action was performed on
	Value  string `json:"value"`  // Optional; Value given to the appended or updated field.
}

type UploadStatusContent struct {
	ReportContent
	Filename string `json:"filename"`
	Metadata any    `json:"metadata"`
	// Additional postReceive values:
	Tguid  string `json:"tguid"`
	Offset string `json:"offset"`
	Size   string `json:"size"`
}

func (r *Report) Identifier() string {
	return r.UploadID
}
