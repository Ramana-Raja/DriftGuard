package storage

import (
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

func InitMinIO() {
	var err error

	Client, err = minio.New("minio:9100", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "supersecret", ""),
		Secure: false,
	})

	if err != nil {
		log.Fatal(err)
	}
}

func UploadFile(bucket, objectName string, file io.Reader, size int64) (string, error) {
	ctx := context.Background()

	_, err := Client.PutObject(ctx, bucket, objectName, file, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})

	if err != nil {
		return "", err
	}

	return objectName, nil
}
