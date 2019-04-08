go build -v -o ./bin/olscheduler ./main.go

# Static build for docker image from busybox
# CGO_ENABLED=1 go build -v -o ./bin/olscheduler -ldflags '-w -extldflags "-static"' ./src/main.go
