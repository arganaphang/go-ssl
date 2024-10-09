set dotenv-load

# Format Golang
format:
  gofumpt -l -w .
  goimports-reviser -rm-unused -set-alias ./...
  golines -w -m 120 **/*.go
  betteralign -apply ./...

# build -> build application
build:
  go build -o main ./main.go

# run -> application
run:
  ./main

# dev -> run build then run it
dev:
  watchexec -r -c -e go -- just build run

test:
  gotestdox ./...

# health -> Hit Health Check Endpoint
health:
  curl -s http://localhost:8000/healthz | jq

cert:
  ./gen-cert.sh
  docker compose -f "docker-compose.yaml" down
  docker compose -f "docker-compose.yaml" up -d --build