cd ${0%/*}
export GOPATH=`pwd`/hack/go
go build -o ./bin/olscheduler ./src/main.go