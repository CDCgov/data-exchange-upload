const fs = require("fs")
const tus = require("tus-js-client")
const path = require("path")

// sendFilesToAPI, Function to read files from a folder and send them to an tusd resumable API
async function sendFilesToAPI() {
  const folderPath = `./test-files.1`
  const apiUrl = `http://localhost:8080`

  try {
    // Get list of files in the folder
    const files = fs.readdirSync(folderPath)

    // Array to hold promises for each file upload
    const uploadPromises = files.map(async (file) => {
      const filePath = path.join(folderPath, file)
      // Create a readable stream for the file
      const fileStream = fs.createReadStream(filePath)

      const endpoint = `${apiUrl}/files/`
      console.log('endpoint', endpoint)

      const options = {
        endpoint,
        headers: {},
        chunkSize: 40000000,
        parallelUploads: 1,
        retryDelays: [0, 1000, 3000, 5000],
        metadata: {
          filename: file,
          meta_destination_id: "dextesting",
          meta_ext_event: "testevent1",
        },
        onError(error) {
          console.error("An error occurred:")
          console.error(error)
          process.exitCode = 1
        },
        onProgress(bytesUploaded, bytesTotal) {
          // const percentage = ((bytesUploaded / bytesTotal) * 100).toFixed(2);
          // console.log(bytesUploaded, bytesTotal, `${percentage}%`);
        },
        onSuccess() {
          console.log("Upload finished:", upload.url)
        },
      } // .options

      const upload = new tus.Upload(fileStream, options)
      upload.start()
    })

    // Wait for all file uploads to complete
    await Promise.all(uploadPromises)

    console.log("All files sent successfully.")
  } catch (error) {
    console.error("Error occurred:", error)
  }
}

sendFilesToAPI()
