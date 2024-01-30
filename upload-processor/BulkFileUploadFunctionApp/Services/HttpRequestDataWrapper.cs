
using Microsoft.Azure.Functions.Worker.Http;

namespace BulkFileUploadFunctionApp.Services
{
    public class HttpRequestDataWrapper : IHttpRequestDataWrapper
{
    private readonly HttpRequestData _request;

    public HttpRequestDataWrapper(HttpRequestData request)
    {
        _request = request;
    }

    public IHttpResponseDataWrapper CreateResponse()
    {
        var response = _request.CreateResponse();
        return new HttpResponseDataWrapper(response);
    }

}
}