import {
  HttpRequest,
  HttpResponse,
  PreviousUpload as TusPreviousUpload,
  UploadOptions as TusUploadOptions
} from 'tus-js-client';

export {
  DetailedError,
  HttpRequest,
  HttpResponse,
  UploadOptions as TusUploadOptions,
  UrlStorage
} from 'tus-js-client';

export type PreviousUpload = TusPreviousUpload & {
  urlStorageKey: string;
  uploadUrl: string | null;
  parallelUploadUrls: string[] | null;
};

export type UploadOptions = Required<
  Pick<TusUploadOptions, 'metadata' | 'endpoint' | 'urlStorage'>
> &
  Pick<
    TusUploadOptions,
    | 'headers'
    | 'retryDelays'
    | 'chunkSize'
    | 'onBeforeRequest'
    | 'onAfterResponse'
    | 'onProgress'
    | 'onChunkComplete'
  > & {
    onUploadStarted?: (response: UploadResponse) => void;
    onUploadCreated?: (response: UploadResponse) => void;
    onUploadPaused?: (response: UploadResponse) => Promise<void>;
    onUploadResumed?: (response: UploadResponse) => void;
    shouldTerminateUpload?: boolean;
  };

export type EventType = 'initiated' | 'created' | 'complete' | 'paused' | 'terminated' | 'resumed';
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
