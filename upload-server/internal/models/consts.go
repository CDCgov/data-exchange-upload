package models

const (
	STATUS_UP         = "UP"
	STATUS_DEGRADED   = "DEGRADED"
	STATUS_DOWN       = "DOWN"
	HEALTH_ISSUE_NONE = "None reported"
	//
	AZ_TEST_CONTAINER_NAME = "testcontainernameempty"
	AZ_BLOB_CLIENT_NA      = "error: client not available, check config"
	//
	PROCESSING_STATUS_APP = "Processing Status"
	SERVICE_BUS           = "Azure Service Bus"
	REDIS_LOCKER          = "Redis Locker"

	META_DESTINATION_ID = "meta_destination_id"
	META_EXT_EVENT      = "meta_ext_event"
	FILENAME            = "filename"
	DATE_YYYY_MM_DD     = "date_YYYY_MM_DD"
	CLOCK_TICKS         = "clock_ticks"

	DEX_INGEST_DATE_TIME_KEY_NAME = "dex_ingest_datetime"

	TARGET_DEX_ROUTER = "dex_routing"

	EVENT_UPLOAD_ID           = "event.Upload.ID"
	TGUID_KEY                 = "tguid"
	TUS_STORAGE_HEALTH_PREFIX = "Tus storage"
) // .const
