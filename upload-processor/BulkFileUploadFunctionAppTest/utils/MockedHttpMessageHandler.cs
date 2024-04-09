using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionAppTest.utils
{
    public class MockedHttpMessageHandler : HttpMessageHandler
    {
        private readonly Func<Task<HttpResponseMessage>> _responseFactory;
        private readonly HttpResponseMessage? _httpResponseMessage;

        public MockedHttpMessageHandler(HttpResponseMessage httpResponseMessage)
        {
            _responseFactory = () => Task.FromResult(httpResponseMessage);
            _httpResponseMessage = httpResponseMessage;
        }

        public MockedHttpMessageHandler(Func<Task<HttpResponseMessage>>? responseFactory)
        {
            _responseFactory = responseFactory ?? throw new ArgumentNullException(nameof(responseFactory));
        }

        protected override async Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            cancellationToken.ThrowIfCancellationRequested();
            if (_responseFactory != null)
            {
                return await _responseFactory.Invoke();
            }

            return _httpResponseMessage ?? await Task.FromResult(new HttpResponseMessage(HttpStatusCode.BadRequest));
        }
    }

}
