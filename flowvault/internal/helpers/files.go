package helpers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

const (
	putFailedStatus  = "PUT failed: %d"
	putFailedRequest = "PUT failed: %s"
	notRegularFile   = "not a regular file"
	contentType      = "Content-Type"
)

// ResolvedFile is a file column's source plus the name and content type to upload it with.
// Like Java, a FilePath or FileObject is streamed when it is uploaded, not held in memory;
// base64 content is decoded up front.
type ResolvedFile struct {
	Name        string
	ContentType string
	Content     []byte   // decoded base64
	Path        string   // FilePath, opened for the upload
	Object      *os.File // FileObject, read from its current offset; the caller closes it
}

// ResolveFile checks the column's file source and works out its name and content type. A
// FilePath must be a readable regular file and a FileObject must be open; neither is read.
func ResolveFile(column fvcommon.UploadFilesRequestColumn) (ResolvedFile, *fverrors.SkyflowError) {
	file := ResolvedFile{Name: column.FileName}
	switch {
	case column.FilePath != "":
		if err := checkReadable(column.FilePath); err != nil {
			return file, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fmt.Sprintf(fverrors.FILE_NOT_READABLE, column.Column, err.Error()))
		}
		file.Path = column.FilePath
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
		if _, err := column.FileObject.Stat(); err != nil {
			return file, fverrors.NewSkyflowError(fverrors.INVALID_INPUT_CODE, fmt.Sprintf(fverrors.FILE_NOT_READABLE, column.Column, err.Error()))
		}
		file.Object = column.FileObject
		if file.Name == "" && column.FileObject.Name() != "" {
			file.Name = filepath.Base(column.FileObject.Name())
		}
	}
	if file.Name != "" {
		file.ContentType = mime.TypeByExtension(filepath.Ext(file.Name))
	}
	return file, nil
}

// checkReadable reports why path can't be uploaded: it must be a regular file that opens.
func checkReadable(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New(notRegularFile)
	}
	return nil
}

// body returns the file's content to send and its length, or -1 when the length isn't
// known. A FilePath is opened here and closed by the HTTP client; a FileObject is left open.
func (f ResolvedFile) body() (io.ReadCloser, int64, error) {
	switch {
	case f.Path != "":
		file, err := os.Open(f.Path)
		if err != nil {
			return nil, 0, err
		}
		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return nil, 0, err
		}
		return file, info.Size(), nil
	case f.Object != nil:
		return io.NopCloser(f.Object), remaining(f.Object), nil
	}
	return io.NopCloser(bytes.NewReader(f.Content)), int64(len(f.Content)), nil
}

// remaining is the length from the file's current offset to its end, or -1 when it can't
// be told, e.g. for a pipe.
func remaining(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return -1
	}
	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return -1
	}
	return info.Size() - offset
}

// PutFile uploads one file to its signed URL. The signature is in the URL, so no
// Authorization header is sent.
func PutFile(ctx context.Context, client *http.Client, signedURL string, file ResolvedFile) (fvcommon.UploadStatus, string) {
	content, length, err := file.body()
	if err != nil {
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedRequest, err.Error())
	}
	if length == 0 {
		_ = content.Close()
		content = http.NoBody
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, signedURL, content)
	if err != nil {
		_ = content.Close()
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedRequest, withoutURL(err).Error())
	}
	req.ContentLength = length
	if file.ContentType != "" {
		req.Header.Set(contentType, file.ContentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedRequest, withoutURL(err).Error())
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fvcommon.UploadStatusFailed, fmt.Sprintf(putFailedStatus, resp.StatusCode)
	}
	return fvcommon.UploadStatusUploaded, ""
}

// withoutURL drops the request URL that http.Client.Do adds to its errors, because a signed
// URL is a write credential and the error is returned to the caller.
func withoutURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err
	}
	return err
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
