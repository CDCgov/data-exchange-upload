using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Threading.Tasks;
using tusdotnet;
using tusdotnet.Interfaces;
using tusdotnet.Models;
using tusdotnet.Models.Configuration;
using tusdotnet.Stores;

namespace TUSTestHelpers.Behaviors.Server
{
    public class DefaultServer
    {

        public static Task ProcessPostUpload(ITusFile file, FileCompleteContext eventContext) {
            //react to a completed upload here...
            //await DoSomeProcessing(content, metadata);
            return Task.CompletedTask;
        }

    }
}
