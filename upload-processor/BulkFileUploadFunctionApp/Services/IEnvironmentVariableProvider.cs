using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp
{
    public interface IEnvironmentVariableProvider
    {
        string GetEnvironmentVariable(string name);
    }
}