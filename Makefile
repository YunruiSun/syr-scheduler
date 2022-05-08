all: local

local: 
	GOOS=linux GOARCH=amd64 go build  -o=bin/syr-scheduler ./cmd/scheduler

build:  local
	docker build --no-cache . -t github.com/YunruiSun/syr-scheduler:1.0

push:   build
	docker push github.com/yunruisun/syr-scheduler:1.0
