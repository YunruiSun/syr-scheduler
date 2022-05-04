all: local

local: fmt vet
	GOOS=linux GOARCH=amd64 go build  -o=bin/syr-scheduler ./cmd/scheduler

build:  local
	docker build --no-cache . -t github.com/YunruiSun/syr-scheduler:1.0

push:   build
	docker push github.com/YunruiSun/syr-scheduler:1.0

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

clean: fmt vet
	sudo rm -f syr-scheduler