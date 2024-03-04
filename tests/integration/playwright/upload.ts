import fs from 'fs';
import * as tus from 'tus-js-client';

export async function uploadFileAndGetId(accessToken, fileName, url, metadata: Record<string, string>) {
    
    const file = fs.createReadStream(fileName);

    return new Promise((resolve, reject) => {
        const options = {
            endpoint: `${url}/upload`,
            headers: {
                Authorization: `Bearer ${accessToken}`,
            },
            metadata,
            onError: (error) => {
                console.error("An error occurred:", error);
                reject(error);
            },
            onSuccess: () => {
                console.log("Upload finished:", upload.url);

                try {
                    const uploadId = extractUploadId(upload.url);
                    resolve(uploadId);
                } catch (error) {
                    console.error("Failed to extract uploadId:", error);
                    reject(error);
                }
            },
        };

        const upload = new tus.Upload(file, options);
        upload.start();
    });
}

function extractUploadId(uploadUrl) {
    const uploadIdPattern = /files\/([a-zA-Z0-9]+)/;
    const matches = uploadUrl.match(uploadIdPattern);
    if (matches && matches[1]) {
        console.log("Extracted uploadId:", matches[1]);
        return matches[1];
    } else {
        throw new Error("Could not extract uploadId from upload URL.");
    }
}

