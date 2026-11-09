package detect

import (
	"fmt"
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

// validateDeidentifyString checks Text. ConfigurationId and Configuration are optional and
// sent as given, as in the Java SDK.
func (c *Controller) validateDeidentifyString(request DeidentifyStringRequest) *skyflowError.SkyflowError {
	if blank(request.Text) {
		return c.invalid(errInvalidTextDeidentifyString)
	}
	return nil
}

func (c *Controller) validateDeidentifyFile(request DeidentifyFileRequest, options DeidentifyFileOptions) *skyflowError.SkyflowError {
	if err := c.validateDeidentifyFileSource(request); err != nil {
		return err
	}
	if options.Poll == nil {
		if !blank(options.OutputDirectory) {
			return c.invalid(errOutputDirectoryNeedPoll)
		}
		return nil
	}
	if wait := options.Poll.WaitTime; wait != nil && (*wait < minWaitTimeSeconds*time.Second || *wait > maxWaitTimeSeconds*time.Second) {
		return c.invalid(fmt.Sprintf(errInvalidWaitTime, minWaitTimeSeconds, maxWaitTimeSeconds))
	}
	if attempts := options.Poll.MaxAttempts; attempts != nil && (*attempts < minMaxAttempts || *attempts > maxMaxAttempts) {
		return c.invalid(fmt.Sprintf(errInvalidMaxAttempts, minMaxAttempts, maxMaxAttempts))
	}
	return nil
}

// validateDeidentifyFileSource checks, as the Java SDK does, that the file comes from exactly
// one of FilePath, FileObject and Value, each counted when set (Value even when blank), and
// that Value has a DataSource and is not blank. DataSource is ignored for a local file, and
// its value is sent as given: the server rejects one its vault does not accept. Whether a
// local file exists is checked when it is read.
func (c *Controller) validateDeidentifyFileSource(request DeidentifyFileRequest) *skyflowError.SkyflowError {
	switch countSet(request.FilePath != "", request.FileObject != nil, request.Value != "") {
	case 0:
		return c.invalid(errMissingFileSource)
	case 1:
	default:
		return c.invalid(errMultipleFileSources)
	}
	if request.Value == "" {
		return nil
	}
	if deidentifySourceName(request.DataSource) == "" {
		return c.invalid(errInvalidDeidentifyDataSource)
	}
	if blank(request.Value) {
		return c.invalid(errInvalidDeidentifyValue)
	}
	return nil
}

// validateReidentifyFile checks the file's DataSource and Value, as the Java SDK does. The
// DataSource is sent as given, and redaction rules are optional (nil rules are skipped).
func (c *Controller) validateReidentifyFile(request ReidentifyFileRequest) *skyflowError.SkyflowError {
	if reidentifySourceName(request.DataSource) == "" {
		return c.invalid(errInvalidReidentifyDataSource)
	}
	if blank(request.Value) {
		return c.invalid(errInvalidReidentifyValue)
	}
	return nil
}
