constraints:
- generate random data to stream for any given test
    - can be synthetic hl7 eventually
- any number of concurrent connections of different sizes
- very clear logging of what's happening where
- can handle auth

wanted
- highly configurable

# Usage:
```
# see usage info
go run main.go -h
```

Run against a local server in benchmark mode:
```
go run main.go
```