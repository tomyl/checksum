
.PHONY: all generate install lint

build:
	go build

install:
	go install

all: generate lint install

generate:
	go generate ./ent

lint:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
