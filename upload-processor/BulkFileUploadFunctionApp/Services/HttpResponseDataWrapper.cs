using System.Net;
using Microsoft.Azure.Functions.Worker.Http;

namespace BulkFileUploadFunctionApp.Services
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

    }

}