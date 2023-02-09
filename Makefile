.PHONY: build

build:
	go build -ldflags="-w -s -X github.com/trybefore/linksbot/internal/config.Commit=$$(git rev-parse --short HEAD)" -o ./linksbot .