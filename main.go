package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	http.HandleFunc("POST /", uploadHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if r.Header.Get("Content-Type") != "application/json" {
		slog.Error("Invalid content type", slog.String("content-type", r.Header.Get("Content-Type")))
		http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})

	if err != nil {
		slog.Error("Failed to create AWS session", slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	uploader := s3manager.NewUploader(sess)
	bucket := os.Getenv("S3_BUCKET")
	prefix := "/"

	if p := os.Getenv("S3_KEY_PREFIX"); p != "" {
		prefix = p
	}

	ts := time.Now().Format("2006-01-02-15:04:05.000")
	id, err := generateRandomString(8)

	if err != nil {
		slog.Error("Failed generating random ID", slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	key := fmt.Sprintf("%s%s.%s.json", prefix, ts, id)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   io.NopCloser(r.Body),
	})

	if err != nil {
		slog.Error("Failed to upload to S3", slog.Any("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("Successfully uploaded", slog.String("bucket", bucket), slog.String("key", key))
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
