.PHONY: all tidy build

all: tidy build

tidy:
	go mod tidy

build:
	go build -o foulbot main.go
	chmod +x foulbot

run: build
	./foulbot