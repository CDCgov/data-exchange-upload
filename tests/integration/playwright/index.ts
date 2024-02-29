import fs from "fs";
import * as tus from "tus-js-client";
import { v4 as uuidv4 } from "uuid";

import axios from "axios";
import c, { LoginResponse } from "./client";
import config from "./config";



export async function start(): Promise<void> {
  const [username, password, url, ps_url] = config.validateEnv();
 
  const loginResponse: LoginResponse | null = await c.login(username, password, url);  
  
  if (loginResponse === null) {
    console.error("Login failed. Exiting program.");
    process.exit(1);
  }

  const path: string = `${__dirname}/../../upload-files/10MB-test-file`;
  const file = fs.createReadStream(path);
 

  const options: tus.UploadOptions = {
    endpoint: `${url}/upload`,
    headers: {
      Authorization: `Bearer ${loginResponse?.accessToken}`,
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
    onError(error) {
      console.error("An error occurred:");
      console.error(error);
      process.exitCode = 1;
    },
    onProgress(bytesUploaded, bytesTotal) {
      const percentage = ((bytesUploaded / bytesTotal) * 100).toFixed(2);
      console.log(bytesUploaded, bytesTotal, `${percentage}%`);
    },
    onSuccess() {
      console.log("Upload finished:", url);

      // Extract uploadId from the upload.url
      const uploadIdPattern = /files\/([a-zA-Z0-9]+)/;
      const matches = url.match(uploadIdPattern);
      if (matches && matches[1]) {
        const uploadId = matches[1];
        console.log("Extracted uploadId:", uploadId);        

        // Now make a GET request to the PS API with this uploadId
        getTraceResponse(uploadId, loginResponse.accessToken);
      } else {
        console.error("Could not extract uploadId from upload URL.");
      }
    },
  };

  const upload = new tus.Upload(file, options);
  upload.start();
}

export async function getTraceResponse(uploadId: string, accessToken: string): Promise<void> {
  const psApiUrl: string = `${config.validateEnv()[3]}/api/trace/uploadId/${uploadId}`;
  try {
    const response = await axios.get(psApiUrl, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });
    console.log("psApiUrl:", psApiUrl);
    console.log("Trace response:", response.data);
  } catch (error) {
    console.error("Failed to fetch trace response:", error);
  }
}

start();
