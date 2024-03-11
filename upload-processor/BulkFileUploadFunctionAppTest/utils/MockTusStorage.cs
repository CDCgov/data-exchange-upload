using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionAppTest.utils
{
    [TestClass]
    internal class MockTusStorage
    {
        public string? Container { get; set; }

        public string? Key { get; set; }

        public string? Type { get; set; }
    }
}
