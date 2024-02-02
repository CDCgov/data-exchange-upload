using System.Net;
using System.Text;

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

    public Uri Url => _request.Url;

    public Stream Body => _request.Body;

    public HttpHeadersCollection Headers => _request.Headers;

    public IHttpResponseDataWrapper CreateResponse()
    {
        if (_request != null)
        {
            var response = _request.CreateResponse();
            return new HttpResponseDataWrapper(response);
        }
        else
        {
            return CreateDefaultResponse();
        }
    }

    private IHttpResponseDataWrapper CreateDefaultResponse()
    {
        // Implement logic to create a default IHttpResponseDataWrapper.
        // This might be a mock or dummy implementation suitable for your application.
        return new DefaultHttpResponseDataWrapper(); // Example placeholder.
    }

}

public class DefaultHttpResponseDataWrapper : IHttpResponseDataWrapper
{
    private HttpStatusCode _statusCode;
    private readonly StringBuilder _contentBuilder;

    public DefaultHttpResponseDataWrapper()
    {
        _contentBuilder = new StringBuilder();
        _statusCode = HttpStatusCode.OK; // Default status code
    }

    public async Task WriteStringAsync(string responseContent)
    {
        // Simulate writing to the response
        await Task.Run(() => _contentBuilder.Append(responseContent));
    }

    public HttpStatusCode StatusCode
    {
        get => _statusCode;
        set => _statusCode = value;
    }
       
    }

}