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
- Server including router
- Configuration
- Logs
- Tests (unit)
- Application structure
- Multi cloud, based on cofig, azure store or aws s3
- Docker image
- Deploy to azure