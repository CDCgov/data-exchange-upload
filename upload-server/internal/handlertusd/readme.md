# Benchmarking

$ go test -bench=.
go test -bench=. -cpuprofile=cpu.prof
go test -bench=. -memprofile=mem.prof
go tool pprof -http=:8080 cpu.prof
go tool pprof -http=:8080 mem.prof


$ go test -run='^$' -bench='.' -cpuprofile='cpu.prof' -memprofile='mem.prof'
$ go test -run='^$' -bench='.' -cpuprofile='cpu.prof' -memprofile='mem.prof'

# PPROF

go install github.com/google/pprof@latest

go tool pprof cpu.prof
go tool pprof mem.prof

top

top 20

sort=cum

# GraphViz

(optional) https://graphviz.org/download/

$ go get github.com/goccy/go-graphviz