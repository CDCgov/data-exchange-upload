import fs from 'fs';
import * as tus from 'tus-js-client';


export interface LoginResponse {
    access_token: string;
}

export class UploadClient {
    private baseURL: string;

    constructor(baseURL: string) {
        this.baseURL = baseURL;
    }

    async login(username: string, password: string): Promise<string | null> {
        const params = new URLSearchParams({
            username: username,
            password: password,
        });

        try {
            const response = await fetch(`${this.baseURL}/oauth`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: params,
            });

            if (!response.ok) {
                console.error(`Client login failed, error code is ${response.status}, error message is ${response.statusText}`);
                return null;
            }

            const data = await response.json() as LoginResponse;

            return data.access_token;
        } catch (error) {
            console.error('An error occurred during login:', error);
            return null;
        }
    }

    async uploadFileAndGetId(accessToken: string, fileName: string, metadata: Record<string, string>): Promise<string> {
        const file = fs.createReadStream(fileName);

        return new Promise((resolve, reject) => {
            const options: tus.UploadOptions = {
                endpoint: `${this.baseURL}/upload`,
                headers: {
                    Authorization: `Bearer ${accessToken}`,
                },
                metadata,
                onError: (error: any) => {
                    console.error('An error occurred:', error);
                    reject(error);
                },
                onSuccess: function () {
                    console.log('Upload finished:', upload.url);

                    try {
                        const uploadId = UploadClient.extractUploadId(upload.url);
                        resolve(uploadId);
                    } catch (error) {
                        console.error('Failed to extract uploadId:', error);
                        reject(error);
                    }
                },
            };

            const upload = new tus.Upload(file, options);
            upload.start();
        });
    }

    private static extractUploadId(uploadUrl: string): string {
        const uploadIdPattern = /files\/([a-zA-Z0-9]+)/;
        const matches = uploadIdPattern.exec(uploadUrl);
        const uploadId = matches?.[1] ?? null;
        if (uploadId) {
            console.log('Extracted uploadId:', uploadId);
            return uploadId;
        } else {
            throw new Error('Could not extract uploadId from upload URL.');
        }
    }
}
