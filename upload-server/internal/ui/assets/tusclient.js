// ------------------------------------------
// Variables
// ------------------------------------------

let upload = null;
let previousUpload = null;
let uploadIsRunning = false;
let file = null;

const fileInput = document.querySelector("input[type=file]");
const pauseButton = document.querySelector("#pause-upload-button");
const resumeButton = document.querySelector("#resume-upload-button");

const progressContainer = document.querySelector(".progress");
const progressBar = progressContainer.querySelector(".bar");

// Values also set in upload-server/pkg/info/info.go
// these values must match
const UPLOAD_INITIATED = "Initiated";
const UPLOAD_IN_PROGRESS = "In Progress";
const UPLOAD_COMPLETE = "Complete";

const UPLOAD_STATUS_LABEL_INITIALIZED = " Upload Initialized At: ";
const UPLOAD_STATUS_LABEL_IN_PROGRESS = " Last Chunk Received At: ";
const UPLOAD_STATUS_LABEL_COMPLETE = " Upload Completed At: ";
const UPLOAD_STATUS_LABEL_DEFAULT = " Uploaded At: ";

// ------------------------------------------
// Functions
// ------------------------------------------

// Hides or shows an element
function _toggleVisibility(element, show) {
  if (show) {
    element.classList.remove("hidden");
  } else {
    element.classList.add("hidden");
  }
}

// Hides or shows the progress bar
function _toggleProgressBar(show) {
  const uploaderContainer = document.querySelector(".uploader-container");
  _toggleVisibility(uploaderContainer, show);
  _toggleVisibility(progressContainer, show);
  _toggleVisibility(progressBar, show);

  const pauseResumeContainer = document.querySelector(
    ".pause-resume-upload-container"
  );
  _toggleVisibility(pauseResumeContainer, show);
  if (show) {
    _togglePauseButton(show);
  } else {
    _toggleVisibility(pauseButton, false);
    _toggleVisibility(resumeButton, false);
  }
}

function _togglePauseButton(pause) {
  if (pause) {
    _toggleVisibility(pauseButton, true);
    _toggleVisibility(resumeButton, false);
  } else {
    _toggleVisibility(pauseButton, false);
    _toggleVisibility(resumeButton, true);
  }
}

// Hides or shows the upload forms and the progress bar
function _toggleUploadContainer(show) {
  const uploadContainer = document.querySelector(".upload-container");
  _toggleVisibility(uploadContainer, show);
}

// Hides or shows the full upload form
function _toggleNewUploadForm(show) {
  const matches = document.querySelectorAll(".new-upload");
  for (const element of matches) {
    _toggleVisibility(element, show);
  }
}

// Hides or shows the resume upload form
function _toggleResumeUploadForm(show) {
  const matches = document.querySelectorAll(".resume-upload");
  for (const element of matches) {
    _toggleVisibility(element, show);
  }
}

function _toggleFormContainer(show) {
  const formContainer = document.querySelector(".form-container");
  _toggleVisibility(formContainer, show);
}

// Hides or shows the file info
function _toggleInfoContainer(show) {
  const infoContainer = document.querySelector(".file-detail-container");
  _toggleVisibility(infoContainer, show);
}

function _toggleNewUploadButtonContainer(show) {
  const buttonContainer = document.querySelector(
    ".new-upload-button-container"
  );
  _toggleVisibility(buttonContainer, show);
}

// Sets the view up as the initial upload form
function _showInitiatedUploadForm() {
  _toggleUploadContainer(true);
  _toggleNewUploadForm(true);
  _toggleResumeUploadForm(false);
  _toggleProgressBar(false);
  _toggleInfoContainer(false);
  _toggleNewUploadButtonContainer(false);
}

// Set the view up as the resume upload form
function _showResumableUploadForm() {
  _toggleUploadContainer(true);
  _toggleNewUploadForm(false);
  _toggleResumeUploadForm(true);
  _toggleProgressBar(false);
  _toggleInfoContainer(true);
  _toggleNewUploadButtonContainer(false);
}

// Sets the view up to only show the file info
function _showReadOnlyFileInfo(statusLabel) {
  _toggleUploadContainer(false);
  _toggleInfoContainer(true);
  _toggleNewUploadButtonContainer(true);
  _setUploadLastChunkReceivedLabel(statusLabel);
}

function _updateUploadStatusInProgress() {
  const statusValue = document.querySelector("#upload-status-value");
  statusValue.innerHTML = "In Progress";
  statusValue.classList.remove("upload-initiated");
  statusValue.classList.remove("upload-complete");
  statusValue.classList.add("upload-in-progress");
  _setUploadLastChunkReceivedLabel(UPLOAD_STATUS_LABEL_IN_PROGRESS);
}

