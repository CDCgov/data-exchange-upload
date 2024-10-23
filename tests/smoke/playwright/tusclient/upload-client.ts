import { EventEmitter } from 'events';
import fs from 'fs';
import * as tus from 'tus-js-client';
import { MemoryStorage, PreviousUpload } from './memory-storage';

type DetailedError = tus.DetailedError;
export type HttpRequest = tus.HttpRequest;
export type HttpResponse = tus.HttpResponse;

export type UploadOptions = {
  metadata: { [key: string]: string };
  endpoint: string;
  urlStorage: MemoryStorage;

  headers?: { [key: string]: string } | null;
  retryDelays?: number[] | null;
  chunkSize?: number | null;
};

export type UploadHooks = {
  onUploadInitiated?: (response: UploadResponse) => void;
  onUploadCreated?: (response: UploadResponse) => void;
  onRequestSent?: (response: UploadResponse) => void;
  onResponseReceived?: (response: UploadResponse) => void;
  onChunkSent?: (response: UploadResponse) => void;
  onChunkComplete?: (response: UploadResponse) => void;
  onUploadPaused?: (response: UploadResponse) => Promise<void>;
  onUploadResumed?: (response: UploadResponse) => void;
  shouldTerminateUpload?: boolean;
};

type EventType =
  | 'initiated'
  | 'created'
  | 'request'
  | 'response'
  | 'chunkSent'
  | 'chunkComplete'
  | 'complete'
  | 'paused'
  | 'terminated'
  | 'resumed';
export type UploadStatusType = 'Initiated' | 'In Progress' | 'Complete' | 'Failed';
export type UploadResponse = Readonly<Response>;

type Response = {
  filename: string;
  uploadId?: string;
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

  httpRequests: tus.HttpRequest[];
  httpResponses: tus.HttpResponse[];
};

class UploadClient {
  private file: fs.ReadStream;
  private options: tus.UploadOptions;
  private emitter: EventEmitter;
  private tusClient: tus.Upload;
  private previousUpload: PreviousUpload | null = null;
  private response: Response;
  private _isUploading: boolean;

  constructor(filename: string, options: UploadOptions) {
    this.emitter = new EventEmitter();
    this.file = fs.createReadStream(filename);
    this.response = {
      filename,
      httpRequests: [],
      httpResponses: []
    };

    this.options = {
      uploadLengthDeferred: true,
      storeFingerprintForResuming: true,
      removeFingerprintOnSuccess: true,
      retryDelays: options.retryDelays ?? [0, 1000, 3000, 5000],
      chunkSize: options.chunkSize ?? 40000000,
      metadata: options.metadata,
      endpoint: options.endpoint,
      urlStorage: options.urlStorage,
      headers: {
        'Tus-Resumable': '1.0.0',
        'Content-Length': '0',
        ...options.headers
      },
      onBeforeRequest: (req: tus.HttpRequest) => this.onBeforeRequest(req),
      onAfterResponse: (req: tus.HttpRequest, res: tus.HttpResponse) =>
        this.onAfterResponse(req, res),
      onProgress: (bytesSent: number, bytesTotal: number) => this.onProgress(bytesSent, bytesTotal),
      onChunkComplete: (chunkSize: number, bytesAccepted: number, bytesTotal: number) =>
        this.onChunkComplete(chunkSize, bytesAccepted, bytesTotal),
      onUploadUrlAvailable: () => this.onUploadUrlAvailable(),
      onSuccess: () => this.onSuccess(),
      onError: (error: Error | DetailedError) => this.onError(error)
    };
    this.tusClient = new tus.Upload(this.file, this.options);
    this._isUploading = false;
    this.response.uploadStatus = 'Initiated';
    this.response.startTime = Date.now();
  }

  private onBeforeRequest(req: tus.HttpRequest) {
    this.response.httpRequests.push(req);

    this.emit('request', this.response);
  }

  private onAfterResponse(req: tus.HttpRequest, res: tus.HttpResponse) {
    this.response.httpResponses.push(res);

    this._isUploading = true;
    this.response.uploadStatus = 'In Progress';
    this.response.lastProgressTime = Date.now();

    this.emit('response', this.response);
  }

  private onProgress(bytesSent: number, bytesTotal: number) {
    this.response.bytesSent = bytesSent;
    this.response.bytesTotal = bytesTotal;

    this.emit('chunkSent', this.response);
  }

  private onChunkComplete(chunkSize: number, bytesAccepted: number, bytesTotal: number): void {
    this.response.chunkSize = chunkSize;
    this.response.bytesAccepted = bytesAccepted;
    this.response.bytesTotal = bytesTotal;

    this.emit('chunkComplete', this.response);
  }

  private onUploadUrlAvailable(): void {
    if (this.tusClient.url) {
      this.response.uploadUrl = this.tusClient.url;

      const uploadIdPattern = /files\/([a-zA-Z0-9]+)/;
      const matches = uploadIdPattern.exec(this.response.uploadUrl);
      const uploadId = matches?.[1] ?? null;
      if (uploadId) {
        this.response.uploadId = uploadId;
      }
      this.findPreviousUploadFromUploadUrl(this.tusClient.url).then(upload => {
        this.previousUpload = upload;
      });
      this.emit('created', this.response);
    } else {
      setTimeout(() => this.onUploadUrlAvailable(), 1000);
    }
  }

