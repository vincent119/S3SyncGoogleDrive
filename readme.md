# S3SyncGoogleDrive

一個用於將 Amazon S3 檔案同步到 Google Drive 的 Go 應用程式。此工具可以自動下載指定 S3 前綴路徑下的所有檔案，並上傳到 Google Drive 指定資料夾中。

## 功能特色

- 🚀 **平行處理**: 支援多執行緒並行上傳，提高同步效率
- 📊 **進度顯示**: 即時顯示上傳進度條
- 🔄 **增量同步**: 基於 ETag 檢查，避免重複上傳相同檔案
- 🎯 **路徑映射**: 自動維護 S3 資料夾結構到 Google Drive
- ⚙️ **可配置**: 支援自定義並發數、AWS 和 Google Drive 設定

## 系統需求

- Go 1.24.0 或更高版本
- AWS S3 存取權限
- Google Drive API 存取權限

## 安裝與設定

### 1. 克隆專案

```bash
git clone https://github.com/vincent119/S3SyncGoogleDrive.git
cd S3SyncGoogleDrive
```

### 2. 安裝依賴

```bash
go mod download
```

### 3. 配置設定檔

複製範例配置檔案並編輯：

```bash
cp config/base_sample.yaml config/base.yaml
```

編輯 `config/base.yaml`，填入以下資訊：

```yaml
AWSConfig:
S3:
  bucketName: "your-s3-bucket-name"
  region: "ap-southeast-1"
  accessKeyId: "<your-aws-access-key-id>"
  secretAccess: "<your-aws-secret-access-key>"

Drive:
  client_id: "<your-google-client-id>.apps.googleusercontent.com"
  client_secret: "<your-google-client-secret>"
  refresh_token: "<your-google-refresh-token>"
  folder_id: "<your-google-drive-folder-id>"
  maxConcurrent: 10
```

### 4. Google Drive API 設定

