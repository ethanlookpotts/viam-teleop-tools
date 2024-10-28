build:
    rm -rf dist | true # macOS has surprising behavior with executables that change in-place
    mkdir -p dist
    go build -o dist/module ./module

setup:
    go mod tidy

fmt:
    @just format

format:
    gofmt -w .
