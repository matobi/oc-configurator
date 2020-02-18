NOW = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -extldflags -static -s -w

build:
	mkdir -p ./buildtarget
	env CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o ./buildtarget/oc-configurator ./cmd/oc-configurator

clean:
	rm -rf ./buildtarget/*

docker: clean build
