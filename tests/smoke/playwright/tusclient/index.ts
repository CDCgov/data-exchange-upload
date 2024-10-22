import { expect } from '@playwright/test';
import * as http from 'http';
import { MemoryStorage } from './memory-storage';
import * as client from './upload-client';

type ClientOptions = Pick<client.UploadOptions, 'headers' | 'retryDelays' | 'chunkSize'>;

type UploadResponse = client.UploadResponse;
type UploadStatusType = client.UploadStatusType;
type HttpRequest = client.HttpRequest;
type HttpResponse = client.HttpResponse;
type RawHttpRequest = http.ClientRequest;
type RawHttpResponse = http.IncomingMessage;

function newContext(baseUrl: string): ClientContext {
  return new ClientContext(baseUrl);
}

export default {
  newContext
};

class ClientContext {
  private baseURL: string;
  private storage: MemoryStorage;
  private options: ClientOptions;

  constructor(baseURL: string, options: ClientOptions = {}) {
    this.baseURL = baseURL;
    this.storage = new MemoryStorage();
    this.options = options;
  }

  setAuthenticationHeaders(authHeaders: { [key: string]: string }): void {
    this.options.headers = {
      ...authHeaders
    };
  }

  setRetryDelays(retryDelays: number[]) {
    this.options.retryDelays = retryDelays;
  }

  setChunkSize(chunkSize: number) {
    this.options.chunkSize = chunkSize;
  }

  async upload(filename: string, metadata: { [key: string]: string }): Promise<UploadContext> {
    const response = await client.uploadFile(filename, {
      ...this.options,
      metadata,
      endpoint: this.baseURL,
      urlStorage: this.storage
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

  getUploadId(): string | null {
    return this.response.uploadId ?? null;
  }

  getUploadStatus(): string | null {
    return this.response.uploadStatus ?? null;
  }

  private evaluateNotNull(obj: any): any | null {
    expect(obj).not.toBeNull();
    return obj;
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

  assertValidationErrors(expectedError: string): void {
    expect(this.getResponseBodyJson()?.validation_errors).not.toBeNull();
    expect(this.getResponseBodyJson()?.validation_errors).toContain(expectedError);
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
