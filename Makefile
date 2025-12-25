
TARGET=webrtc-transcriber

.PHONY: default build-frontend build-backend

default: build-frontend build-backend

build-frontend:
	cd frontend && npm install && npm run build

build-backend:
	go build -o $(TARGET) ./cmd/transcribe-server/main.go