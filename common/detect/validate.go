package detect

import (
	"os"
	"slices"
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
	return nil
}

func (c *Controller) validateDeidentifyFile(request DeidentifyFileRequest, options DeidentifyFileOptions) *skyflowError.SkyflowError {
	if err := c.validateDeidentifyFileSource(request.DataSource, request.Value, request.FilePath, request.FileObject); err != nil {
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

// validateDeidentifyFileSource checks that the file comes from exactly one of Value, FilePath
// and FileObject, and that Value has a DataSource the SDK's vaults accept. DataSource is
// ignored for a local file.
func (c *Controller) validateDeidentifyFileSource(dataSource DeidentifySourceType, value, filePath string, fileObject *os.File) *skyflowError.SkyflowError {
	if countSet(!blank(value), !blank(filePath), fileObject != nil) != 1 {
		return c.invalid(errFileSource, methodDeidentifyFile)
	}
	if blank(value) {
		return nil
	}
	if dataSource == "" {
		return c.invalid(errMissingDeidentifySource, methodDeidentifyFile, c.deidentifySourceList())
	}
	if !c.acceptsDeidentifySource(dataSource) {
		return c.invalid(errUnsupportedDeidentifySource, methodDeidentifyFile, dataSource, c.deidentifySourceList())
	}
	return nil
}

// validateReidentifyFile checks the file's DataSource and Value, and the redaction levels.
func (c *Controller) validateReidentifyFile(request ReidentifyFileRequest) *skyflowError.SkyflowError {
	if request.DataSource == "" {
		return c.invalid(errMissingReidentifySource, methodReidentifyFile, sourceList(reidentifySources))
	}
	if !slices.Contains(reidentifySources, request.DataSource) {
		return c.invalid(errUnsupportedReidentifySource, methodReidentifyFile, request.DataSource, sourceList(reidentifySources))
	}
	if blank(request.Value) {
		return c.invalid(errEmptyValue, methodReidentifyFile)
	}
	return c.validateRedactionLevels(methodReidentifyFile, request.RedactionLevel)
}

// validateRedactionLevels checks each level has exactly one source and one replacement.
// Whether they suit the vault type is not checked here or by the server; see RedactionLevel.
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
