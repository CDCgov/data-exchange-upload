import { PreviousUpload, UrlStorage } from './index.d';

export class MemoryStorage implements UrlStorage {
  private uploadMap: Map<string, PreviousUpload>;

  constructor() {
    this.uploadMap = new Map<string, PreviousUpload>();
  }

  findAllUploads(): Promise<PreviousUpload[]> {
    const results: PreviousUpload[] = Array.from(this.uploadMap.values()) || [];
    return new Promise(resolve => resolve(results));
  }

  findUploadsByFingerprint(fingerprint: string): Promise<PreviousUpload[]> {
    return new Promise(resolve => {
      const results: PreviousUpload[] = [];

      this.uploadMap.forEach((value, key) => {
        if (key.indexOf(fingerprint) == 0) {
          results.push(value);
        }
      });

      resolve(results);
    });
  }

  findUploadByUploadUrl(uploadUrl: string): Promise<PreviousUpload | null> {
    return new Promise(resolve => {
      this.uploadMap.forEach(value => {
        if (value.uploadUrl == uploadUrl) {
          resolve(value);
        }
      });

      resolve(null);
    });
  }

  removeUpload(urlStorageKey: string): Promise<void> {
    return new Promise(resolve => {
      const uploadsToDelete: string[] = [];

      this.uploadMap.forEach((value, key) => {
        if (value.urlStorageKey == urlStorageKey) {
          uploadsToDelete.push(key);
        }
      });

      uploadsToDelete.forEach(key => this.uploadMap.delete(key));

      resolve();
    });
  }

  // Returns the URL storage key, which can be used for removing the upload.
  addUpload(fingerprint: string, upload: PreviousUpload): Promise<string> {
    return new Promise(resolve => {
      const id = Math.round(Math.random() * 1e12);
      const key = `${fingerprint}_${id}`;

      this.uploadMap.set(key, upload);

      resolve(key);
    });
  }
}
