import { MemoryStorage } from "./memory-storage";
import { DetailedError, UploadClient } from "./upload-client";

export type OnProgressCallback = ((bytesSent: number, bytesTotal: number) => void)
export type OnChunkCompleteCallback = ((chunkSize: number, bytesAccepted: number, bytesTotal: number) => void)
export type OnShouldRetryCallback = ((error: DetailedError, retryAttempt: number) => boolean)
export type OnUploadUrlAvailable = (() => void)

export type ClientOptions = {
    authHeaders?: { [key: string]: string } | null
    retryDelays?: number[] | null
    onProgress?: OnProgressCallback | null
    onChunkComplete?: OnChunkCompleteCallback | null
    onShouldRetry?: OnShouldRetryCallback | null
    onUploadUrlAvailable?: OnUploadUrlAvailable | null
}

export function newContext(baseUrl: string): ClientContext {
    return new ClientContext(baseUrl);
}

class ClientContext {
    private baseURL: string;
    private storage: MemoryStorage;
    private options: ClientOptions;
    
    constructor(baseURL: string, options: ClientOptions = {}) {
        this.baseURL = baseURL;
        this.storage = new MemoryStorage();
        this.options = options;
    }
    
    setAuthenticationHeaders(authHeaders: { [key: string]: string }): void {
        this.options.authHeaders = authHeaders;
    }

    setRetryDelays(retryDelays: number[]) {
        this.options.retryDelays = retryDelays;
    }

    setOnProgress(cb: OnProgressCallback): void {
        this.options.onProgress = cb;
    }

    setOnChunkComplete(cb: OnChunkCompleteCallback): void {
        this.options.onChunkComplete = cb;
    }

    setOnShouldRetry(cb: OnShouldRetryCallback): void {
        this.options.onShouldRetry = cb;
    }

    setOnUploadUrlAvailable(cb: OnUploadUrlAvailable): void {
        this.options.onUploadUrlAvailable = cb;
    }

    newUploadContext(filename: string, metadata: { [key: string]: string }) {
        return new UploadContext(filename, metadata, this.baseURL, this.storage, this.options);
    }

    async upload(filename: string, metadata: { [key: string]: string }): Promise<{ errorMessage?: string }> {
        const client = new UploadClient(filename, metadata, this.baseURL, this.storage, this.options);
        return client.upload();
    }
}

class UploadContext {
    private client: UploadClient;

    constructor(filename: string, metadata: { [key: string]: string }, endpoint: string, storage: MemoryStorage, options: ClientOptions) {
        this.client = new UploadClient(filename, metadata, endpoint, storage, options);
    }

    upload(): Promise<{ errorMessage?: string }> {
        return this.client.upload();
    }

    async pause(): Promise<{ errorMessage?: string }> {
        return this.client.pause();
    }

    async resume(): Promise<{ errorMessage?: string }> {
        return this.client.resume();
    }

    async terminate(): Promise<{ errorMessage?: string }> {
        return this.client.terminate();
    }
}
