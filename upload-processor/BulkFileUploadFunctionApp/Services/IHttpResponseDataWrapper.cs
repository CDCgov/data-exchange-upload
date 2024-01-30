
using System.Net;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IHttpResponseDataWrapper
    {
        Task WriteStringAsync(string responseContent);
        HttpStatusCode StatusCode { get; set; }
    }
}