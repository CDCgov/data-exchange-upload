using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp
{
    public interface IHttpRequestDataWrapper
    {
        IHttpResponseDataWrapper CreateResponse(); 
    }
}