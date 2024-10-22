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

export type EventType = 'created' | 'updated' | 'complete' | 'paused' | 'terminated' | 'resumed';
export type UploadStatusType = 'Initiated' | 'In Progress' | 'Complete' | 'Failed';
export type UploadResponse = Readonly<Response>;

function initResponse(filename: string): Response {
  return {
    filename: filename,
    uploadId: null,
    uploadUrl: null,
    uploadStatus: null,

    startTime: null,
    endTime: null,
    lastProgressTime: null,

    chunkSize: null,
    bytesSent: null,
    bytesAccepted: null,
    bytesTotal: null,

    errorMessage: null,

    httpRequests: [],
    httpResponses: []
  };
}

type Response = {
  filename: string;
  uploadId: string | null;
  uploadUrl: string | null;
  uploadStatus: UploadStatusType | null;

  startTime: number | null;
  endTime: number | null;
  lastProgressTime: number | null;

  chunkSize: number | null;
  bytesSent: number | null;
  bytesAccepted: number | null;
  bytesTotal: number | null;

  errorMessage: string | null;

  httpRequests: tus.HttpRequest[];
  httpResponses: tus.HttpResponse[];
};

class UploadClient {
  private file: fs.ReadStream;
  private options: tus.UploadOptions;
  private emitter: EventEmitter;
  private tusClient: tus.Upload;
  private response: Response;
  private _isUploading: boolean;

  constructor(filename: string, options: UploadOptions) {
    this.emitter = new EventEmitter();
    this.file = fs.createReadStream(filename);
    this.response = initResponse(filename);

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

    this.emit('updated', this.response);
  }

  private onAfterResponse(req: tus.HttpRequest, res: tus.HttpResponse) {
    this.response.httpResponses.push(res);

    this._isUploading = true;
    this.response.uploadStatus = 'In Progress';
    this.response.lastProgressTime = Date.now();

    this.emit('updated', this.response);
  }

  private onProgress(bytesSent: number, bytesTotal: number) {
    this.response.bytesSent = bytesSent;
    this.response.bytesTotal = bytesTotal;

    this.emit('updated', this.response);
  }

  private onChunkComplete(chunkSize: number, bytesAccepted: number, bytesTotal: number): void {
    this.response.chunkSize = chunkSize;
    this.response.bytesAccepted = bytesAccepted;
    this.response.bytesTotal = bytesTotal;

    this.emit('updated', this.response);
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
    }

    this.emit('created', this.response);
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

  async start(): Promise<{ errorMessage?: string }> {
    this.tusClient.start();
    return {};
  }

  async upload(): Promise<{ errorMessage?: string }> {
    if (this._isUploading == true) {
      return {
        errorMessage: 'Cannot start the upload because it is already active'
      };
    }

    if (this.tusClient.url) {
      return this.resume();
    } else {
      return this.start();
    }
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
      this.emit(shouldTerminate ? 'terminated' : 'paused', this.response);
      this._isUploading = false;
      return {};
    } catch (error: any) {
      return { errorMessage: `${error}` };
    }
  }

  async findPreviousUploads(): Promise<PreviousUpload[]> {
    const tusUploads = await this.tusClient.findPreviousUploads();
    return tusUploads.map(upload => upload as PreviousUpload);
  }

  resumeFromPreviousUpload(upload: PreviousUpload): void {
    return this.tusClient.resumeFromPreviousUpload(upload);
  }

  async resume(): Promise<{ errorMessage?: string }> {
    if (this._isUploading == true) {
      return { errorMessage: 'Cannot resume the upload because it is active' };
    }

    if (this.tusClient.url == null) {
      return {
        errorMessage: 'Cannot resume the upload until uploadUrl has been created'
      };
    }

    const uploadUrl = this.tusClient.url;

    const uploads: PreviousUpload[] = await this.findPreviousUploads();
    const upload = uploads.find(value => value.uploadUrl == uploadUrl);

    if (!upload) {
      return { errorMessage: 'Previous Upload not found' };
    }

    this.resumeFromPreviousUpload(upload);
    this.emit('resumed', this.response);
    this._isUploading = true;
    return {};
  }

  static async terminate(url: string, options: tus.UploadOptions): Promise<void> {
    return tus.Upload.terminate(url, options);
  }
}

export async function uploadFile(
  filename: string,
  options: UploadOptions
): Promise<UploadResponse> {
  const client = new UploadClient(filename, options);
  return new Promise(resolve => {
    client.addListener('complete', response => {
      resolve(response);
    });

    client.upload();
  });
}

export async function startUploadAndPause(
  filename: string,
  options: UploadOptions
): Promise<UploadResponse> {
  const client = new UploadClient(filename, options);
  return new Promise(resolve => {
    client.addListener('paused', response => resolve(response));
    client.addListener('created', () => {
      client.pause();
    });

    client.upload();
  });
}

export async function startUploadAndPauseAndResume(
  filename: string,
  options: UploadOptions,
  cb?: (response: UploadResponse) => Promise<void>
): Promise<UploadResponse> {
  const client = new UploadClient(filename, options);
  return new Promise(resolve => {
    client.addListener('resumed', response => resolve(response));
    client.addListener('paused', response => {
      if (cb && typeof cb == 'function') {
        cb(response).then(() => client.resume());
      } else {
        client.resume();
      }
    });
    client.addListener('created', () => {
      client.pause();
    });

    client.upload();
  });
}
