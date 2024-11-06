import { expect } from '@playwright/test';
import {
  HttpRequest,
  HttpResponse,
  RawHttpRequest,
  RawHttpResponse,
  UploadContextOptions,
  UploadOptions,
  UploadResponse,
  UploadStatusType
} from './index.d';
import { MemoryStorage } from './memory-storage';
import { UploadClient } from './upload-client';

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

  private getUploadContext(
    filename: string,
    metadata: { [key: string]: string },
    options: UploadContextOptions = {}
  ): UploadContext {
    const { chunkSize, ...contextOptions } = options;

    const uploadOptions: UploadOptions = {
      metadata,
      headers: this.options.headers,
      endpoint: this.baseURL,
      urlStorage: this.storage
    };
    if (options.chunkSize) uploadOptions.chunkSize = options.chunkSize;
    if (this.options.retryDelays) uploadOptions.retryDelays = this.options.retryDelays;

    return new UploadContext(filename, uploadOptions, contextOptions);
  }

  async upload(
    filename: string,
    metadata: { [key: string]: string },
    chunkSize?: number
  ): Promise<ResponseContext> {
    const client = this.getUploadContext(filename, metadata, { chunkSize });
    return client.upload();
  }

  newUploadContext(
    filename: string,
    metadata: { [key: string]: string },
    options: UploadContextOptions = {}
  ): UploadContext {
    return this.getUploadContext(filename, metadata, options);
  }
}

class UploadContext {
  private client: UploadClient;
  private shouldPauseInitialized: boolean;
  private shouldPauseInProgress: boolean;
  private lastResponse: UploadResponse | null;

  constructor(
    filename: string,
    uploadOptions: UploadOptions,
    contextOptions: UploadContextOptions = {}
  ) {
    this.client = new UploadClient(filename, uploadOptions);
    this.lastResponse = null;
    this.shouldPauseInitialized = contextOptions?.shouldPauseInitialized ?? false;
    this.shouldPauseInProgress = contextOptions?.shouldPauseInProgress ?? false;
  }

  async upload(): Promise<ResponseContext> {
    if (this.lastResponse == null && this.shouldPauseInitialized) {
      this.lastResponse = await this.client.uploadInitiated();
    } else if (
      (this.lastResponse == null || this.lastResponse.uploadStatus == 'Initiated') &&
      this.shouldPauseInProgress
    ) {
      this.lastResponse = await this.client.uploadInProgress();
    } else {
      this.lastResponse = await this.client.uploadComplete();
    }

    return new ResponseContext(this.lastResponse);
  }

  isComplete(): boolean {
    return this.lastResponse?.uploadStatus == 'Complete' || false;
  }

  async cleanup(): Promise<void> {
    this.client.terminate();
  }
}

class ResponseContext {
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

  assertInitiatedSuccess() {
    this.assertUploadStatus('Initiated');
    this.assertResponseStatusCode(200);
    expect(this.response.errorMessage).toBeNull();
  }

  assertInProgressSuccess() {
    this.assertUploadStatus('In Progress');
    this.assertNotResponseStatusCode(201);
    expect(this.response.errorMessage).toBeNull();
  }

  assertSuccess() {
    this.assertUploadStatus('Complete');
    this.assertResponseStatusCode(204);
    expect(this.response.errorMessage).toBeNull();
  }

  assertError(expectedStatusCode: number) {
    this.assertUploadStatus('Failed');
    this.assertResponseStatusCode(expectedStatusCode);
  }
}
