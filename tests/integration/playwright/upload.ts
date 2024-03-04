import fs from 'fs';
import path from 'path';
import * as tus from 'tus-js-client';
import { v4 as uuidv4 } from 'uuid';
import config from './config';

export async function uploadFileAndGetId(accessToken) {
  const [_, __, url] = config.validateEnv();   
  const filePath = path.resolve(__dirname,'..','upload-files', '10MB-test-file');   
  const file = fs.createReadStream(filePath);

  return new Promise((resolve, reject) => {
    const options = {
      endpoint: `${url}/upload`,
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
      metadata: {
        filename: "10MB-test-file",
        filetype: "text/plain",
        meta_destination_id: "dextesting",
        meta_ext_event: "testevent1",
        meta_ext_source: "IZGW",
        meta_ext_sourceversion: "V2022-12-31",
        meta_ext_entity: "DD2",
        meta_username: "ygj6@cdc.gov",
        meta_ext_objectkey: uuidv4(),
        meta_ext_filename: "10MB-test-file",
        meta_ext_submissionperiod: '1',
      },
      onError: (error) => {
        console.error("An error occurred:", error);
        reject(error);
      },
      onSuccess: () => {
        console.log("Upload finished:", upload.url);

        // Extract uploadId from the upload.url
        extractUploadId(upload.url)
          .then(uploadId => resolve(uploadId))
          .catch(error => reject(error));
      },
    };

    const upload = new tus.Upload(file, options);
    upload.start();
  });
}

function extractUploadId(uploadUrl) {
    return new Promise((resolve, reject) => {
      const uploadIdPattern = /files\/([a-zA-Z0-9]+)/;
      const matches = uploadUrl.match(uploadIdPattern);
      if (matches && matches[1]) {
        console.log("Extracted uploadId:", matches[1]);
        resolve(matches[1]);
      } else {
        reject(new Error("Could not extract uploadId from upload URL."));
      }
    });
  }