package detect

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

// Outputs hold de-identified data, so they are readable by the owner only. An existing output
// file is replaced, so it takes this mode too; an existing directory keeps its mode.
const (
	outputDirMode  = 0o700
	outputFileMode = 0o600
)

// localFile is a local input read for sending.
type localFile struct {
	name   string // base name, used to name written outputs
	base64 string
	format FileDataFormat // inferred from the extension; empty if unknown
}

var errNotRegularFile = errors.New(notRegularFile)

// readLocalFile reads FileObject, or else FilePath. It returns nil when neither is set.
func (c *Controller) readLocalFile(filePath string, fileObject *os.File) (*localFile, *skyflowError.SkyflowError) {
	if fileObject == nil {
		if filePath == "" {
			return nil, nil
		}
		file, err := os.Open(filePath)
		if err != nil {
			return nil, c.unreadable(err)
		}
		defer file.Close()
		fileObject = file
	}
	content, err := readRegularFile(fileObject)
	if err != nil {
		return nil, c.unreadable(err)
	}
	name := filepath.Base(fileObject.Name())
	return &localFile{name: name, base64: base64.StdEncoding.EncodeToString(content), format: formatFromName(name)}, nil
}

// unreadable is the error for a local file that could not be read, with the Java and Python
// SDKs' texts: one that does not exist or is not a regular file was not found; any other
// failure (permissions, I/O) failed to encode.
func (c *Controller) unreadable(err error) *skyflowError.SkyflowError {
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, errNotRegularFile) {
		return c.invalid(errFileNotFound)
	}
	return c.invalid(errFailedToEncodeFile)
}

// checkLocalFile reports, before anything is read, whether the local input is a regular file,
// as the Java and Python SDKs check it during validation. A file that can't be read is
// reported later, when it is read (unreadable).
func (c *Controller) checkLocalFile(filePath string, fileObject *os.File) *skyflowError.SkyflowError {
	var info os.FileInfo
	var err error
	if fileObject != nil {
		info, err = fileObject.Stat()
	} else {
		info, err = os.Stat(filePath)
	}
	if err != nil || !info.Mode().IsRegular() {
		return c.invalid(errFileNotFound)
	}
	return nil
}

// readRegularFile reads the whole of file, refusing anything but a regular file: a device or
// pipe such as /dev/zero has no end and would never finish reading. It reads from the start
// whatever the file's offset, and leaves the offset unchanged, so a FileObject that was
// written or read before is still sent in full.
func readRegularFile(file *os.File) ([]byte, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errNotRegularFile
	}
	return io.ReadAll(io.NewSectionReader(file, 0, info.Size()))
}

// knownFormats maps each format's lower-case wire value (also the file extension) to the
// FileDataFormat constant.
var knownFormats = func() map[string]FileDataFormat {
	formats := map[string]FileDataFormat{}
	for _, f := range []FileDataFormat{
		FileDataFormatMP3, FileDataFormatWAV, FileDataFormatPDF, FileDataFormatTXT, FileDataFormatCSV,
		FileDataFormatJSON, FileDataFormatJPG, FileDataFormatJPEG, FileDataFormatTIF, FileDataFormatTIFF,
		FileDataFormatPNG, FileDataFormatBMP, FileDataFormatXLS, FileDataFormatXLSX, FileDataFormatDOC,
		FileDataFormatDOCX, FileDataFormatPPT, FileDataFormatPPTX, FileDataFormatXML, FileDataFormatDCM,
		FileDataFormatJSONL, FileDataFormatZIP, FileDataFormatGIF,
	} {
		formats[wireFormat(f)] = f
	}
	return formats
}()

// wireFormat is the API's lower-case value for f, or "" for an empty f.
func wireFormat(f FileDataFormat) string {
	return strings.ToLower(string(f))
}

// formatFromWire returns the known format for an extension or wire value in any case,
// or "" when it is not a known format.
func formatFromWire(value string) FileDataFormat {
	return knownFormats[strings.ToLower(value)]
}

// formatFromName infers the format from the file extension, or returns "" so the server
// infers it.
func formatFromName(name string) FileDataFormat {
	return formatFromWire(strings.TrimPrefix(filepath.Ext(name), extensionSeparator))
}