  private onSuccess(): void {
    this._isUploading = false;
    this.response.uploadStatus = 'Complete';
    this.response.endTime = Date.now();

    this.emit('complete', this.response);
  }

  private onError(error: Error | DetailedError) {
    this._isUploading = false;
    this.response.uploadStatus = 'Failed';
    this.response.endTime = Date.now();
    this.response.errorMessage = `${error}`;

    this.emit('complete', this.response);
  }

  addListener(event: EventType, fn: (response: UploadResponse) => void) {
    this.emitter.on(event, fn);
  }

  private emit(event: EventType, ...args: any[]) {
    this.emitter.emit(event, ...args);
  }

  isUploading(): boolean {
    return this._isUploading;
  }

  async upload(): Promise<{ errorMessage?: string }> {
    if (this.previousUpload || this.tusClient.url) {
      return this.resume();
    } else {
      return this.start();
    }
  }

  async start(): Promise<{ errorMessage?: string }> {
    if (this._isUploading == true) {
      return {
        errorMessage: 'Cannot start the upload because it is already active'
      };
    }

    this.tusClient.start();
    this._isUploading = true;
    this.emit('initiated', this.response);
    return {};
  }

  async pause(): Promise<{ errorMessage?: string }> {
    return this.abort(false);
  }

  async terminate(): Promise<{ errorMessage?: string }> {
    return this.abort(true);
  }

  private async abort(shouldTerminate: boolean): Promise<{ errorMessage?: string }> {
    if (this._isUploading == false) {
      return {
        errorMessage: `Cannot ${shouldTerminate ? 'terminate' : 'pause'} the upload because the upload is not active`
      };
    }

    if (this.tusClient.url == null) {
      return {
        errorMessage: `Cannot ${shouldTerminate ? 'terminate' : 'pause'} until uploadUrl has been created`
      };
    }

    try {
      await this.tusClient.abort(shouldTerminate);
      this._isUploading = false;
      this.emit(shouldTerminate ? 'terminated' : 'paused', this.response);
      return {};
    } catch (error: any) {
      return { errorMessage: `${error}` };
    }
  }

  private async findPreviousUploads(): Promise<PreviousUpload[]> {
    const tusUploads = await this.tusClient.findPreviousUploads();
    return tusUploads.map(upload => upload as PreviousUpload);
  }

  async findPreviousUploadFromUploadUrl(uploadUrl: string): Promise<PreviousUpload | null> {
    const uploads: PreviousUpload[] = await this.findPreviousUploads();
    const upload = uploads.find(value => value.uploadUrl == uploadUrl);

    return upload ?? null;
  }

  async resume(): Promise<{ errorMessage?: string }> {
    if (this._isUploading == true) {
      return { errorMessage: 'Cannot resume the upload because it is already active' };
    }

    if (this.tusClient.url == null) {
      return {
        errorMessage: 'Cannot resume the upload until uploadUrl has been created'
      };
    }

    let upload: PreviousUpload | null = null;
    if (this.previousUpload) {
      upload = this.previousUpload;
    } else {
      upload = await this.findPreviousUploadFromUploadUrl(this.tusClient.url);
    }

    if (!upload) {
      return { errorMessage: 'Previous Upload not found' };
    }

    this.tusClient.resumeFromPreviousUpload(upload);
    this.tusClient.start();
    this._isUploading = true;
    this.emit('resumed', this.response);
    return {};
  }

  static async terminate(url: string, options: tus.UploadOptions): Promise<void> {
    return tus.Upload.terminate(url, options);
  }
}

export async function uploadFile(
  filename: string,
  options: UploadOptions,
  {
    onUploadInitiated,
    onUploadCreated,
    onRequestSent,
    onResponseReceived,
    onChunkSent,
    onChunkComplete,
    onUploadPaused,
    onUploadResumed,
    shouldTerminateUpload
  }: UploadHooks = {}
): Promise<UploadResponse> {
  const client = new UploadClient(filename, options);
  return new Promise(resolve => {
    client.addListener('initiated', response => {
      if (onUploadInitiated && typeof onUploadInitiated == 'function') {
        onUploadInitiated(response);
      }
    });

    client.addListener('created', response => {
      if (onUploadCreated && typeof onUploadCreated == 'function') {
        onUploadCreated(response);
      }

      if (shouldTerminateUpload) {
        client.addListener('terminated', response => {
          resolve(response);
        });

        client.terminate();
      } else if (onUploadPaused && typeof onUploadPaused == 'function') {
        client.addListener('paused', response => {
          onUploadPaused(response).then(() => client.resume());
        });

        client.pause();
      }
    });
    client.addListener('request', response => {
      if (onRequestSent && typeof onRequestSent == 'function') {
        onRequestSent(response);
      }
    });
    client.addListener('response', response => {
      if (onResponseReceived && typeof onResponseReceived == 'function') {
        onResponseReceived(response);
      }
    });
    client.addListener('chunkSent', response => {
      if (onChunkSent && typeof onChunkSent == 'function') {
        onChunkSent(response);
      }
    });
    client.addListener('chunkComplete', response => {
      if (onChunkComplete && typeof onChunkComplete == 'function') {
        onChunkComplete(response);
      }
    });
    client.addListener('resumed', response => {
      if (onUploadResumed && typeof onUploadResumed == 'function') {
        onUploadResumed(response);
      }
    });

    client.addListener('complete', response => {
      resolve(response);
    });

    client.upload();
  });
}
