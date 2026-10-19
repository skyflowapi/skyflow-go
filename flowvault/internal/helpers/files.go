package helpers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

const (
	putFailedStatus  = "PUT failed: %d"
	putFailedRequest = "PUT failed: %s"
	contentType      = "Content-Type"
)

// ResolvedFile is a file column's content plus the name and content type to upload it with.
type ResolvedFile struct {
	Name        string
	ContentType string
	Content     []byte
}

// ResolveFile reads the column's file from FilePath, Base64 or FileObject and works out
// its name and content type.
func ResolveFile(column fvcommon.UploadFilesRequestColumn) (ResolvedFile, *fverrors.SkyflowError) {
	file := ResolvedFile{Name: column.FileName}
	switch {
	case column.FilePath != "":
		content, err := os.ReadFile(column.FilePath)
		if err != nil {
			return file, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fmt.Sprintf(fverrors.FILE_NOT_READABLE, column.Column, err.Error()))
		}
		file.Content = content
		if file.Name == "" {
			file.Name = filepath.Base(column.FilePath)
		}
	case column.Base64 != "":
		content, err := base64.StdEncoding.DecodeString(column.Base64)
		if err != nil {
			return file, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fmt.Sprintf(fverrors.INVALID_BASE64, column.Column))
		}
		file.Content = content
	default:
		content, err := io.ReadAll(column.FileObject)
		if err != nil {
			return file, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fmt.Sprintf(fverrors.FILE_NOT_READABLE, column.Column, err.Error()))
		}
		file.Content = content
		if file.Name == "" && column.FileObject.Name() != "" {
			file.Name = filepath.Base(column.FileObject.Name())
		}
	}
	if file.Name != "" {
		file.ContentType = mime.TypeByExtension(filepath.Ext(file.Name))
	}
	return file, nil
}

// PutFile uploads one file to its signed URL. The signature is in the URL, so no
// Authorization header is sent.
func PutFile(ctx context.Context, client *http.Client, signedURL string, file ResolvedFile) (fvcommon.UploadStatus, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, bytes.NewReader(file.Content))
	if err != nil {
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedRequest, err.Error())
	}
	if file.ContentType != "" {
		req.Header.Set(contentType, file.ContentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedRequest, err.Error())
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedStatus, resp.StatusCode)
	}
	return fvcommon.UploadStatusUploaded, ""
}

// RunConcurrently runs tasks with at most limit running at once, and waits for all of them.
func RunConcurrently(tasks []func(), limit int) {
	var wg sync.WaitGroup
	slots := make(chan struct{}, limit)
	for _, task := range tasks {
		wg.Add(1)
		slots <- struct{}{}
		go func(run func()) {
			defer wg.Done()
			defer func() { <-slots }()
			run()
		}(task)
	}
	wg.Wait()
}
