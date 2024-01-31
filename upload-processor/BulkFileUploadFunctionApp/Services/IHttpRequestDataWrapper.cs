
namespace BulkFileUploadFunctionApp.Services
{
    public interface IHttpRequestDataWrapper
    {
        IHttpResponseDataWrapper CreateResponse();
    }
}