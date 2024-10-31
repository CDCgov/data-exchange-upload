import { expect } from '@playwright/test';
import {
  HttpRequest,
  HttpResponse,
  RawHttpRequest,
  RawHttpResponse,
  ResponseContext as ResponseContextInterface,
  UploadOptions,
  UploadResponse,
  UploadStatusType
} from './index.d';
import { MemoryStorage } from './memory-storage';
import { UploadClient } from './upload-client';

export { ResponseContextInterface as ResponseContext };
export type ContextOptions = Pick<UploadOptions, 'headers' | 'retryDelays'>;

function newContext(baseUrl: string, options: ContextOptions = {}): ClientContext {
  return new ClientContext(baseUrl, options);
}

export default {
  newContext
};

class ClientContext {
  private baseURL: string;
  private storage: MemoryStorage;
  private options: ContextOptions;

  constructor(baseURL: string, options: ContextOptions = {}) {
    this.baseURL = baseURL;
    this.storage = new MemoryStorage();
    this.options = options;
  }

  setAuthenticationToken(authToken: string): void {
    this.options.headers = {
      ...this.options.headers,
      Authorization: `Bearer ${authToken}`
    };
  }

  setRetryDelays(retryDelays: number[]) {
    this.options.retryDelays = retryDelays;
  }

  async upload(filename: string, metadata: { [key: string]: string }): Promise<ResponseContext> {
    return new Promise(resolve => {
      const uploader = new UploadClient(filename, {
        ...this.options,
        metadata,
        endpoint: this.baseURL,
        urlStorage: this.storage,
        onComplete: (response: UploadResponse) => {
          resolve(new ResponseContext(response));
        }
      });
      uploader.upload();
    });
  }

  async uploadInitiated(
    filename: string,
    metadata: { [key: string]: string }
  ): Promise<ResponseContext> {
    return new Promise(resolve => {
      const uploader = new UploadClient(filename, {
        ...this.options,
        metadata,
        endpoint: this.baseURL,
        urlStorage: this.storage,
        onInitiated: (response: UploadResponse) => {
          resolve(new ResponseContext(response));
        },
        onComplete: (_: UploadResponse) => {}
      });
      return uploader.upload();
    });
  }

  async uploadInProgress(
    filename: string,
    metadata: { [key: string]: string }
  ): Promise<ResponseContext> {
    return new Promise(resolve => {
      const uploader = new UploadClient(filename, {
        ...this.options,
        metadata,
        endpoint: this.baseURL,
        urlStorage: this.storage,
        onInProgress: (response: UploadResponse) => {
          resolve(new ResponseContext(response));
        },
        onComplete: (_: UploadResponse) => {}
      });
      return uploader.upload();
    });
  }

  async removeUpload(uploadUrl: string): Promise<void> {
    return UploadClient.terminate(uploadUrl);
  }
}

class ResponseContext implements ResponseContextInterface {
  private response: UploadResponse;
  private lastRequest: HttpRequest | null;
  private lastRawRequest: RawHttpRequest | null;
  private lastResponse: HttpResponse | null;
  private lastRawResponse: RawHttpResponse | null;

  constructor(response: UploadResponse) {
    this.response = response;

    this.lastRequest = response.httpRequests?.slice(-1)?.pop() ?? null;
    this.lastRawRequest = (this.lastRequest?.getUnderlyingObject() as RawHttpRequest) ?? null;

    this.lastResponse = response.httpResponses?.slice(-1)?.pop() ?? null;
    this.lastRawResponse = (this.lastResponse?.getUnderlyingObject() as RawHttpResponse) ?? null;
  }

  getLastRequest(): HttpRequest | null {
    return this.lastRequest;
  }

  getLastRawRequest(): RawHttpRequest | null {
    return this.lastRawRequest;
  }

  getLastResponse(): HttpResponse | null {
    return this.lastResponse;
  }

  getLastRawResponse(): RawHttpResponse | null {
    return this.lastRawResponse;
  }

  getResponseStatusCode(): number | null {
    return this.getLastResponse()?.getStatus() ?? null;
  }

  getResponseBodyString(): string | null {
    return this.getLastResponse()?.getBody() ?? null;
  }

  getResponseBodyJson(): { [key: string]: string } | null {
    const bodyString = this.getResponseBodyString();
    if (bodyString) {
      return JSON.parse(bodyString);
    }
    return null;
  }

  getUploadUrl(): string | null {
    return this.response.uploadUrl ?? null;
  }

  getUploadUrlId(): string | null {
    return this.response.uploadUrlId ?? null;
  }

  getUploadId(): string | null {
    return this.response.uploadId ?? null;
  }

  getUploadStatus(): string | null {
    return this.response.uploadStatus ?? null;
  }

  assertUploadStatus(expectedStatus: UploadStatusType): void {
    expect(this.response.uploadStatus).toEqual(expectedStatus);
  }

  assertResponseStatusCode(expectedStatusCode: number) {
    expect(this.getResponseStatusCode()).toEqual(expectedStatusCode);
  }

  assertNotResponseStatusCode(expectedStatusCodeNot: number): void {
    expect(this.getResponseStatusCode()).not.toEqual(expectedStatusCodeNot);
  }

  assertResponseBody(expectedBodySubstring: string): void {
    expect(this.getResponseBodyString()).toContain(expectedBodySubstring);
  }

  assertResponse(expectedStatusCode: number, expectedBodySubstring: string): void {
    expect(this.getResponseStatusCode()).toEqual(expectedStatusCode);
    expect(this.getResponseBodyString()).toContain(expectedBodySubstring);
  }

  assertSuccess() {
    this.assertUploadStatus('Complete');
    this.assertResponseStatusCode(204);
  }

  assertError(expectedStatusCode: number) {
    this.assertUploadStatus('Failed');
    this.assertResponseStatusCode(expectedStatusCode);
  }
}
