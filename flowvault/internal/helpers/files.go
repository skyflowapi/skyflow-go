package helpers

import (
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
	"strings"
	"sync"
	"syscall"

	fvconsts "github.com/skyflowapi/skyflow-go/flowvault/internal/constants"
	fvcommon "github.com/skyflowapi/skyflow-go/flowvault/utils/common"
	fverrors "github.com/skyflowapi/skyflow-go/flowvault/utils/error"
)

const (
	putFailedStatus  = "PUT failed: %d"
	putFailedRequest = "PUT failed: %s"
	notRegularFile   = "not a regular file"
	base64LineBreak  = "base64 content contains a line break"
	contentType      = "Content-Type"
)

// ResolvedFile is a file column's source plus the name and content type to upload it with.
// Like Java, nothing is held in memory: a FilePath or FileObject is streamed when it is
// uploaded, and base64 content is checked up front but decoded only while it is uploaded.
type ResolvedFile struct {
	Name        string
	ContentType string
	Base64      string   // base64 content, decoded while it is uploaded
	DecodedSize int64    // length of the decoded base64 content
	Path        string   // FilePath, opened for the upload
	Object      *os.File // FileObject; the caller closes it
	// objectSize is the FileObject's size, or -1 when it is not a regular file. As in the Java
	// and Python SDKs, a regular file is uploaded whole, from its start, whatever its current
	// offset. It is read by position, not through the shared offset, so the offset is left
	// as it was and the same *os.File can back several columns that upload at the same time.
	objectSize int64
	sized      bool // objectSize was set by ResolveFile
}

// ResolveFile checks the column's file source and works out its name and content type. A
// FilePath must be a readable regular file and a FileObject must be open; neither is read.
func ResolveFile(column fvcommon.UploadFilesRequestColumn) (ResolvedFile, *fverrors.SkyflowError) {
	file := ResolvedFile{Name: column.FileName}
	switch {
	case column.FilePath != "":
		if err := checkReadable(column.FilePath); err != nil {
			return file, fverrors.NewSkyflowErrorWithCause(fverrors.INVALID_INPUT_CODE, fverrors.INVALID_FILE_PATH, err)
		}
		file.Path = column.FilePath
		if file.Name == "" {
			file.Name = filepath.Base(column.FilePath)
		}
	case column.Base64 != "":
		// Decoding into io.Discard checks the content without keeping a decoded copy. As in the
		// Java and Python SDKs, the content must be strict base64: line breaks, which Go's
		// decoder would skip, are refused too.
		size, err := io.Copy(io.Discard, base64Reader(column.Base64))
		if err == nil && strings.ContainsAny(column.Base64, "\r\n") {
			err = errors.New(base64LineBreak)
		}
		if err != nil {
			return file, fverrors.NewSkyflowErrorWithCause(fverrors.INVALID_INPUT_CODE, fverrors.INVALID_BASE64, err)
		}
		file.Base64, file.DecodedSize = column.Base64, size
	default:
		if _, err := column.FileObject.Stat(); err != nil {
			return file, fverrors.NewSkyflowErrorWithCause(fverrors.INVALID_INPUT_CODE, fverrors.INVALID_FILE_OBJECT, err)
		}
		file.Object = column.FileObject
		file.objectSize, file.sized = fileSize(column.FileObject), true
		if file.Name == "" && column.FileObject.Name() != "" {
			file.Name = filepath.Base(column.FileObject.Name())
		}
	}
	file.ContentType = UploadContentType(file.Name)
	return file, nil
}

// UploadContentType is the Content-Type an uploaded file is sent with, as in the Java SDK:
// the SDK's own type for the name's extension (case-insensitive), else Go's, else
// application/octet-stream.
func UploadContentType(name string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if ext == "" {
		return fvconsts.DEFAULT_UPLOAD_CONTENT_TYPE
	}
	if known, ok := fvconsts.UploadContentTypes[ext]; ok {
		return known
	}
	if guessed := mime.TypeByExtension("." + ext); guessed != "" {
		return guessed
	}
	return fvconsts.DEFAULT_UPLOAD_CONTENT_TYPE
}

// checkReadable reports why path can't be uploaded: it must be a regular file that opens.
func checkReadable(path string) error {
	f, _, err := openRegular(path)
	if err != nil {
		return err
	}
	return f.Close()
}

// openRegular opens path for reading and returns it with its size, or an error when it is not
// a regular file. It opens without blocking, so a FIFO or device, even one swapped in after an
// earlier check, is refused instead of hanging the call; the check is made on the handle that
// is then read.
func openRegular(path string) (*os.File, int64, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, 0, err
	}
	if info, err := f.Stat(); err == nil && info.Mode().IsRegular() {
		return f, info.Size(), nil
	}
	_ = f.Close()
	return nil, 0, errors.New(notRegularFile)
}

// body returns the file's content to send and its length, or -1 when the length isn't
// known. A FilePath is opened here and closed by the HTTP client; a FileObject is left open.
func (f ResolvedFile) body() (io.ReadCloser, int64, error) {
	switch {
	case f.Path != "":
		file, size, err := openRegular(f.Path)
		if err != nil {
			return nil, 0, err // not a nil *os.File wrapped in a non-nil io.ReadCloser
		}
		return file, size, nil
	case f.Object != nil:
		size := f.objectSize
		if !f.sized {
			size = fileSize(f.Object)
		}
		if size < 0 {
			// A pipe or similar can't be rewound, so it is read as a stream from where it is.
			return io.NopCloser(f.Object), -1, nil
		}
		return io.NopCloser(io.NewSectionReader(f.Object, 0, size)), size, nil
	}
	return io.NopCloser(base64Reader(f.Base64)), f.DecodedSize, nil
}

// base64Reader decodes standard base64. Go's decoder skips line breaks, so ResolveFile refuses
// content with any before it gets here.
func base64Reader(content string) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, strings.NewReader(content))
}

// fileSize is the size of a regular file, or -1 when it isn't one (a pipe, a device), whose
// length can't be told up front.
func fileSize(file *os.File) int64 {
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return -1
	}
	return info.Size()
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
	if limit < 1 {
		limit = 1
	}
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
