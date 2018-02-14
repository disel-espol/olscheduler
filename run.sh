cd ${0%/*}
export GOPATH=`pwd`/hack/go
go run ./src/main.go start -c ./config/olscheduler.json