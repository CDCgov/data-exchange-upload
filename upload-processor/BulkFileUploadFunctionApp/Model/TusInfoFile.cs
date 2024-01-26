
namespace BulkFileUploadFunctionApp.Model
{
    internal class TusStorage
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
    internal class TusInfoFile
    {
        public string? ID { get; set; }

        public long Size { get; set; }

        public bool SizeIsDeferred { get; set; }

        public long Offset { get; set; }

        public bool IsPartial { get; set; }

        public bool IsFinal { get; set; }

        public Dictionary<string, string>? MetaData { get; set; }

        public TusStorage? Storage { get; set; }
    }
}
