
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Model
{
    public class TusStorage
    {
        public string? Container { get; set; }

        public string? Key { get; set; }

        public string? Type { get; set; }
    }

    /// <summary>
    /// Sample data for a tus .info file
    /// {
    ///    "ID":"a0e127caec153d6047ee966bc2acd8cb",
    ///    "Size":7952,
    ///    "SizeIsDeferred":false,
    ///    "Offset":0,
    ///    "MetaData":{
    ///       "filename":"flower.jpeg",
    ///       "meta_field":"meta_value"
    ///    },
    ///    "IsPartial":false,
    ///    "IsFinal":false,
    ///    "PartialUploads":null,
    ///    "Storage":{
    ///       "Container":"bulkuploads",
    ///       "Key":"tus-prefix/a0e127caec153d6047ee966bc2acd8cb",
    ///       "Type":"azurestore"
    ///    }
    /// }
    /// </summary>
    public class TusInfoFile
    {
        public string? ID { get; init; }

        public long Size { get; init; }

        public bool SizeIsDeferred { get; init; }

        public long Offset { get; init; }

        public bool IsPartial { get; init; }

        public bool IsFinal { get; init; }

        public TusStorage? Storage { get; init; }
        public Dictionary<string, string>? MetaData { get; set; }

        public MetadataVersion GetMetadataVersion()
        {
            // Default to v1.
            MetadataVersion version = MetadataVersion.V1;

            if (MetaData == null)
            {
                return version;
            }

            string verStr = MetaData.GetValueOrDefault("version", Constants.METADATA_VERSION_ONE);

            switch (verStr)
            {
                case Constants.METADATA_VERSION_ONE:
                    version = MetadataVersion.V1;
                    break;
                case Constants.METADATA_VERSION_TWO:
                    version = MetadataVersion.V2;
                    break;
            }

            return version;
        }
        
        public string GetUseCase()
        {
            if (MetaData == null)
            {
                throw new TusInfoFileException("Cannot get use case when metadata is null.");
            }

            MetadataVersion versionNum = GetMetadataVersion();

            string? useCase = versionNum == MetadataVersion.V1 
                ? MetaData.GetValueOrDefault(Constants.USE_CASE_FIELDNAME_V1, null) 
                : versionNum == MetadataVersion.V2
                ? MetaData.GetValueOrDefault(Constants.USE_CASE_FIELDNAME_V2, null) : null;
            

            if (useCase == null)
            {
                throw new TusInfoFileException($"No use case provided in metadata: {MetaData}");
            }

            return useCase;
        }

        public string GetUseCaseCategory()
        {
            if (MetaData == null)
            {
                throw new TusInfoFileException("Cannot get use case when metadata is null.");
            }

            MetadataVersion versionNum = GetMetadataVersion();

            string? useCaseCategory = versionNum == MetadataVersion.V1
                ? MetaData.GetValueOrDefault(Constants.USE_CASE_CATEGORY_FIELDNAME_V1, null)
                : versionNum == MetadataVersion.V2
                ? MetaData.GetValueOrDefault(Constants.USE_CASE_CATEGORY_FIELDNAME_V2, null) : null;

            if (useCaseCategory == null)
            {
                throw new TusInfoFileException($"No use case category provided in metadata: {MetaData}");
            }

            return useCaseCategory;
        }
    }
}
