import { ReadStream, createReadStream, statSync } from 'fs';
import * as tus from 'tus-js-client';
import {
  DetailedError,
  HttpRequest,
  HttpResponse,
  PreviousUpload,
  TusUploadOptions,
  UploadOptions,
  UploadResponse,
  UploadStatusType
} from './index.d';
import { ResponseBuilder } from './response';

export class UploadClient {
  private options: TusUploadOptions;
  private tusClient: tus.Upload;
  private builder: ResponseBuilder;
  private previousUpload: PreviousUpload | null = null;
  private status: UploadStatusType | null = null;
  private _isUploading: boolean = false;
  private _initiatedPromise: Promise<UploadResponse>;
  private _inProgressPromise: Promise<UploadResponse>;
  private _completedPromise: Promise<UploadResponse>;

  constructor(filename: string, options: UploadOptions) {
    this.builder = new ResponseBuilder(filename);

    let _resolveInitiated: (value: UploadResponse) => void;
    this._initiatedPromise = new Promise(resolve => {
      _resolveInitiated = resolve;
    });

    let _resolveInProgress: (value: UploadResponse) => void;
    this._inProgressPromise = new Promise(resolve => {
      _resolveInProgress = resolve;
    });

    let _resolveCompleted: (value: UploadResponse) => void;
    this._completedPromise = new Promise(resolve => {
      _resolveCompleted = resolve;
    });

    const file: ReadStream = createReadStream(filename);
    const fileSize = statSync(filename)?.size;
    // the default chunkSize is half of the file size so that
    // we can support waiting for an In Progress status
    const defaultChunk: number = fileSize / 2;

    this.options = {
      storeFingerprintForResuming: true,
      removeFingerprintOnSuccess: true,
      metadata: options.metadata,
      endpoint: options.endpoint,
      urlStorage: options.urlStorage,
      retryDelays: options.retryDelays ?? [0, 1000, 3000, 5000],
      chunkSize: options.chunkSize ?? defaultChunk,
      headers: {
        'Tus-Resumable': '1.0.0',
        ...options.headers
      },
      onBeforeRequest: (req: HttpRequest): void => {
        this.builder.addRequest(req);
      },
      onAfterResponse: (req: HttpRequest, res: HttpResponse): void => {
        this.builder.addResponse(res);
      },
      onProgress: (bytesSent: number, bytesTotal: number): void => {
        this.builder.setProgress(bytesSent, bytesTotal);
      },
      onChunkComplete: (chunkSize: number, bytesAccepted: number, bytesTotal: number): void => {
        // only need to report In Progress the first time
        if (chunkSize == bytesAccepted) {
          this.builder.setUploadStatus('In Progress');
          this.status = 'In Progress';
        }

        // resolve the promise waiting for the In Progress status
        _resolveInProgress(this.builder.getResponse());
        this.builder.setChunkComplete(chunkSize, bytesAccepted, bytesTotal);
      },
      onUploadUrlAvailable: (): void => {
        const uploadUrl = this.tusClient?.url ?? '';
        const uploadUrlId = uploadUrl?.split('/')?.slice(-1)[0] ?? '';
        const uploadIdPattern = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/;
        const uploadId = uploadIdPattern.exec(uploadUrl)?.[0] ?? '';

        this.builder.uploadCreated();
        this.builder.setUploadId(uploadId, uploadUrlId, uploadUrl);
        this.status = 'Initiated';

        // resolve the promise waiting for the Initiated status
        _resolveInitiated(this.builder.getResponse());

        this.findPreviousUploadFromUploadUrl(uploadUrl).then(upload => {
          this.previousUpload = upload;
        });
      },
      onSuccess: () => {
        this.builder.uploadSuccessful();
        this._isUploading = false;
        this.status = 'Complete';

        // resolve the promise waiting for the Complete status
        _resolveCompleted(this.builder.getResponse());
      },
      onError: (error: Error | DetailedError) => {
        this.builder.uploadFailure(error);
        this._isUploading = false;
        this.status = 'Failed';

        // resolve the promise waiting for the Complete status
        _resolveCompleted(this.builder.getResponse());
      }
    };

    this.tusClient = new tus.Upload(file, this.options);
  }

  async uploadComplete(): Promise<UploadResponse> {
    if (this.status == 'Complete') {
      return this.builder.getResponse();
    }
    this.start();
    return this._completedPromise;
  }

  async uploadInitiated(): Promise<UploadResponse> {
    if (this.status != null) {
      this.builder.setErrorMessage(
        'Cannot pause at Initiated because the upload was already initiated'
      );
      return this.builder.getResponse();
    }
    this.start();
    return new Promise(resolve => {
      this._initiatedPromise.then(response => {
        // once the upload is initiated pause the upload so that tests can be performed on this status
        this.pause().then(() => {
          resolve(response);
        });
      });
    });
  }

  async uploadInProgress(): Promise<UploadResponse> {
    if (this.status == 'Complete') {
      this.builder.setErrorMessage('Cannot pause at In Progress because the upload is complete');
      return this.builder.getResponse();
    }
    this.start();
    return new Promise(resolve => {
      this._inProgressPromise.then(response => {
        // once the upload is in progress pause the upload
        // so that tests can be performed on this status
        this.pause().then(() => {
          // waiting for 1 second for the file system to update
          setTimeout(() => resolve(response), 1000);
        });
      });
    });
  }

  async start(): Promise<void> {
    if (!this._isUploading) {
      if (this.previousUpload) {
        this.tusClient.resumeFromPreviousUpload(this.previousUpload);
      } else if (this.tusClient.url) {
        const upload: PreviousUpload | null = await this.findPreviousUploadFromUploadUrl(
          this.tusClient.url
        );
        if (upload) {
          this.tusClient.resumeFromPreviousUpload(upload);
        }
      }

      this.tusClient.start();
      this._isUploading = true;
    }
  }

  async pause(): Promise<void> {
    if (this._isUploading) {
      this.tusClient.abort(false);
      this._isUploading = false;
    }
  }

  async resume(): Promise<void> {
    return this.start();
  }

  async terminate(): Promise<void> {
    this.tusClient.abort(true);
  }

  private async findPreviousUploads(): Promise<PreviousUpload[]> {
    const tusUploads = await this.tusClient.findPreviousUploads();
    return tusUploads.map(upload => upload as PreviousUpload);
  }

  private async findPreviousUploadFromUploadUrl(uploadUrl: string): Promise<PreviousUpload | null> {
    const uploads: PreviousUpload[] = await this.findPreviousUploads();
    const upload = uploads.find(value => value.uploadUrl == uploadUrl);

    return upload ?? null;
  }

  static async removeUpload(url: string): Promise<void> {
    return tus.Upload.terminate(url, {});
  }
}
