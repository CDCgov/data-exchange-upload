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
    // TODO create a custom FileReader
    // the way we are opening files does not allow for chunking, we need to implement our
    // own FileReader so that we can upload the file in chunks
    // this chunk setting is here so that we can test out uploads that are in progress,
    // otherwise they immediately goes from Initiated to Complete
    const chunkSize =
      onInProgress && typeof onInProgress == 'function' ? Math.floor(fileSize / 2) : Infinity;

    this.options = {
      uploadLengthDeferred: true,
      storeFingerprintForResuming: true,
      removeFingerprintOnSuccess: true,
      chunkSize,
      metadata: options.metadata,
      endpoint: options.endpoint,
      urlStorage: options.urlStorage,
      retryDelays: options.retryDelays ?? [0, 1000, 3000, 5000],
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
        // if the chunkSize and bytesAccepted match, that means the first chunk has been sent
        // and the status should have switched to In Progress
        if (chunkSize == bytesAccepted) {
          this.builder.setUploadStatus('In Progress');
          // if the onInProgress callback has been set
          // pause the upload, wait a second for the file
          // system to catch up and then call the callback
          // we must pause so that the upload doesn't complete
          // while we are testing the status
          if (onInProgress && typeof onInProgress == 'function') {
            this.pause().then(() => {
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

        // if the onInitiated callback has been set
        // pause the upload and then call the callback
        // we must pause so that the upload doesn't complete
        // while we are testing the status
        if (onInitiated && typeof onInitiated == 'function') {
          this.pause().then(() => {
            onInitiated(this.builder.getResponse());
          });
        }
      },
      onSuccess: () => {
        this.builder.uploadSuccessful();

        // if the onComplete callback has been set,
        // then call the callback when the upload is complete
        if (onComplete && typeof onComplete == 'function') {
          onComplete(this.builder.getResponse());
        }
      },
      onError: (error: Error | DetailedError) => {
        this.builder.uploadFailure(error);

        // if the onComplete callback has been set,
        // then call the callback when the upload has an error
        if (onComplete && typeof onComplete == 'function') {
          onComplete(this.builder.getResponse());
        }
      }
    };

    this.tusClient = new tus.Upload(this.file, this.options);
  }

  upload(): void {
    this.tusClient.start();
  }

  private async pause(): Promise<void> {
    this.tusClient.abort(false);
  }

  static async terminate(url: string, options: tus.UploadOptions = {}): Promise<void> {
    return tus.Upload.terminate(url, options);
  }
}
