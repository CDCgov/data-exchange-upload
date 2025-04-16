module github.com/cdcgov/bad-uploader

go 1.22.1
toolchain go1.24.1

replace github.com/eventials/go-tus => github.com/whytheplatypus/go-tus v0.0.0-20240709121510-b5e0bef51f72

require (
	github.com/eventials/go-tus v0.0.0-20220610120217-05d0564bb571
	github.com/hasura/go-graphql-client v0.12.2
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
)

require (
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	nhooyr.io/websocket v1.8.11 // indirect
)
