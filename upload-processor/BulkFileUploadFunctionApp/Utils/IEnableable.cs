using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Utils
{
    public interface IEnableable
    {
        string? featureFlagKey { get; set; }
        void DoIfEnabled(string featureFlagKey, Action callback);
    }
}
