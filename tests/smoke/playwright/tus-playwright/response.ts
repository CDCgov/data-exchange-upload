import { DetailedError, HttpRequest, HttpResponse, Response, UploadResponse } from './index.d';

export class ResponseBuilder {
  private response: Response;

  constructor(filename: string) {
    this.response = {
      filename,
      httpRequests: [],
      httpResponses: []
    };
  }

  addRequest(req: HttpRequest): void {
    this.response.httpRequests.push(req);
  }

  addResponse(res: HttpResponse): void {
    this.response.httpResponses.push(res);
    this.response.uploadStatus = 'In Progress';
    this.response.lastProgressTime = Date.now();
  }

  setProgress(bytesSent: number, bytesTotal: number) {
    this.response.bytesSent = bytesSent;
    this.response.bytesTotal = bytesTotal;
  }

  setChunkComplete(chunkSize: number, bytesAccepted: number, bytesTotal: number): void {
    this.response.chunkSize = chunkSize;
    this.response.bytesAccepted = bytesAccepted;
    this.response.bytesTotal = bytesTotal;
  }

  uploadStarted(): void {
    this.response.uploadStatus = 'Initiated';
    this.response.startTime = Date.now();
  }

  uploadCreated(uploadId: string, uploadUrl: string): void {
    this.response.uploadId = uploadId;
    this.response.uploadUrl = uploadUrl;
  }

  uploadSuccessful(): void {
    this.response.uploadStatus = 'Complete';
    this.response.endTime = Date.now();
  }

  uploadFailure(error: Error | DetailedError) {
    this.response.errorMessage = `${error}`;
    this.response.uploadStatus = 'Failed';
    this.response.endTime = Date.now();
  }

  getResponse(): UploadResponse {
    return this.response;
  }
}
