
using Microsoft.Azure.Functions.Worker.Http;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IHttpRequestDataWrapper
    {
        Uri Url { get; }
        Stream Body { get; }
        HttpHeadersCollection Headers { get; }
        IHttpResponseDataWrapper CreateResponse();
    }
}