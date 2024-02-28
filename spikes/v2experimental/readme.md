# V2 Experiment Go

- Using TUSD official Go implementation as a library and building a custom Go server
- Refactor current python hooks check in Go
- Refactor current c# function into the Go custom server



# Why

- native with TUSD official implementation
- cross cloud using docker similar TUSD cli
- simpler implementation vs. current 
- potentially cost saving for life cycle of product, including mininal dependecies
- simpler error handling and reports
- can be free of event queue limitations


# Timeline
By end of current quarter, March13/14, deployed POC in DEV with similar existing functionality as current Upload API, including hooks, and copy

# Steps 

### Prod. Level:
- Application layout structure: [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- Server including router: `type server struct`, `server.go`, `router.go` 
- Server custom configuration: [tusd cli flags](https://github.com/tus/tusd/blob/main/cmd/tusd/cli/serve.go)
- Load environment variables from file [gadotenv](https://github.com/joho/godotenv)
  - Configuration [envconfig](https://github.com/sethvargo/go-envconfig)
- Logs [slog](https://go.dev/blog/slog)
- Errors [errors package](https://pkg.go.dev/errors), [working with errors](https://go.dev/blog/go1.13-errors)
- Tests (unit) [add a test](https://go.dev/doc/tutorial/add-a-test)
- Multi cloud, based on config, azure store or aws s3 [tusd cli](https://github.com/tus/tusd/tree/main/cmd/tusd/cli)
- Docker image [tusd docker](https://github.com/tus/tusd/blob/main/Dockerfile)
- Deploy to azure and/or k8s
- Go concurrency [goroutines](https://go.dev/tour/concurrency/1)
- Go [context](https://pkg.go.dev/context)
- Go [commnad line flags](https://pkg.go.dev/flag)

### Specifics [from main branch]:
- Run flags, decide if/what is needed with flags vs. static. [tusd cli flags](https://github.com/tus/tusd/blob/main/cmd/tusd/cli/serve.go)
- Based on metadata v1 or v2 or both?
- Read hooks configuration from the files, current v1 files available
- Check sender manifest using pre-hooks
- Configure the other hooks, (post?) if needed
- Define error package, inline with processing status API requirements
- Integration with processing status based on errors
- Health endpoint, checks storage connections
- Version endpoint, using git actions to populate 
- Configuration for azure store containers:Raw, DEX, EDAV.
- Write storage config object.
- Write storage connection, at least one retry. error should be at connection.
- Write copy function (functional style), 1 function copy A -> B, if feasible wrap in interface to work multi-cloud, function takes the config objects. retry? 
- write some unit tests
- docker file
- publish to image hub (github actions?)
- deploy - where?