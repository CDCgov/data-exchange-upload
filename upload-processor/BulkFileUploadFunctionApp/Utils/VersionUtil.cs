using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Utils
{
    public static class VersionUtil
    {
       public static MetadataVersion FromString(string version)
       {
          switch(version)
          {
            case "1.0":
                return MetadataVersion.V1;
            case "2.0":
                return MetadataVersion.V2;
            default:
                throw new ArgumentException($"{version} is not supported version.");
          }
       }
    }

}