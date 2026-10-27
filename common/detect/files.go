package detect

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

// localFile is a local input read for sending.
type localFile struct {
	name   string // base name, used to name written outputs
	base64 string
	format FileDataFormat // inferred from the extension; empty if unknown
}

// readLocalFile reads FilePath or FileObject. It returns nil when neither is set.
func (c *Controller) readLocalFile(method, filePath string, fileObject *os.File) (*localFile, *skyflowError.SkyflowError) {
	var (
		name    string
		content []byte
		err     error
	)
	switch {
	case fileObject != nil:
		name = fileObject.Name()
		content, err = io.ReadAll(fileObject)
	case !blank(filePath):
		name = filePath
		content, err = os.ReadFile(filePath)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, c.invalid(errFileNotReadable, method, err.Error())
	}
	name = filepath.Base(name)
	return &localFile{name: name, base64: base64.StdEncoding.EncodeToString(content), format: formatFromName(name)}, nil
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
	return knownFormats[strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))]
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
		if err := os.MkdirAll(dir, 0o750); err != nil {
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
		if err := os.WriteFile(target, content, 0o600); err != nil {
			return c.writeFailure(target, err)
		}
		logger.Info(fmt.Sprintf(c.logMsg(logOutputFileWritten), target))
	}
	return nil
}

func outputFileName(stem string, output FileOutput, used map[string]bool) string {
	fileType := strings.ToLower(string(output.ProcessedFileType))
	entities := strings.Contains(fileType, "entities")
	extension := strings.ToLower(string(output.ProcessedFileExtension))
	if extension == "" || extension == string(FileDataFormatUnknown) {
		extension = "bin"
		if entities {
			extension = "json"
		}
	}
	prefix := processedFilePrefix
	if entities {
		// ENTITIES -> processed-entities-<stem>, OBJECT_ENTITIES -> processed-object-entities-<stem>
		prefix += strings.ReplaceAll(fileType, "_", "-") + "-"
	}
	name := prefix + stem + "." + extension
	if used[name] {
		suffix := fileType
		if suffix == "" {
			suffix = fmt.Sprint(len(used))
		}
		name = prefix + stem + "." + suffix + "." + extension
	}
	used[name] = true
	return name
}

func (c *Controller) writeFailure(path string, err error) *skyflowError.SkyflowError {
	message := fmt.Sprintf(c.errMsg(errOutputNotWritten), path, err.Error())
	logger.Error(message)
	return skyflowError.NewSkyflowError(skyflowError.SERVER, message)
}
