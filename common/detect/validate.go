package detect

import (
	"fmt"
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

// validateDeidentifyString checks Text. ConfigurationId and Configuration are optional and
// sent as given.
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

// validateDeidentifyFileSource checks that the file comes
// from exactly one of FilePath, FileObject and Value, each counted when set (Value only when
// non-empty: Go can't tell an empty string from an unset one), that Value has a DataSource
// and is not blank, and that a local file is a regular file, before the poll options are
// checked. The DataSource must be one of the vault type's de-identify sources (a pointer to
// one is accepted). It is ignored for a local file.
func (c *Controller) validateDeidentifyFileSource(request DeidentifyFileRequest) *skyflowError.SkyflowError {
	switch countSet(request.FilePath != "", request.FileObject != nil, request.Value != "") {
	case 0:
		return c.invalid(errMissingFileSource)
	case 1:
	default:
		return c.invalid(errMultipleFileSources)
	}
	if request.Value == "" {
		return c.checkLocalFile(request.FilePath, request.FileObject)
	}
	source := deidentifySourceValue(request.DataSource)
	if source == nil || source.deidentifySource() == "" {
		return c.invalid(errInvalidDeidentifyDataSource)
	}
	if !slices.Contains(c.sources.Deidentify, source) {
		return c.invalid(fmt.Sprintf(errUnsupportedDeidentifyDataSource,
			source.deidentifySource(), sourceList(c.sources.Deidentify, DeidentifyFileSource.deidentifySource)))
	}
	if blank(request.Value) {
		return c.invalid(errInvalidDeidentifyValue)
	}
	return nil
}

// validateReidentifyFile checks the file's DataSource and Value. The DataSource must be one
// of the vault type's re-identify sources (a pointer to one is accepted). Redaction rules are
// optional (nil rules are skipped).
func (c *Controller) validateReidentifyFile(request ReidentifyFileRequest) *skyflowError.SkyflowError {
	source := reidentifySourceValue(request.DataSource)
	if source == nil || source.reidentifySource() == "" {
		return c.invalid(errInvalidReidentifyDataSource)
	}
	if !slices.Contains(c.sources.Reidentify, source) {
		return c.invalid(fmt.Sprintf(errUnsupportedReidentifyDataSource,
			source.reidentifySource(), sourceList(c.sources.Reidentify, ReidentifyFileSource.reidentifySource)))
	}
	if blank(request.Value) {
		return c.invalid(errInvalidReidentifyValue)
	}
	return nil
}

// sourceList joins the wire values of a vault type's sources for an error message.
func sourceList[S any](sources []S, wire func(S) string) string {
	names := make([]string, len(sources))
	for i, source := range sources {
		names[i] = wire(source)
	}
	return strings.Join(names, ", ")
}
