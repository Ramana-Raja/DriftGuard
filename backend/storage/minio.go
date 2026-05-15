package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var client *azblob.Client

func InitBlobStorage() {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_KEY")

	if accountName == "" || accountKey == "" {
		log.Fatal("Fatal: Missing Azure Storage Account or Access Key environment parameters!")
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatalf("Fatal: Invalid Azure Storage Credentials: %v", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)

	client, err = azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		log.Fatalf("Fatal: Failed to construct Azure Blob Client engine: %v", err)
	}

	log.Println("Azure Blob Storage initialized successfully")
}

func UploadFile(containerName, objectName string, file io.Reader, fileSize int64) (string, error) {
	if client == nil {
		return "", fmt.Errorf("storage client is not initialized")
	}

	log.Printf("Uploading %s to Azure container %s...", objectName, containerName)

	_, err := client.UploadStream(context.TODO(), containerName, objectName, file, nil)
	if err != nil {
		log.Printf("Azure Upload Failed: %v", err)
		return "", err
	}

	storageAccount := os.Getenv("AZURE_STORAGE_ACCOUNT")
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", storageAccount, containerName, objectName)

	log.Printf("Azure Upload Success! Asset URL path: %s", blobURL)
	return blobURL, nil
}
