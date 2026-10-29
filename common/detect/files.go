package detect

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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
func (c *Controller) readLocalFile(method, filePath string, fileObject *os.File) (*localFile, *skyflowError.SkyflowError) {
	if fileObject == nil {
		if blank(filePath) {
			return nil, nil
		}
		file, err := os.Open(filePath)
		if err != nil {
			return nil, c.invalid(errFileNotReadable, method, err.Error())
		}
		defer file.Close()
		fileObject = file
	}
	content, err := readRegularFile(fileObject)
	if err != nil {
		return nil, c.invalid(errFileNotReadable, method, err.Error())
	}
	name := filepath.Base(fileObject.Name())
	return &localFile{name: name, base64: base64.StdEncoding.EncodeToString(content), format: formatFromName(name)}, nil
}

// readRegularFile reads file, refusing anything but a regular file: a device or pipe such as
// /dev/zero has no end and would never finish reading.
func readRegularFile(file *os.File) ([]byte, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errNotRegularFile
	}
	return io.ReadAll(file)
}

var knownFormats = func() map[string]FileDataFormat {
	formats := map[string]FileDataFormat{}
	for _, f := range []FileDataFormat{
		FileDataFormatMP3, FileDataFormatWAV, FileDataFormatPDF, FileDataFormatTXT, FileDataFormatCSV,
		FileDataFormatJSON, FileDataFormatJPG, FileDataFormatJPEG, FileDataFormatTIF, FileDataFormatTIFF,
		FileDataFormatPNG, FileDataFormatBMP, FileDataFormatXLS, FileDataFormatXLSX, FileDataFormatDOC,
		FileDataFormatDOCX, FileDataFormatPPT, FileDataFormatPPTX, FileDataFormatXML, FileDataFormatDCM,
		FileDataFormatJSONL, FileDataFormatZIP, FileDataFormatGIF,
	} {
		formats[string(f)] = f
	}
	return formats
}()

// formatFromName infers the format from the file extension, or returns "" so the server
// infers it.
func formatFromName(name string) FileDataFormat {
	return knownFormats[strings.ToLower(strings.TrimPrefix(filepath.Ext(name), extensionSeparator))]
}

// writeOutputs decodes each base64 output into dir (the current directory when empty).
// Presigned-URL outputs are not fetched, so nothing is written for them.
//
// Names: processed-<input name>.<ext> for a processed file, and
// processed-entities-<input name>.json / processed-object-entities-<input name>.json for
// entity outputs. If two outputs would share a name, the output type goes before the
// extension.
func (c *Controller) writeOutputs(outputs []FileOutput, outputType DataSourceType, inputName, dir string) *skyflowError.SkyflowError {
	if len(outputs) == 0 {
		return nil
	}
	if outputType != DataSourceTypeBase64 {
		logger.Info(fmt.Sprintf(c.logMsg(logOutputNotWritten), outputType))
		return nil
	}
	if dir != "" {
		if err := os.MkdirAll(dir, outputDirMode); err != nil {
			return c.writeFailure(dir, err)
		}
	}
	stem := strings.TrimSuffix(inputName, filepath.Ext(inputName))
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
		if err := writeOutputFile(target, content); err != nil {
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
	extension := string(knownFormats[strings.ToLower(string(output.ProcessedFileExtension))])
	if extension == "" {
		extension = binaryExtension
		if entities {
			extension = string(FileDataFormatJSON)
		}
	}
	prefix := processedFilePrefix
	if entities {
		// ENTITIES -> processed-entities-<stem>, OBJECT_ENTITIES -> processed-object-entities-<stem>
		prefix += strings.ReplaceAll(fileType, fileTypeSeparator, nameSeparator) + nameSeparator
	}
	name := prefix + stem + extensionSeparator + extension
	if used[name] {
		suffix := fileType
		if suffix == "" {
			suffix = fmt.Sprint(len(used))
		}
		name = prefix + stem + extensionSeparator + suffix + extensionSeparator + extension
	}
	used[name] = true
	return name
}

// writeOutputFile writes content to a new owner-only file at target. An existing regular file is
// removed first, so its mode is not kept. Anything else at target, such as a symlink or a
// directory, is refused, and O_EXCL fails rather than follow a link created in the meantime.
func writeOutputFile(target string, content []byte) error {
	if info, err := os.Lstat(target); err == nil {
		if !info.Mode().IsRegular() {
			return errNotRegularFile
		}
		if err := os.Remove(target); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, outputFileMode)
	if err != nil {
		return err
	}
	_, err = file.Write(content)
	return errors.Join(err, file.Close())
}

func (c *Controller) writeFailure(path string, err error) *skyflowError.SkyflowError {
	message := fmt.Sprintf(c.errMsg(errOutputNotWritten), path, err.Error())
	logger.Error(message)
	return skyflowError.NewSkyflowError(skyflowError.SERVER, message)
}
