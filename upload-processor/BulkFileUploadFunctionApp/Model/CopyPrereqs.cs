namespace BulkFileUploadFunctionApp.Model
{
    public class CopyPrereqs
    {
        public string? UploadId { get; init; }
        public string? SourceBlobUrl { get; init; }
        public string? TusPayloadFilename { get; init; }
        public string? UseCase { get; init; }
        public string? UseCaseCategory { get; init; }
        public string? DexBlobFileName { get; init; }
        public string? DexBlobFolderName { get; init; }
        public List<CopyTargetsEnum>? Targets { get; init; }
        public Trace? Trace { get; init; }
        public string? DexBlobUrl { get; set; }
        public Dictionary<string, string>? Metadata { get; set; }
        

        public CopyPrereqs() { }

        public CopyPrereqs(string sourceBlobUrl)
        {
            SourceBlobUrl = sourceBlobUrl;
        }

        // create a Default constructor
        public CopyPrereqs(string uploadId, string sourceBlobUrl, string tusPayloadFilename, string useCase, string useCaseCategory, string dexBlobFileName, Dictionary<string, string> metadata, List<CopyTargetsEnum> targets, Trace trace)
        {
            UploadId = uploadId;
            SourceBlobUrl = sourceBlobUrl;
            TusPayloadFilename = tusPayloadFilename;
            UseCase = useCase;
            UseCaseCategory = useCaseCategory;
            DexBlobFileName = dexBlobFileName;
            DexBlobFolderName = $"{useCase}-{useCaseCategory}".ToLower();
            Metadata = metadata;
            Targets = targets;
            Trace = trace;
        }
    }
}