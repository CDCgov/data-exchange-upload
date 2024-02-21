(async () => {

  console.log('ready to upload files')


  let upload          = null
  let uploadIsRunning = false

  const input           = document.querySelector('input[type=file]')
  const progress        = document.querySelector('.progress')
  const progressBar     = progress.querySelector('.bar')
  const alertBox        = document.querySelector('#support-alert')
  const uploadList      = document.querySelector('#upload-list')
  const chunkInput      = document.querySelector('#chunksize')
  const parallelInput   = document.querySelector('#paralleluploads')
  const endpointInput   = document.querySelector('#endpoint')
  const metadataInput   = document.querySelector('#metadata_json')
  metadataInput.value =  '{"meta_destination_id":"dextesting", "destination":"dextesting-testevent1"}'


  function reset (startTimeUpload, fileListBytesUploaded, fileListBytesTotal) {

    // ----------------------------------------------------
    if ( fileListBytesUploaded >= fileListBytesTotal) {

      console.log("files uploaded ok!")
      const durationUpload = (new Date().getTime()) - startTimeUpload
  
      console.log(`total upload duration [ms]: ${durationUpload}, [s]: ${durationUpload / 1000 }`)  

      input.value = ''

      uploadIsRunning = false

    } // .if

  } // .reset


  function startUpload() {
    const fileList = Array.from(input.files) //[0]
    // Only continue if a file has actually been selected.
    // IE will trigger a change event even if we reset the input element
    // using reset() and we do not want to blow up later.

    // console.log('fileList', fileList)
    if (!fileList ) { return }
    progressBar.style.width = 0

    const startTimeUpload = new Date().getTime()
    console.log('')
    console.log('starting time: ', startTimeUpload) 

    //
    // uploads set-up
    // ----------------------------------------------------
    const endpoint = endpointInput.value
    console.log('starting upload to endpoint: ', endpoint) 

    let chunkSize = parseInt(chunkInput.value, 10)
    if (Number.isNaN(chunkSize)) {
      chunkSize = Infinity
    } // .if

    let parallelUploads = parseInt(parallelInput.value, 10)
    if (Number.isNaN(parallelUploads)) {
      // currently this is disabled as only parallelUploads = 1 works
      // multiple uploads can be started concurrent ( new tus.Upload ), however each one sends chuncks serially to the server
      // tusd azure does not support chuncks concatenation, ref: https://github.com/tus/tusd/issues/843 
      parallelUploads = 1
    } // .if

    let metadataJSON = JSON.parse(metadataInput.value)
    let metadataJSONstr = JSON.stringify(metadataJSON, null, 4)
    console.log(`metadata JSON: ${metadataJSONstr}`)



    // toggleBtn.textContent = 'pause upload'

    const fileListBytesTotal = fileList.reduce( 
      (acc, curr ) => acc + curr.size,
      0
    ) // .fileListBytesTotal
    console.log('fileListBytesTotal: ', fileListBytesTotal)

    //
    // uploading files
    // ----------------------------------------------------
    let fileListBytesUploaded = 0

    fileList.forEach( (file, index) => {

      console.log(`start uploading file: ${file.name}, file index: ${index}`)

      let lastChunckNotAdded = true 
      let prevFileBytesUploaded = 0

      const metadata = {
        ...metadataJSON,
  
        // REQUIRED: original file name
        filename: file.name, 
      } // .metadata

      const options = {
        endpoint,
        headers: {
        },
        chunkSize,
        retryDelays: [0, 1000, 3000, 5000],
        parallelUploads,
        metadata: metadata,
        // metadata   : {
        //   filename: file.name,
        //   filetype: file.type,
        // },
        onError (error) {
          if (error.originalRequest) {
            if (window.confirm(`Failed because: ${error}\nDo you want to retry?`)) {
              upload.start()
              uploadIsRunning = true
              return
            }
          } else {
            window.alert(`Failed because: ${error}`)
          }

          reset(startTimeUpload, fileListBytesUploaded, fileListBytesTotal)
        },
        onProgress (fileBytesUploaded, fileBytesTotal) {


          if (fileBytesUploaded !== fileBytesTotal || lastChunckNotAdded ) {
            fileListBytesUploaded = fileListBytesUploaded + fileBytesUploaded - prevFileBytesUploaded
            prevFileBytesUploaded = fileBytesUploaded
            
          } else {
            lastChunckNotAdded = false 
          }

  
          const percentageFile = ((fileBytesUploaded / fileBytesTotal) * 100).toFixed(2)
          const percentageTotal = ((fileListBytesUploaded / fileListBytesTotal) * 100).toFixed(2)
  
          progressBar.style.width = `${percentageTotal}%`
  
          console.log('file:', file.name, fileBytesUploaded, fileBytesTotal, `${percentageFile}%`)
          console.log('fileList (total):', fileListBytesUploaded, fileListBytesTotal, `${percentageTotal}%`)
  
        },
        onSuccess () {
  
          console.log(`file: ${file.name} uploaded`)
  
          // console.log(`upload tus status url (needs bearer token): ${env.DEX_URL}/upload/status/${upload.url}`)
          // console.log(`upload status url (supplemental api, needs bearer token): ${env.DEX_URL}/status/${upload.url}`)
  
          const anchor = document.createElement('a')
          anchor.textContent = `Download ${upload.file.name} (${upload.file.size} bytes)`
          anchor.href = upload.url
          anchor.className = 'btn btn-success'
          uploadList.appendChild(anchor)

          reset(startTimeUpload, fileListBytesUploaded, fileListBytesTotal)
        },
      } // .options
      
      let upload = new tus.Upload(file, options)
      upload.start()
      uploadIsRunning = true


    }) // .fileList.forEach

  } // .startUpload

  if (!tus.isSupported) {
    alertBox.classList.remove('hidden')
  } // .if

  input.addEventListener('change', startUpload)

})() // async
