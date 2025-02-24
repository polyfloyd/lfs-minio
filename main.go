package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"lfs-minio/lfs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	lfsEvents, lfsRespond := lfs.Begin()

	_ = (<-lfsEvents).(*lfs.Init)
	minioClient, minioBucket, err := minioInit()
	if err != nil {
		lfsRespond(lfs.InitErr(err))
		return
	}
	srv := &transferService{
		minioClient: minioClient,
		minioBucket: minioBucket,
		lfsRespond:  lfsRespond,
	}
	lfsRespond(lfs.InitOK())

	for event := range lfsEvents {
		switch t := event.(type) {
		case *lfs.Upload:
			srv.upload(t)
		case *lfs.Download:
			srv.download(t)
		case *lfs.Terminate:
			return
		}
	}
}

func minioInit() (*minio.Client, string, error) {
	endpoint, err := env("MINIO_ENDPOINT")
	if err != nil {
		return nil, "", err
	}
	bucket, err := env("MINIO_BUCKET")
	if err != nil {
		return nil, "", err
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewEnvMinio(),
		Secure: os.Getenv("MINIO_SECURE") != "0",
	})
	if err != nil {
		return nil, "", err
	}
	return client, bucket, nil
}

func env(k string) (string, error) {
	v, ok := os.LookupEnv(k)
	if !ok {
		return "", fmt.Errorf("missing %s env", k)
	}
	return v, nil
}

type transferService struct {
	minioClient *minio.Client
	minioBucket string
	lfsRespond  func(lfs.Response)
}

func (srv *transferService) upload(event *lfs.Upload) {
	ctx := context.Background()

	ifile, err := os.Open(event.Path)
	if err != nil {
		srv.lfsRespond(lfs.TransferError(event.OID, err))
		return
	}
	defer ifile.Close()

	r := lfs.ProgressReader(ifile, srv.lfsRespond, event.OID, event.Size)

	if _, err := srv.minioClient.PutObject(ctx, srv.minioBucket, event.OID, r, event.Size, minio.PutObjectOptions{}); err != nil {
		srv.lfsRespond(lfs.TransferError(event.OID, err))
		return
	}

	srv.lfsRespond(lfs.UploadComplete(event.OID))
}

func (srv *transferService) download(event *lfs.Download) {
	ctx := context.Background()

	ofile, err := os.CreateTemp(".", ".lfsdl-*") // TODO: Clean up leaked file in case of error
	if err != nil {
		srv.lfsRespond(lfs.TransferError(event.OID, err))
		return
	}
	defer ofile.Close()

	obj, err := srv.minioClient.GetObject(ctx, srv.minioBucket, event.OID, minio.GetObjectOptions{})
	if err != nil {
		srv.lfsRespond(lfs.TransferError(event.OID, err))
		return
	}

	r := lfs.ProgressReader(obj, srv.lfsRespond, event.OID, event.Size)

	if _, err := io.Copy(ofile, r); err != nil {
		srv.lfsRespond(lfs.TransferError(event.OID, err))
		return
	}

	srv.lfsRespond(lfs.DownloadComplete(event.OID, ofile.Name()))
}
