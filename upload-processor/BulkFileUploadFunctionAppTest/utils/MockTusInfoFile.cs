using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionAppTest.utils
{
    [TestClass]
    internal class MockTusInfoFile
    {

        public string? ID { get; set; }

        public long Size { get; set; }

        public bool SizeIsDeferred { get; set; }

        public long Offset { get; set; }

        public bool IsPartial { get; set; }

        public bool IsFinal { get; set; }

        public Dictionary<string, string>? MetaData { get; set; }

        public MockTusStorage? Storage { get; set; }


        public MockTusInfoFile GetObjectFromBlobJsonContent<TusInfoFile>(string connectionString, string sourceContainerName, string blobPathname)
        {
            return new MockTusInfoFile
            {
                ID = "a0e127caec153d6047ee966bc2acd8cb",
                Size = 7952,
                SizeIsDeferred = false,
                Offset = 0,
                MetaData = new Dictionary<string, string>{
                        {"meta_destination_id", "flower.jpeg"},
                        {"meta_ext_event","meta_value"}
                    },
                IsPartial = false,
                IsFinal = false,
                Storage = new MockTusStorage
                {
                    Container = "bulkuploads",
                    Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb",
                    Type = "azurestore"
                }
            };
        }
    }

}
