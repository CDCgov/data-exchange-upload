import { EventEmitter } from 'events';
import fs from 'fs';
import * as tus from 'tus-js-client';
import {
  DetailedError,
  EventType,
  HttpRequest,
  HttpResponse,
  PreviousUpload,
  TusUploadOptions,
  UploadOptions,
  UploadResponse
} from './index.d';
import { ResponseBuilder } from './response';

class UploadClient {
  private file: fs.ReadStream;
  private options: TusUploadOptions;
  private emitter: EventEmitter;
  private tusClient: tus.Upload;
  private previousUpload: PreviousUpload | null = null;
  private _isUploading: boolean;

  constructor(filename: string, options: UploadOptions) {
    this.emitter = new EventEmitter();
    this.file = fs.createReadStream(filename);

    this.options = {
      uploadLengthDeferred: true,
      storeFingerprintForResuming: true,
      removeFingerprintOnSuccess: true,
      retryDelays: options.retryDelays ?? [0, 1000, 3000, 5000],
      chunkSize: options.chunkSize ?? 40000000,
      ...options,
      headers: {
        'Tus-Resumable': '1.0.0',
        'Content-Length': '0',
        ...options.headers
      },
      onUploadUrlAvailable: () => this.onUploadUrlAvailable(),
      onSuccess: () => this.onSuccess(),
      onError: (error: Error | DetailedError) => this.onError(error)
    };

    this.tusClient = new tus.Upload(this.file, this.options);
    this._isUploading = false;
  }

  private onUploadUrlAvailable(): void {
    const uploadUrl = this.tusClient.url;

    if (uploadUrl) {
      const uploadUrlId = uploadUrl.split('/').slice(-1)[0];
      const uploadIdPattern = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/;
      const uploadId = uploadIdPattern.exec(uploadUrl)?.[0] ?? undefined;
      this.emit('created', uploadId, uploadUrlId, uploadUrl);

      this.findPreviousUploadFromUploadUrl(uploadUrl).then(upload => {
        this.previousUpload = upload;
      });
    } else {
      setTimeout(() => this.onUploadUrlAvailable(), 1000);
    }
  }

  private onSuccess(): void {
    this._isUploading = false;

    this.emit('complete');
  }

  private onError(error: Error | DetailedError) {
    this._isUploading = false;

    this.emit('complete', error);
  }

  addListener(event: EventType, fn: (...args: any[]) => void) {
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

    this._isUploading = true;
    this.tusClient.start();
    this.emit('initiated');
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
      this._isUploading = false;
      await this.tusClient.abort(shouldTerminate);
      this.emit(shouldTerminate ? 'terminated' : 'paused');
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

    this._isUploading = true;
    this.tusClient.resumeFromPreviousUpload(upload);
    this.tusClient.start();
    this.emit('resumed');
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
  const {
    onUploadStarted,
    onUploadCreated,
    onUploadPaused,
    onUploadResumed,
    shouldTerminateUpload,
    ...configs
  } = options;
  const builder = new ResponseBuilder(filename);
  const client = new UploadClient(filename, {
    ...configs,
    onBeforeRequest: (req: HttpRequest) => {
      builder.addRequest(req);
      if (configs.onBeforeRequest && typeof configs.onBeforeRequest == 'function') {
        configs.onBeforeRequest(req);
      }
    },
    onAfterResponse: (req: HttpRequest, res: HttpResponse) => {
      builder.addResponse(res);
      if (configs.onAfterResponse && typeof configs.onAfterResponse == 'function') {
        configs.onAfterResponse(req, res);
      }
    },
    onProgress: (bytesSent: number, bytesTotal: number) => {
      builder.setProgress(bytesSent, bytesTotal);
      if (configs.onProgress && typeof configs.onProgress == 'function') {
        configs.onProgress(bytesSent, bytesTotal);
      }
    },
    onChunkComplete: (chunkSize: number, bytesAccepted: number, bytesTotal: number) => {
      builder.setChunkComplete(chunkSize, bytesAccepted, bytesTotal);
      if (configs.onChunkComplete && typeof configs.onChunkComplete == 'function') {
        configs.onChunkComplete(chunkSize, bytesAccepted, bytesTotal);
      }
    }
  });

  return new Promise(resolve => {
    client.addListener('initiated', () => {
      builder.uploadStarted();
      if (onUploadStarted && typeof onUploadStarted == 'function') {
        onUploadStarted(builder.getResponse());
      }
    });

    client.addListener('created', (uploadId: string, uploadUrlId: string, uploadUrl: string) => {
      builder.uploadCreated(uploadId, uploadUrlId, uploadUrl);
      if (onUploadCreated && typeof onUploadCreated == 'function') {
        onUploadCreated(builder.getResponse());
      }

      if (onUploadPaused && typeof onUploadPaused == 'function') {
        client.addListener('paused', () => {
          onUploadPaused(builder.getResponse())?.then(() => client.resume());
        });

        client.pause();
      } else if (shouldTerminateUpload) {
        client.addListener('terminated', () => {
          resolve(builder.getResponse());
        });

        client.terminate();
      }
    });

    if (onUploadResumed && typeof onUploadResumed == 'function') {
      onUploadResumed(builder.getResponse());
    }

    client.addListener('complete', (error?: Error | DetailedError) => {
      if (error) {
        builder.uploadFailure(error);
      } else {
        builder.uploadSuccessful();
      }
      resolve(builder.getResponse());
    });

    client.upload();
  });
}
