import {
  HttpRequest,
  HttpResponse,
  PreviousUpload as TusPreviousUpload,
  UploadOptions as TusUploadOptions
} from 'tus-js-client';

import { ClientRequest, IncomingMessage } from 'http';

export {
  DetailedError,
  HttpRequest,
  HttpResponse,
  UploadOptions as TusUploadOptions,
  UrlStorage
} from 'tus-js-client';

export type RawHttpRequest = ClientRequest;
export type RawHttpResponse = IncomingMessage;

export type PreviousUpload = TusPreviousUpload & {
  urlStorageKey: string;
  uploadUrl: string | null;
  parallelUploadUrls: string[] | null;
};

export type UploadOptions = Required<
  Pick<TusUploadOptions, 'metadata' | 'endpoint' | 'urlStorage'>
> &
  Pick<TusUploadOptions, 'headers' | 'retryDelays'> & {
    onInitiated?: (response: UploadResponse) => void;
    onInProgress?: (response: UploadResponse) => void;
    onComplete: (response: UploadResponse) => void;
  };

export type EventType = 'created' | 'started' | 'completed' | 'paused' | 'terminated' | 'resumed';
export type UploadStatusType = 'Initiated' | 'In Progress' | 'Complete' | 'Failed';
export type UploadResponse = Readonly<Response>;

export type Response = {
  filename: string;
  uploadId?: string;
  uploadUrlId?: string;
  uploadUrl?: string;
  uploadStatus?: UploadStatusType;

  startTime?: number;
  endTime?: number;
  lastProgressTime?: number;

  chunkSize?: number;
  bytesSent?: number;
  bytesAccepted?: number;
  bytesTotal?: number;

  errorMessage?: string;

  httpRequests: HttpRequest[];
  httpResponses: HttpResponse[];
};

export interface ResponseContext {
  getLastRequest(): HttpRequest | null;
  getLastRawRequest(): RawHttpRequest | null;
  getLastResponse(): HttpResponse | null;
  getLastRawResponse(): RawHttpResponse | null;
  getResponseStatusCode(): number | null;
  getResponseBodyString(): string | null;
  getResponseBodyJson(): { [key: string]: string } | null;
  getUploadUrl(): string | null;
  getUploadUrlId(): string | null;
  getUploadId(): string | null;
  getUploadStatus(): string | null;
  assertUploadStatus(expectedStatus: UploadStatusType): void;
  assertResponseStatusCode(expectedStatusCode: number): void;
  assertNotResponseStatusCode(expectedStatusCodeNot: number): void;
  assertResponseBody(expectedBodySubstring: string): void;
  assertResponse(expectedStatusCode: number, expectedBodySubstring: string): void;
  assertSuccess(): void;
  assertError(expectedStatusCode: number): void;
}