function _setUploadLastChunkReceivedLabel(text) {
  document.querySelector("#upload-datetime-label").innerHTML = text;
}

function _updateLastChunkReceived() {
  const currTime = new Date();
  document.querySelector("#upload-datetime-value").innerHTML =
    currTime.toUTCString();
}

function _refreshPage() {
  // Adding a 1 second wait to make the refresh look a little smoother
  // when the file uploads really quickly
  setTimeout(() => {
    location.reload();
  }, 1000);
}

// Triggered by a file being selected. Gets the file from
// the file input. Gets the other values from the form, if
// this is a new upload, or from the previous upload metadata,
// if this is a resumed upload. Then sends those values to the
// uploadFile function to start the upload.
async function submitUploadForm() {
  const fileList = Array.from(fileInput.files); //[0]
  // Only continue if a file has actually been selected.
  // IE will trigger a change event even if we reset the input element
  // using reset() and we do not want to blow up later.

  if (!fileList || fileList.length < 1) {
    return;
  }

  // Retrieving the first file because this is only handling uploading
  // one file at a time
  file = fileList[0];

  let chunkSize;
  let parallelUploads;
  if (previousUpload) {
    // If there is a previous upload in local storage
    // retrieve the previously entered values for that upload
    const { metadata } = previousUpload;
    if (!metadata) {
      return;
    }

    // check the file to make sure it matches the previous file
    const { filename, fileType, fileSize, fileLastModified } = metadata;
    if (
      file.name != filename ||
      file.type != fileType ||
      fileSize != file.size ||
      fileLastModified != file.lastModified
    ) {
      fileInput.value = "";
      // if it doesn't match the expected file, alert the user that they should try again
      window.alert(
        `This file does not match the previously partially uploaded file.\nPlease try with another file.`
      );
      return;
    }
    ({ chunkSize, parallelUploads } = metadata);
  } else {
    const chunkInput = document.querySelector("#chunksize");
    chunkSize = parseInt(chunkInput.value, 10);
    if (Number.isNaN(chunkSize)) {
      chunkSize = Infinity;
    }

    // currently this is disabled as only parallelUploads = 1 works
    // multiple uploads can be started concurrent ( new tus.Upload ), however each one sends chunks serially to the server
    // tusd azure does not support chunks concatenation, ref: https://github.com/tus/tusd/issues/843
    parallelUploads = 1;
  }

  // hide the forms
  _toggleFormContainer(false);

  // Upload the file
  await uploadFile(file, { chunkSize, parallelUploads });
}

