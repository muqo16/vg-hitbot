.PHONY: build run clean test

build:
	go build -o vgbot.exe ./cmd/vgbot

run: build
	./vgbot.exe -domain example.com -pages 5 -duration 2 -hpm 10

clean:
	rm -f vgbot vgbot.exe

test:
	go test ./...
