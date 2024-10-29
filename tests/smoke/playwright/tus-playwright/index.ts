import { expect } from '@playwright/test';
import * as http from 'http';
import {
  UploadOptions as ClientOptions,
  HttpRequest,
  HttpResponse,
  UploadResponse,
  UploadStatusType
} from './index.d';
import { MemoryStorage } from './memory-storage';
import * as client from './upload-client';

export type ContextOptions = Pick<ClientOptions, 'headers' | 'retryDelays' | 'chunkSize'>;
export type UploadHooks = Pick<
  ClientOptions,
  | 'onBeforeRequest'
  | 'onAfterResponse'
  | 'onProgress'
  | 'onChunkComplete'
  | 'onUploadStarted'
  | 'onUploadCreated'
  | 'onUploadPaused'
  | 'onUploadResumed'
>;

type RawHttpRequest = http.ClientRequest;
type RawHttpResponse = http.IncomingMessage;

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

  setChunkSize(chunkSize: number) {
    this.options.chunkSize = chunkSize;
  }

  async upload(
    filename: string,
    metadata: { [key: string]: string },
    hooks: UploadHooks = {}
  ): Promise<UploadContext> {
    const response = await client.uploadFile(filename, {
      ...this.options,
      metadata,
      endpoint: this.baseURL,
      urlStorage: this.storage,
      ...hooks
    });
    return new UploadContext(response);
  }
}

class UploadContext {
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
