# DEX Upload API File Delivery Retry Tool

This is a tool for the maintainers of the DEX Upload API service to help retry the delivery of files that have been uploaded successfully, but were unsuccessful in reaching their final delivery destination targets.

## Usage

This tool is run by building and running the Golang program, and providing the following command line arguments:

- `url` - The FQDN of the upload server on which to retry the files.
- `csvFiles` - A comma-separated list of relative paths to one or more CSV files containing a list of upload IDs and their corresponding delivery targets. These CSV files must have two columns, where the first column is the ID and the second is the target name.
- `parallelism` - An optional argument for performing the retry in parallel. When omitted, it uses the max number of CPUs on the machine.
- `v` - An optional argument for increasing the logging output.

Run the tool with `go run ./main.go -url=https://ocio-ede-prd-tusd-app-service.azurewebsites.net -csvFiles=input/file1.csv,input/file2.csv`.