1. 前往 [Google Cloud Console](https://console.cloud.google.com/)
2. 建立新專案或選擇現有專案
3. 啟用 Google Drive API
4. 建立 OAuth 2.0 憑證
5. 取得 refresh token（可參考 `config/refesh_token_eample.txt`）

## 使用方法

### 基本使用

```bash
go run ./cmd/main.go -p <s3-prefix-path>
```

### 範例

```bash
# 同步 S3 bucket 中 test999/ 路徑下的所有檔案
go run ./cmd/main.go -p test999

# 指定 Google Drive 根資料夾 ID
go run ./cmd/main.go -p test999 -droot <folder-id>

# 啟用除錯模式
go run ./cmd/main.go -p test999 -d
```

### 參數說明

- `-p`: **必要** S3 前綴路徑 (例如: test999)
- `-droot`: Google Drive 根資料夾 ID (預設: "root")
- `-d`: 啟用除錯日誌

## 編譯

### 本地編譯

```bash
go build -o S3SyncGoogleDrive ./cmd
```

### 跨平台編譯

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ./S3SyncGoogleDrive ./cmd

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o ./S3SyncGoogleDrive.exe ./cmd

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o ./S3SyncGoogleDrive ./cmd
```

## 專案結構

```text
S3SyncGoogleDrive/
├── cmd/
│   ├── main.go              # 主程式入口
│   └── s3sync/              # S3 同步相關命令
├── config/
│   ├── base_sample.yaml     # 配置檔案範例
│   └── refesh_token_eample.txt # Refresh Token 取得說明
├── internal/
│   ├── awsSDK/
│   │   ├── auth.go          # AWS 認證
│   │   └── S3/
│   │       └── S3.go        # S3 操作邏輯
│   ├── configs/
│   │   └── initConfig.go    # 配置初始化
│   ├── GoogleSDK/
│   │   ├── auth.go          # Google 認證
│   │   └── drive/
│   │       ├── progressBar.go # 進度條處理
│   │       └── upload.go    # 上傳邏輯
│   └── pkg/
│       └── progressReader/
│           └── progress.go  # 進度讀取器
├── go.mod
├── go.sum
├── init.sh                  # 初始化腳本
└── readme.md               # 專案說明文件
```

## 工作原理

1. **取得 S3 檔案列表**: 根據指定前綴路徑列出所有 S3 物件
2. **建立資料夾結構**: 在 Google Drive 中創建對應的資料夾結構
3. **檢查檔案存在性**: 使用 ETag 檢查檔案是否已存在於 Google Drive
4. **平行上傳**: 使用多執行緒並行處理檔案上傳
5. **進度追蹤**: 即時顯示每個檔案的上傳進度

## 注意事項

- 請確保 AWS 憑證有足夠的 S3 讀取權限
- Google Drive API 有每日請求限制，大量檔案同步時請注意
- 建議在初次使用時先用小量檔案測試
- ETag 比對機制可能在某些情況下不完全準確，建議定期檢查同步結果

## 疑難排解

### 常見錯誤

1. **AWS 認證失敗**

   - 檢查 `accessKeyId` 和 `secretAccess` 是否正確
   - 確認 AWS 憑證有 S3 讀取權限

2. **Google Drive API 錯誤**

   - 檢查 `client_id`、`client_secret` 和 `refresh_token` 是否正確
   - 確認 Google Drive API 已啟用

3. **檔案上傳失敗**
   - 檢查網路連線
   - 確認目標資料夾 ID 是否正確且有寫入權限

## 貢獻

歡迎提交 Issue 和 Pull Request 來改善此專案。

## 授權

此專案採用 MIT 授權條款。

---

## S3SyncGoogleDrive (English)

A Go application for synchronizing Amazon S3 files to Google Drive. This tool automatically downloads all files from a specified S3 prefix path and uploads them to a designated Google Drive folder.

## Features

- 🚀 **Parallel Processing**: Supports multi-threaded concurrent uploads for improved sync efficiency
- 📊 **Progress Display**: Real-time upload progress bars
- 🔄 **Incremental Sync**: ETag-based checking to avoid duplicate uploads of identical files
- 🎯 **Path Mapping**: Automatically maintains S3 folder structure in Google Drive
- ⚙️ **Configurable**: Supports custom concurrency settings, AWS and Google Drive configurations

## System Requirements

- Go 1.24.0 or higher
- AWS S3 access permissions
- Google Drive API access permissions

## Installation & Setup

### 1. Clone the Project

```bash
git clone https://github.com/vincent119/S3SyncGoogleDrive.git
cd S3SyncGoogleDrive
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Settings

Copy the sample configuration file and edit it:

```bash
cp config/base_sample.yaml config/base.yaml
```

Edit `config/base.yaml` with the following information:

```yaml
AWSConfig:
S3:
  bucketName: "your-s3-bucket-name"
  region: "ap-southeast-1"
  accessKeyId: "<your-aws-access-key-id>"
  secretAccess: "<your-aws-secret-access-key>"

Drive:
  client_id: "<your-google-client-id>.apps.googleusercontent.com"
  client_secret: "<your-google-client-secret>"
  refresh_token: "<your-google-refresh-token>"
  folder_id: "<your-google-drive-folder-id>"
  maxConcurrent: 10
```

### 4. Google Drive API Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing project
3. Enable Google Drive API
4. Create OAuth 2.0 credentials
5. Obtain refresh token (refer to `config/refesh_token_eample.txt`)

## Usage

### Basic Usage

```bash
go run ./cmd/main.go -p <s3-prefix-path>
```

### Examples

```bash
# Sync all files under test999/ path in S3 bucket
go run ./cmd/main.go -p test999

# Specify Google Drive root folder ID
go run ./cmd/main.go -p test999 -droot <folder-id>

# Enable debug mode
go run ./cmd/main.go -p test999 -d
```

### Parameter Description

- `-p`: **Required** S3 prefix path (e.g.: test999)
- `-droot`: Google Drive root folder ID (default: "root")
- `-d`: Enable debug logging

## Build

### Local Build

```bash
go build -o S3SyncGoogleDrive ./cmd
```

### Cross-Platform Build

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ./S3SyncGoogleDrive ./cmd

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o ./S3SyncGoogleDrive.exe ./cmd

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o ./S3SyncGoogleDrive ./cmd
```

## Project Structure

```text
S3SyncGoogleDrive/
├── cmd/
│   ├── main.go              # Main program entry point
│   └── s3sync/              # S3 sync related commands
├── config/
│   ├── base_sample.yaml     # Sample configuration file
│   └── refesh_token_eample.txt # Refresh Token acquisition guide
├── internal/
│   ├── awsSDK/
│   │   ├── auth.go          # AWS authentication
│   │   └── S3/
│   │       └── S3.go        # S3 operation logic
│   ├── configs/
│   │   └── initConfig.go    # Configuration initialization
│   ├── GoogleSDK/
│   │   ├── auth.go          # Google authentication
│   │   └── drive/
│   │       ├── progressBar.go # Progress bar handling
│   │       └── upload.go    # Upload logic
│   └── pkg/
│       └── progressReader/
│           └── progress.go  # Progress reader
├── go.mod
├── go.sum
├── init.sh                  # Initialization script
└── readme.md               # Project documentation
```

## How It Works

1. **Fetch S3 File List**: List all S3 objects based on the specified prefix path
2. **Create Folder Structure**: Create corresponding folder structure in Google Drive
3. **Check File Existence**: Use ETag to check if files already exist in Google Drive
4. **Parallel Upload**: Use multi-threading for concurrent file upload processing
5. **Progress Tracking**: Real-time display of upload progress for each file

## Notes

- Ensure AWS credentials have sufficient S3 read permissions
- Google Drive API has daily request limits, be mindful when syncing large numbers of files
- Recommended to test with a small number of files initially
- ETag comparison mechanism may not be completely accurate in some cases, recommend periodic verification of sync results

## Troubleshooting

### Common Errors

1. **AWS Authentication Failed**

   - Check if `accessKeyId` and `secretAccess` are correct
   - Verify AWS credentials have S3 read permissions

2. **Google Drive API Error**

   - Check if `client_id`, `client_secret`, and `refresh_token` are correct
   - Verify Google Drive API is enabled

3. **File Upload Failed**
   - Check network connection
   - Verify target folder ID is correct and has write permissions

## Contributing

Issues and Pull Requests are welcome to improve this project.

## License

This project is licensed under the MIT License.
