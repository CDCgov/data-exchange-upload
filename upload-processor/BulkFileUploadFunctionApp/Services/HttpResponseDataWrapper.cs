using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading.Tasks;
using Microsoft.Azure.Functions.Worker.Http;

namespace BulkFileUploadFunctionApp
{
    public class HttpResponseDataWrapper : IHttpResponseDataWrapper
{
    private readonly HttpResponseData _response;

    public HttpResponseDataWrapper(HttpResponseData response)
    {
        _response = response;
    }

    public async Task WriteStringAsync(string responseContent)
    {
        await _response.WriteStringAsync(responseContent);
    }

    public HttpStatusCode StatusCode
    {
        get => (HttpStatusCode)_response.StatusCode;
        set => _response.StatusCode = (HttpStatusCode)value;
    }
        HttpStatusCode IHttpResponseDataWrapper.StatusCode { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }
    }

}