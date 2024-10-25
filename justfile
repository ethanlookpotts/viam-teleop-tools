build:
	mkdir -p dist
	go build -o dist/module module/main.go

setup:
    go mod tidy