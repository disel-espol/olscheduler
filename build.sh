cd ${0%/*}
export GOPATH=`pwd`/hack/go

# Build
# go build -v -o ./bin/olscheduler ./src/main.go

# Static build for docker image from busybox
CGO_ENABLED=1 go build -v -o ./bin/olscheduler -ldflags '-w -extldflags "-static"' ./src/main.go