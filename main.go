package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "embed"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var accessKeyIDFlag = flag.String("access-key-id", "", "The minio access key ID.")
var secretAccessKeyFileFlag = flag.String("secret-access-key-file", "", "The file to load the minio secret access key from.")
var bucketNameFlag = flag.String("bucket", "", "The name of the minio bucket to use.")
var endpointFlag = flag.String("endpoint", "", "The minio endpoint to use.")
var fileToUploadFlag = flag.String("file", "", "The file to upload.")

func main() {
	loggingLevel := new(slog.LevelVar)
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: loggingLevel, AddSource: true}))
	slog.SetDefault(log)
	loggingLevel.Set(slog.LevelDebug)

	flag.Parse()
	var missingFlags []string
	if *accessKeyIDFlag == "" {
		missingFlags = append(missingFlags, "access-key-id")
	}
	if *secretAccessKeyFileFlag == "" {
		missingFlags = append(missingFlags, "secret-access-key-file")
	}
	if *bucketNameFlag == "" {
		missingFlags = append(missingFlags, "bucket")
	}
	if *endpointFlag == "" {
		missingFlags = append(missingFlags, "endpoint")
	}
	if *fileToUploadFlag == "" {
		missingFlags = append(missingFlags, "file")
	}
	if len(missingFlags) > 0 {
		log.Error("missing required flags", slog.Any("flags", missingFlags))
		os.Exit(1)
	}

	// Create minio client.
	t := http.DefaultTransport
	log.Info("loading minio secret access key from file", slog.String("file", *secretAccessKeyFileFlag))
	keyBytes, err := os.ReadFile(*secretAccessKeyFileFlag)
	if err != nil {
		log.Error("failed to load minio secret access key file", slog.Any("error", err))
		os.Exit(1)
	}
	minioSecretAccessKey := string(keyBytes)
	mc, err := minio.New(*endpointFlag, &minio.Options{
		Creds:     credentials.NewStaticV4(*accessKeyIDFlag, minioSecretAccessKey, ""),
		Secure:    true,
		Transport: t,
	})
	if err != nil {
		log.Error("failed to create minio client", slog.Any("error", err))
		os.Exit(1)
	}

	// Create presigned URL.
	pp, err := mc.PresignedPutObject(context.Background(), *bucketNameFlag, *fileToUploadFlag, time.Hour)
	if err != nil {
		log.Error("failed to create presigned URL", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("presigned URL created", slog.String("url", pp.String()))
}
