package detect

import (
	"os"
	"strings"
	"time"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
)

func blank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func countSet(values ...bool) int {
	n := 0
	for _, set := range values {
		if set {
			n++
		}
	}
	return n
}

func (c *Controller) validateDeidentifyString(request DeidentifyStringRequest) *skyflowError.SkyflowError {
	if blank(request.Text) {
		return c.invalid(errEmptyText, methodDeidentifyString)
	}
	return c.validateConfigurationSource(methodDeidentifyString, request.ConfigurationId, request.Configuration)
}

func (c *Controller) validateDeidentifyFile(request DeidentifyFileRequest, options DeidentifyFileOptions) *skyflowError.SkyflowError {
	if err := c.validateFileSource(methodDeidentifyFile, request.DataSource, request.Value, request.FilePath, request.FileObject); err != nil {
		return err
	}
	if err := c.validateConfigurationSource(methodDeidentifyFile, request.ConfigurationId, request.Configuration); err != nil {
		return err
	}
	if options.Poll == nil {
		if options.OutputDirectory != "" {
			return c.invalid(errOutputDirectoryNeedPoll)
		}
		return nil
	}
	if wait := options.Poll.WaitTime; wait < 0 || wait > maxWaitTimeSeconds*time.Second || (wait > 0 && wait < time.Second) {
		return c.invalid(errInvalidWaitTime)
	}
	if attempts := options.Poll.MaxAttempts; attempts < 0 || attempts > maxMaxAttempts {
		return c.invalid(errInvalidMaxAttempts)
	}
	return nil
}

func (c *Controller) validateConfigurationSource(method, configurationId string, configuration *DetectConfiguration) *skyflowError.SkyflowError {
	if countSet(!blank(configurationId), configuration != nil) != 1 {
		return c.invalid(errConfigurationSource, method)
	}
	return nil
}

// validateFileSource checks that the file comes from exactly one of Value, FilePath and
// FileObject, and that DataSource suits it.
func (c *Controller) validateFileSource(method string, dataSource DataSourceType, value, filePath string, fileObject *os.File) *skyflowError.SkyflowError {
	if countSet(!blank(value), !blank(filePath), fileObject != nil) != 1 {
		return c.invalid(errFileSource, method)
	}
	if !blank(value) {
		if dataSource == "" {
			return c.invalid(errMissingDataSource, method)
		}
		return nil
	}
	if dataSource != "" && dataSource != DataSourceTypeBase64 {
		return c.invalid(errLocalFileDataSource, method)
	}
	return nil
}

// validateRedactionLevels checks each level has exactly one source and one replacement.
// Whether they suit the vault type is checked by the server.
func (c *Controller) validateRedactionLevels(method string, levels []RedactionLevel) *skyflowError.SkyflowError {
	for i, level := range levels {
		if countSet(!blank(level.TokenGroupName), level.EntityName != "") != 1 {
			return c.invalid(errRedactionSource, method, i)
		}
		if countSet(!blank(level.RedactionPattern), level.RedactionType != "") != 1 {
			return c.invalid(errRedactionReplacement, method, i)
		}
	}
	return nil
}