// writeOutputs decodes each base64 output into dir (the current directory when empty).
// Presigned-URL outputs are not fetched, so nothing is written for them.
//
// Names: processed-<input name>.<ext> for a processed file, and
// processed-entities-<input name>.json / processed-object-entities-<input name>.json for
// entity outputs. If two outputs would share a name, the output type (or, when it is
// unknown, a counter) goes before the extension, and a further clash adds -2, -3, and so on.
// Only DeidentifyFile writes outputs, so outputType is the run's DeidentifyFileSource.
func (c *Controller) writeOutputs(outputs []FileOutput, outputType DeidentifyFileSource, inputName, dir string) *skyflowError.SkyflowError {
	if len(outputs) == 0 {
		return nil
	}
	if name := deidentifySourceName(outputType); name != base64Value {
		logger.Info(fmt.Sprintf(c.logMsg(logOutputNotWritten), name))
		return nil
	}
	if blank(dir) {
		dir = "" // a blank directory is the current one, as in the Java SDK
	} else if err := os.MkdirAll(dir, outputDirMode); err != nil {
		return c.writeFailure(dir, err)
	}
	// As in the Java and Python SDKs, only an extension after the first character is dropped,
	// so a dotfile such as ".env" keeps its whole name.
	stem := inputName
	if dot := strings.LastIndex(inputName, extensionSeparator); dot > 0 {
		stem = inputName[:dot]
	}
	if stem == "" {
		stem = defaultOutputName
	}
	used := map[string]bool{}
	for _, output := range outputs {
		if output.ProcessedFile == "" {
			continue
		}
		target := filepath.Join(dir, outputFileName(stem, output, used))
		content, err := base64.StdEncoding.DecodeString(output.ProcessedFile)
		if err != nil {
			return c.writeFailure(target, err)
		}
		if err := writeOutputFile(target, content, os.Rename); err != nil {
			return c.writeFailure(target, err)
		}
		logger.Info(fmt.Sprintf(c.logMsg(logOutputFileWritten), target))
	}
	return nil
}

// outputFileName builds the name of one output. The extension and file type come from the
// server, so only known values are used: the name then never contains a path separator and
// cannot leave the output directory.
func outputFileName(stem string, output FileOutput, used map[string]bool) string {
	outputType := ProcessedFileType(strings.ToUpper(string(output.ProcessedFileType)))
	fileType := ""
	if knownFileTypes[outputType] {
		fileType = strings.ToLower(string(outputType))
	}
	entities := outputType == ProcessedFileTypeEntities || outputType == ProcessedFileTypeObjectEntities
	extension := wireFormat(formatFromWire(string(output.ProcessedFileExtension))) // "" when unknown
	if extension == "" {
		extension = binaryExtension
		if entities {
			extension = wireFormat(FileDataFormatJSON)
		}
	}
	prefix := processedFilePrefix
	if entities {
		// ENTITIES -> processed-entities-<stem>, OBJECT_ENTITIES -> processed-object-entities-<stem>
		prefix += strings.ReplaceAll(fileType, fileTypeSeparator, nameSeparator) + nameSeparator
	}
	// build gives <prefix><stem>[.<qualifier>].<extension>.
	build := func(qualifier string) string {
		if qualifier == "" {
			return prefix + stem + extensionSeparator + extension
		}
		return prefix + stem + extensionSeparator + qualifier + extensionSeparator + extension
	}
	name := build("")
	if used[name] {
		suffix := fileType
		if suffix == "" {
			suffix = fmt.Sprint(len(used))
		}
		name = build(suffix)
		for i := firstClashSuffix; used[name]; i++ {
			name = build(suffix + nameSeparator + fmt.Sprint(i))
		}
	}
	used[name] = true
	return name
}

// writeOutputFile writes content to an owner-only (outputFileMode) file at target. It writes
// a temporary file in the same directory and renames it over target with rename (os.Rename
// outside tests), so a failed write leaves any existing output in place and never a partial
// file at target. An existing regular file is replaced, so its mode is not kept. Anything else
// at target, such as a symlink or a directory, is refused; a link created in the meantime is
// replaced by the rename, not followed.
//
// If the process dies between creating and renaming the temporary file, a hidden
// ".processed-*.tmp" file (owner-only, holding de-identified data) is left in the directory;
// nothing removes it later.
func writeOutputFile(target string, content []byte, rename func(oldpath, newpath string) error) (err error) {
	if info, statErr := os.Lstat(target); statErr == nil && !info.Mode().IsRegular() {
		return errNotRegularFile
	}
	// CreateTemp opens a new file with O_EXCL; the mode is set explicitly below.
	temp, err := os.CreateTemp(filepath.Dir(target), tempFilePrefix+filepath.Base(target)+tempFileSuffix)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = os.Remove(temp.Name())
		}
	}()
	err = temp.Chmod(outputFileMode)
	if err == nil {
		_, err = temp.Write(content)
	}
	if err = errors.Join(err, temp.Close()); err == nil {
		err = rename(temp.Name(), target)
	}
	return err
}

// writeFailure is the error for an output file that could not be written. The failure is
// local, with no HTTP exchange behind it, so, as in the Java SDK, it has no HTTP status: the
// code is TRANSPORT_ERROR ("0") with an empty status. err is kept as the cause, so
// errors.Is(result, fs.ErrPermission) works.
func (c *Controller) writeFailure(path string, err error) *skyflowError.SkyflowError {
	logger.Error(fmt.Sprintf(c.logMsg(logOutputWriteFailed), path, err.Error()))
	message := fmt.Sprintf(c.errMsg(errOutputNotWritten), path)
	return c.base.Profile().NewErrorWithCause(skyflowError.TRANSPORT_ERROR, message, err)
}
