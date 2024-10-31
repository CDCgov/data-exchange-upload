import fs from 'fs';
import * as tus from 'tus-js-client';
import {
  DetailedError,
  HttpRequest,
  HttpResponse,
  TusUploadOptions,
  UploadOptions
} from './index.d';
import { ResponseBuilder } from './response';

export class UploadClient {
  private file: fs.ReadStream;
  private options: TusUploadOptions;
  private tusClient: tus.Upload;
  private builder: ResponseBuilder;

  constructor(filename: string, options: UploadOptions) {
    this.file = fs.createReadStream(filename);
    this.builder = new ResponseBuilder(filename);

    const fileSize = fs.statSync(filename).size;
    const { onInitiated, onInProgress, onComplete } = options;

    this.options = {
      uploadLengthDeferred: true,
      storeFingerprintForResuming: true,
      removeFingerprintOnSuccess: true,
      metadata: options.metadata,
      endpoint: options.endpoint,
      urlStorage: options.urlStorage,
      retryDelays: options.retryDelays ?? [0, 1000, 3000, 5000],
      chunkSize:
        onInProgress && typeof onInProgress == 'function' ? Math.floor(fileSize / 2) : 40000,
      headers: {
        'Tus-Resumable': '1.0.0',
        'Content-Length': '0',
        'Upload-Defer-Length': '1',
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
        if (chunkSize == bytesAccepted) {
          this.builder.setUploadStatus('In Progress');
          if (onInProgress && typeof onInProgress == 'function') {
            this.tusClient.abort(false).then(() => {
              setTimeout(() => {
                onInProgress(this.builder.getResponse());
              }, 1000);
            });
          }
        }
        this.builder.setChunkComplete(chunkSize, bytesAccepted, bytesTotal);
      },
      onUploadUrlAvailable: (): void => {
        this.builder.uploadCreated();

        const uploadUrl = this.tusClient?.url ?? '';
        const uploadUrlId = uploadUrl?.split('/')?.slice(-1)[0] ?? '';
        const uploadIdPattern = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/;
        const uploadId = uploadIdPattern.exec(uploadUrl)?.[0] ?? '';

        this.builder.setUploadId(uploadId, uploadUrlId, uploadUrl);
        if (onInitiated && typeof onInitiated == 'function') {
          this.tusClient.abort(false).then(() => {
            onInitiated(this.builder.getResponse());
          });
        }
      },
      onSuccess: () => {
        this.builder.uploadSuccessful();
        onComplete(this.builder.getResponse());
      },
      onError: (error: Error | DetailedError) => {
        this.builder.uploadFailure(error);
        onComplete(this.builder.getResponse());
      }
    };

    this.tusClient = new tus.Upload(this.file, this.options);
  }

  upload(): void {
    this.tusClient.start();
  }

  async terminate(): Promise<void> {
    await this.tusClient.abort(true);
  }

  static async terminate(url: string, options: tus.UploadOptions = {}): Promise<void> {
    return tus.Upload.terminate(url, options);
  }
}
