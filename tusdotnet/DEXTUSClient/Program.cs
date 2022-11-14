using System;
using System.IO;
using System.Threading.Tasks;
using TusDotNetClient;
using TUSTestHelpers.Behaviors;

namespace DEXTUSClient
{
    class Program
    {
        static void Main(string[] args)
        {
            Console.WriteLine("Starting Upload");
            UploadFile().GetAwaiter().GetResult();
            Console.WriteLine("Finished Upload");
        }

        public static async Task UploadFile()
        {
            var largeFileLocation = @"C:\blah\some-large-file.xyz";
            
            var file = new FileInfo(largeFileLocation);
            var client = new TusClient();

            //set both client and server to run at startup
            var tusEndpoint = "https://localhost:44363/files/";

            var fileUrl = await client.CreateAsync(tusEndpoint, file.Length, ServerExpectedBehaviors.ForceChunkFailure); //http://localhost:1080/
            var uploadOperation = client.UploadAsync(fileUrl, file, chunkSize: 5D);

            uploadOperation.Progressed += (transferred, total) =>
                System.Diagnostics.Debug.WriteLine($"Progress: {transferred}/{total}");

            await uploadOperation;

        }
    }
}
