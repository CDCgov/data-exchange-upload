using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Exceptions
{
    public class ProcStatClientException : Exception
    {
        public ProcStatClientException() { }

        public ProcStatClientException(string message) : base(message) { }

        public ProcStatClientException(string message, Exception inner) : base(message, inner) { }
    }
}
