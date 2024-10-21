import { EventEmitter } from 'events';
import fs from 'fs';
import * as tus from 'tus-js-client';
import { MemoryStorage, PreviousUpload } from './memory-storage';

export type Options = Pick<tus.UploadOptions,
  "headers" |
  "chunkSize" |
  "retryDelays" |
  "onProgress" |
  "onChunkComplete" |
  "onUploadUrlAvailable" |
  "onShouldRetry"
>;
export type DetailedError = tus.DetailedError;

export class UploadClient {
  private file: fs.ReadStream;
  private options: tus.UploadOptions;
  private tusClient: tus.Upload;
  private uploadId: string | null = null;
  private inProgress: boolean | null = null;
  private emitter: EventEmitter;
  private uploadPromise: Promise<{ errorMessage?: string }>;

  constructor(filename: string, metadata: { [key: string]: string }, endpoint: string, storage: MemoryStorage, options: Options = {}) {
    this.emitter = new EventEmitter();
    this.uploadPromise = new Promise((resolve, reject) => {
      this.emitter.on('success', resolve);
      this.emitter.on('error', reject);
    });

    this.file = fs.createReadStream(filename);
    this.options = {
      uploadLengthDeferred: true,
      storeFingerprintForResuming: true,
      removeFingerprintOnSuccess: true,
      retryDelays: [0, 1000, 3000, 5000],
      chunkSize: 40000000,
      metadata,
      endpoint,
      urlStorage: storage,
      ...options,
      headers: {
        "Tus-Resumable": "1.0.0",
        ...options.headers
      }
    }
    this.tusClient = new tus.Upload(this.file, {
      ...this.options,
      onSuccess: this.onSuccess,
      onError: this.onError,
      onProgress: this.onProgress,
      onChunkComplete: this.onChunkComplete,
      onUploadUrlAvailable: this.onUploadUrlAvailable,
    });
  }
  
  getTusClient(): tus.Upload {
    return this.tusClient;
  }

  getEndpoint(): string | null  {
    return this.options.endpoint || null;
  }

  getUploadUrl(): string | null {
    return this.tusClient.url;
  }


  getUploadId(): string | null {
    return this.uploadId;
  }

  getFile(): fs.ReadStream {
    return this.file;
  }

  getOptions(): tus.UploadOptions {
    return this.options;
  }

  getUploadStatus(): string {
    if (this.inProgress == null) {
        return 'Initiated'
    }
    if (this.inProgress == false) {
        return 'Complete'
    }
    return 'In Progress'
  }

  async start(): Promise<{ errorMessage?: string }> {
    this.tusClient.start();
    return this.uploadPromise;
  }

  upload(): Promise<{ errorMessage?: string }> {
    if (this.inProgress == null) {
      this.start();
    } else if (this.inProgress == false) {
      this.resume();
    }
    return this.uploadPromise;
  }

  async pause(): Promise<{ errorMessage?: string }> {
    return this.abort(false);
  }

  async terminate(): Promise<{ errorMessage?: string }> {
    return this.abort(true);
  }

  private async abort(shouldTerminate: boolean): Promise<{ errorMessage?: string }> {
    if (this.inProgress == false) {
      return { errorMessage: `Cannot ${shouldTerminate ? 'terminate' : 'pause'} the upload because the upload is not active` };
    }

    if (this.tusClient.url == null) {
      return { errorMessage: `Cannot ${shouldTerminate ? 'terminate' : 'pause' } until uploadUrl has been created` };
    }

    try {
      await this.tusClient.abort(shouldTerminate);
      this.inProgress = false;
      return {}
    } catch (error: any) {
      return { errorMessage: `${error}` };
    }
  }

  async findPreviousUploads(): Promise<PreviousUpload[]> {
    const tusUploads = await this.tusClient.findPreviousUploads();
    return tusUploads.map((upload) => upload as PreviousUpload);
  }

  resumeFromPreviousUpload(upload: PreviousUpload): void {
    return this.tusClient.resumeFromPreviousUpload(upload);
  }

  async resume(): Promise<{ errorMessage?: string }> {
    if (this.inProgress == true) {
      return { errorMessage: 'Cannot resume the upload because it is active' }
    }

    if (this.tusClient.url == null) {
      return { errorMessage: 'Cannot resume the upload until uploadUrl has been created' };
    }

    const uploadUrl = this.tusClient.url;

    const uploads: PreviousUpload[] = await this.findPreviousUploads();
    const upload = uploads.find(value => value.uploadUrl == uploadUrl);

    if (!upload) {
      return { errorMessage: 'Previous Upload not found' };
    }
    
    this.resumeFromPreviousUpload(upload);
    this.inProgress = true;
    return this.uploadPromise;
  }

  static async terminate(url: string, options: Options = {}): Promise<void> {
    return tus.Upload.terminate(url, options);
  }

  private onProgress(bytesSent: number, bytesTotal: number): void {
    this.useCallback(this.options.onProgress, bytesSent, bytesTotal);
  }

  private onChunkComplete(chunkSize: number, bytesAccepted: number, bytesTotal: number): void {
    this.useCallback(this.options.onChunkComplete, chunkSize, bytesAccepted, bytesTotal);
  }

  private onUploadUrlAvailable(): void {
    if (this.tusClient.url) {
      const uploadIdPattern = /files\/([a-zA-Z0-9]+)/;
      const matches = uploadIdPattern.exec(this.tusClient.url);
      this.uploadId = matches?.[1] ?? null;
    }
    this.useCallback(this.options.onUploadUrlAvailable);
  }

  private onSuccess(): void {
    this.inProgress = false;
    this.emitter.emit('success', {});
  }

  private onError(error: Error | DetailedError): void {
    this.inProgress = false;
    this.emitter.emit('error', { errorMessage: `${error}` });
  }

  private useCallback(cb: any, ...args: any[]): void {
    if (cb != null && typeof cb == 'function') {
      cb(args);
    }
  }
}
