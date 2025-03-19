(async () => {
  console.log("ready to upload files");

  let upload = null;
  let uploadIsRunning = false;

  const input = document.querySelector("input[type=file]");
  const progressContainer = document.querySelector(".progress");
  const progressBar = progressContainer.querySelector(".bar");
  const alertBox = document.querySelector("#support-alert");
  const uploadList = document.querySelector("#upload-list");
  const chunkInput = document.querySelector("#chunksize");
  const parallelInput = document.querySelector("#paralleluploads");
  const endpointInput = document.querySelector("#endpoint");

  function reset(startTimeUpload, fileListBytesUploaded, fileListBytesTotal) {
    // ----------------------------------------------------
    if (fileListBytesUploaded >= fileListBytesTotal) {
      console.log("files uploaded ok!");
      const durationUpload = new Date().getTime() - startTimeUpload;

      console.log(
        `total upload duration [ms]: ${durationUpload}, [s]: ${
          durationUpload / 1000
        }`
      );

      input.value = "";

      uploadIsRunning = false;
    } // .if
  } // .reset

  function startUpload() {
    const fileList = Array.from(input.files); //[0]
    // Only continue if a file has actually been selected.
    // IE will trigger a change event even if we reset the input element
    // using reset() and we do not want to blow up later.

    // console.log('fileList', fileList)
    if (!fileList) {
      return;
    }
    progressBar.style.width = 0;

    const startTimeUpload = new Date().getTime();
    console.log("");
    console.log("starting time: ", startTimeUpload);

    //
    // uploads set-up
    // ----------------------------------------------------
    const endpoint = endpointInput.value;
    console.log("starting upload to endpoint: ", endpoint);

    let chunkSize = parseInt(chunkInput.value, 10);
    if (Number.isNaN(chunkSize)) {
      chunkSize = Infinity;
    } // .if

    let parallelUploads = parseInt(parallelInput.value, 10);
    if (Number.isNaN(parallelUploads)) {
      // currently this is disabled as only parallelUploads = 1 works
      // multiple uploads can be started concurrent ( new tus.Upload ), however each one sends chuncks serially to the server
      // tusd azure does not support chuncks concatenation, ref: https://github.com/tus/tusd/issues/843
      parallelUploads = 1;
    } // .if
    // toggleBtn.textContent = 'pause upload'

    const fileListBytesTotal = fileList.reduce(
      (acc, curr) => acc + curr.size,
      0
    ); // .fileListBytesTotal
    console.log("fileListBytesTotal: ", fileListBytesTotal);

    //
    // uploading files
    // ----------------------------------------------------
    let fileListBytesUploaded = 0;

    fileList.forEach((file, index) => {
      console.log(`start uploading file: ${file.name}, file index: ${index}`);

      let lastChunckNotAdded = true;
      let prevFileBytesUploaded = 0;
      console.log(uploadUrl);
      const options = {
        endpoint,
        headers: {
          "Tus-Resumable": "1.0.0",
        },
        protocol: "ietf-draft-03",
        uploadUrl,
        uploadSize: file.size,
        uploadLengthDeferred: true,
        chunkSize,
        retryDelays: [0, 1000, 3000, 5000],
        parallelUploads,
        onError(error) {
          if (error.originalRequest) {
            if (
              window.confirm(`Failed because: ${error}\nDo you want to retry?`)
            ) {
              upload.start();
              uploadIsRunning = true;
              return;
            }
          } else {
            window.alert(`Failed because: ${error}`);
          }

          reset(startTimeUpload, fileListBytesUploaded, fileListBytesTotal);
        },
        onProgress(fileBytesUploaded, fileBytesTotal) {
          if (fileBytesUploaded !== fileBytesTotal || lastChunckNotAdded) {
            fileListBytesUploaded =
              fileListBytesUploaded + fileBytesUploaded - prevFileBytesUploaded;
            prevFileBytesUploaded = fileBytesUploaded;
          } else {
            lastChunckNotAdded = false;
          }

          const percentageFile = (
            (fileBytesUploaded / fileBytesTotal) *
            100
          ).toFixed(2);
          const percentageTotal = (
            (fileListBytesUploaded / fileListBytesTotal) *
            100
          ).toFixed(2);
          const percentageTotalRound = Math.round(percentageTotal);

          if (percentageTotal >= 100) {
            progressContainer.classList.add("hidden");
            progressBar.style.width = 0;
          } else {
            progressContainer.classList.remove("hidden");
            progressBar.classList.remove("hidden");
            progressBar.style.width = `${percentageTotal}%`;
            progressBar.textContent = `${percentageTotalRound}%`;
          }

          console.log(
            "file:",
            file.name,
            fileBytesUploaded,
            fileBytesTotal,
            `${percentageFile}%`
          );
          console.log(
            "fileList (total):",
            fileListBytesUploaded,
            fileListBytesTotal,
            `${percentageTotal}%`
          );
        },
        onSuccess() {
          console.log(`file: ${file.name} uploaded`);

          // console.log(`upload tus status url (needs bearer token): ${env.DEX_URL}/upload/status/${upload.url}`)
          // console.log(`upload status url (supplemental api, needs bearer token): ${env.DEX_URL}/status/${upload.url}`)

          const anchor = document.createElement("a");
          anchor.textContent = `Download ${upload.file.name} (${upload.file.size} bytes)`;
          anchor.href = upload.url;
          anchor.className = "btn btn-success";
          uploadList.appendChild(anchor);

          reset(startTimeUpload, fileListBytesUploaded, fileListBytesTotal);
        },
      }; // .options

      let upload = new tus.Upload(file, options);
      upload.start();
      uploadIsRunning = true;
    }); // .fileList.forEach
  } // .startUpload

  if (!tus.isSupported) {
    alertBox.classList.remove("hidden");
  } // .if

  input.addEventListener("change", startUpload);
})(); // async
