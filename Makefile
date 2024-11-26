VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.3")
NOTES ?= $(VERSION)
BINARY_NAME=foulbot
.PHONY: all tidy build

all: tidy build

tidy:
	go mod tidy

build:
	go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY_NAME) main.go
	chmod +x $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)*

tag:
	@echo "Creating tag $(VERSION)..."
	git tag $(VERSION)
	git push origin $(VERSION)

build-all:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY_NAME)-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY_NAME)-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui -X main.VERSION=$(VERSION)" -o $(BINARY_NAME)-windows-amd64.exe main.go

	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY_NAME)-linux-arm64 main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.VERSION=$(VERSION)" -o $(BINARY_NAME)-darwin-arm64 main.go
	GOOS=windows GOARCH=arm64 go build -ldflags "-H=windowsgui -X main.VERSION=$(VERSION)" -o $(BINARY_NAME)-windows-arm64.exe main.go

release:
	@echo "Creating release $(VERSION)..."
	@gh release create $(VERSION) \
		$(BINARY_NAME)-linux-amd64 \
		$(BINARY_NAME)-darwin-amd64 \
		$(BINARY_NAME)-windows-amd64.exe \
		$(BINARY_NAME)-linux-arm64 \
		$(BINARY_NAME)-darwin-arm64 \
		$(BINARY_NAME)-windows-arm64.exe \
		--title "Release $(VERSION)" \
		--notes "$(NOTES)" \
		--draft=false