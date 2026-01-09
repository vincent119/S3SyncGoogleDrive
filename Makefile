.PHONY: all build test cover clean

APP_NAME := S3SyncGoogleDrive

# Make 預設目標
all: build

# 編譯專案
build:
	@echo "正在編譯 $(APP_NAME)..."
	@go build -o $(APP_NAME) ./cmd

# 執行單元測試
test:
	@echo "正在執行測試..."
	@go test -v ./...

# 產生測試覆蓋率報告
cover:
	@echo "正在產生覆蓋率報告..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@echo "若要查看 HTML 詳細報告，請執行: go tool cover -html=coverage.out"

# 清理產出物
clean:
	@echo "正在清理..."
	@rm -f $(APP_NAME) coverage.out
