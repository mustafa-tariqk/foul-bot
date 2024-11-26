VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
.PHONY: all tidy build

all: tidy build

tidy:
	go mod tidy

build:
	go build -ldflags "-X main.VERSION=$(VERSION)" -o foulbot main.go
	chmod +x foulbot

run: build
	./foulbot

tag:
    @echo "Creating tag $(VERSION)..."
    git tag $(VERSION)
    git push origin $(VERSION)