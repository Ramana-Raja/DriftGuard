package storage

import (
	"context"
	"log"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

func InitMinIO() {
	var err error

	Client, err = minio.New("localhost:9100", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "password", ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func UploadFile(bucket, objectName string, file multipart.File, size int64) (string, error) {
	ctx := context.Background()

	_, err := Client.PutObject(ctx, bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})

	if err != nil {
		return "", err
	}

	return objectName, nil
}
