#!/bin/bash

set -e

function check_dependencies() {
    command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
    command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
}

function build_docker() {
    echo "Building Docker image..."
    docker build -t a0-telegram-bot .
}

function run_docker() {
    echo "Running Docker container..."
    docker-compose up -d
}

function stop_docker() {
    echo "Stopping Docker container..."
    docker-compose down
}

CMD=$1

check_dependencies

case $CMD in
    "build")
        echo "Building project..."
        go build -o bin/server main.go
        ;;
    "run")
        echo "Running server..."
        go run main.go
        ;;
    "test")
        echo "Running tests..."
        go test -v -race -cover ./...
        ;;
    "clean")
        echo "Cleaning build artifacts..."
        rm -rf bin/
        ;;
    "fmt")
        echo "Formatting code..."
        go fmt ./...
        ;;
    "docker-build")
        build_docker
        ;;
    "docker-run")
        run_docker
        ;;
    "docker-stop")
        stop_docker
        ;;
    "docker-restart")
        stop_docker
        run_docker
        ;;
    *)
        echo "Usage: $0 {build|run|test|clean|fmt|docker-build|docker-run|docker-stop|docker-restart}"
        exit 1
        ;;
esac