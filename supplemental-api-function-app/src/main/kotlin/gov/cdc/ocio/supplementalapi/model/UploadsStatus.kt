package gov.cdc.ocio.supplementalapi.model

class UploadsStatus
{
    var summary = PageSummary()

    var items = mutableListOf<UploadStatus>()
}