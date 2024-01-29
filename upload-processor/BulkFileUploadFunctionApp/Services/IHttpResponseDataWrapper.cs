using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IHttpResponseDataWrapper
    {
        Task WriteStringAsync(string responseContent);
        HttpStatusCode StatusCode { get; set; }
    }
}