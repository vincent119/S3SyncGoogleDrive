#!/bin/bash

rm -rf go.mod go.sum
go mod init S3SyncGoogleDrive



go get github.com/aws/aws-sdk-go-v2@v1.55.0
go get github.com/aws/aws-sdk-go-v2/service/s3@v1.78.2
go get github.com/vbauerster/mpb/v8@v8.8.0


go mod tidy