// Creates the tus client and uploads the file.
// Handles onProgress, onSuccess, and onError.
// Will resume an upload if one has already been started.
async function uploadFile(file, { chunkSize, parallelUploads }) {
  console.log(`start uploading file: ${file.name}`);

  // used to determine the duration
  const startTimeUpload = new Date().getTime();
  // because uploadLengthDeferred is true, tus will not know how large
  // the file is until the upload is complete, if we want to show the
  // upload progress we need to set the value outside of tus
  const fileSize = file.size;

  console.log("starting upload to endpoint: ", endpoint);
  console.log(getCookie("token"))
  const options = {
    headers: {
      "Tus-Resumable": "1.0.0",
      "Content-Type": "application/offset+octet-stream",
      "Authorization": "Bearer " + getCookie("token")
    },
    metadata: {
      filename: file.name,
      fileType: file.type,
      fileSize: file.size,
      fileLastModified: file.lastModified,
      endpoint,
      uploadUrl,
      chunkSize,
      parallelUploads,
    },
    protocol: "ietf-draft-03",
    uploadLengthDeferred: true,
    retryDelays: [0, 1000, 3000, 5000],
    removeFingerprintOnSuccess: true,
    endpoint,
    uploadUrl,
    chunkSize,
    parallelUploads,
    onError(error) {
      if (error.originalRequest) {
        // if the upload failed but is recoverable, ask if the user wants to retry
        if (window.confirm(`Failed because: ${error}\nDo you want to retry?`)) {
          upload.start();
          uploadIsRunning = true;
          return;
        }
      } else {
        // if upload is not recoverable, just show the error
        window.alert(`Failed because: ${error}`);
      }

      console.log(`file: ${file.name} upload failed: ${error}`);

      // Set running flag to false and reset the file input
      uploadIsRunning = false;
      fileInput.value = "";

      // display the file info
      _toggleInfoContainer(true);
    },
    onProgress(bytesUploaded, bytesTotal) {
      // because uploadLengthDeferred is true, the bytesTotal should always
      // be 0, if it is 0, use the fileSize set above
      const total = bytesTotal || fileSize;

      // the percent of the file that has been uploaded
      const percentageTotal = ((bytesUploaded / total) * 100).toFixed(2);
      const percentageTotalRound = Math.round(percentageTotal);

      // update the tool bar to show these new values
      // only show the progress bar while there is progress
      if (percentageTotal >= 0 && percentageTotal <= 100) {
        _toggleProgressBar(true);
        progressBar.style.width = `${percentageTotal}%`;
        progressBar.textContent = `${percentageTotalRound}%`;
      } else {
        _toggleProgressBar(false);
        progressBar.style.width = 0;
      }

      console.log(
        `uploadedBytes: ${bytesUploaded}, totalBytes: ${total}, percentComplete: ${percentageTotal}`
      );

      _updateLastChunkReceived();
    },
    onSuccess() {
      console.log(`file: ${file.name} uploaded successfully`);

      // get the duration of the upload
      const durationUpload = new Date().getTime() - startTimeUpload;

      console.log(
        `total upload duration [ms]: ${durationUpload}, [s]: ${durationUpload / 1000
        }`
      );

      // reset all of the values no longer needed
      upload = null;
      previousUpload = null;
      uploadIsRunning = false;
      file = null;

      fileInput.value = "";

      // refresh the page for the complete upload info
      _refreshPage();
    },
  };

  // create the tus client
  upload = new tus.Upload(file, options);

  if (previousUpload) {
    console.log(`Resuming the upload of ${file.name}`);
    // if there is a previous upload for this file
    // set it here so that it will resume
    upload.resumeFromPreviousUpload(previousUpload);
  }

  // Start the upload
  uploadIsRunning = true;
  upload.start();

  _updateUploadStatusInProgress();
  _toggleInfoContainer(true);
}

function resumeUpload() {
  if (!upload) {
    console.log(`No upload client, please try uploading again`);
    return;
  }

  // Check if there are any previous uploads to continue.
  upload.findPreviousUploads().then(function (previousUploads) {
    // Found previous uploads so we select the first one.
    if (previousUploads.length) {
      previousUpload = previousUploads[0];
      upload.resumeFromPreviousUpload(previousUploads[0]);
    }

    // Start the upload
    upload.start();
    _togglePauseButton(true);
  });
}

function pauseUpload() {
  if (!upload) {
    console.log(`No upload client, please try uploading again`);
    return;
  }

  upload.abort();
  _togglePauseButton(false);
}

// checks the local storage to see if there is a previous upload
// that was not completed and has the same uploadUrl
async function findResumableUpload() {
  let uploads = await tus.defaultOptions.urlStorage.findAllUploads();

  for (const upload of uploads) {
    if (upload.uploadUrl == uploadUrl) {
      return upload;
    }
  }

  return null;
}

function getCookie(cname) {
  let name = cname + "=";
  let ca = document.cookie.split(';');
  for(let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) === ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) === 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

(async () => {
  // initializes what is hidden/shown on the page base on
  // the uploadStatus and upload data in the local storage
  let isHost = false;

  switch (uploadStatus) {
    case UPLOAD_INITIATED:
      isHost = true;
      _showInitiatedUploadForm();
      break;
    case UPLOAD_IN_PROGRESS:
      previousUpload = await findResumableUpload();
      if (previousUpload) {
        isHost = true;
        _showResumableUploadForm();
      } else {
        _showReadOnlyFileInfo(UPLOAD_STATUS_LABEL_IN_PROGRESS);
      }
      break;
    case UPLOAD_COMPLETE:
      _showReadOnlyFileInfo(UPLOAD_STATUS_LABEL_COMPLETE);
      break;
    default:
      console.error(`${uploadStatus} is an invalid status`);
      _showReadOnlyFileInfo(UPLOAD_STATUS_LABEL_DEFAULT);
  }

  if (isHost) {
    // we only need to use tus and add the listener if we are the host
    // otherwise the forms will not be displayed
    if (!tus.isSupported) {
      document.querySelector("#support-alert").classList.remove("hidden");
    } // .if

    fileInput.addEventListener("change", () => submitUploadForm());
    pauseButton.addEventListener("click", () => pauseUpload());
    resumeButton.addEventListener("click", () => resumeUpload());
  }
})();
