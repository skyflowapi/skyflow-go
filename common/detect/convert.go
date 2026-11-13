package detect

import (
	"slices"
	"strings"

	gen "github.com/skyflowapi/skyflow-go/common/internal/generated"
)

// Mapping between the public types and the generated wire types. The generated types use
// pointers for every optional field; an unset public value maps to nil so it is omitted.

func optString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func optFloat(v *float32) *float64 {
	if v == nil {
		return nil
	}
	f := float64(*v)
	return &f
}

// optEnum converts a public enum to the generated enum type of the same values.
func optEnum[G ~string, P ~string](v P) *G {
	if v == "" {
		return nil
	}
	g := G(v)
	return &g
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func derefEnum[G ~string](p *G) string {
	if p == nil {
		return ""
	}
	return string(*p)
}

// ---------------------------------------------------------------------------
// Requests
// ---------------------------------------------------------------------------

// toGenConfiguration adds the vault ID the API requires on an inline configuration. It returns
// nil without one: the request then names a stored configuration by ConfigurationId instead.
// requestConfiguration is the configuration a request sends. An inline
// configuration is always sent; with neither it nor a ConfigurationId, a configuration
// holding only the vault ID is sent and the server decides whether that is enough. With only
// a ConfigurationId, none is sent.
func requestConfiguration(configurationId string, c *DetectConfiguration, vaultId string) *gen.DetectConfigV2 {
	if c == nil && configurationId == "" {
		return &gen.DetectConfigV2{VaultID: vaultId}
	}
	return toGenConfiguration(c, vaultId)
}

func toGenConfiguration(c *DetectConfiguration, vaultId string) *gen.DetectConfigV2 {
	if c == nil {
		return nil
	}
	out := &gen.DetectConfigV2{
		Name:        optString(c.Name),
		Description: optString(c.Description),
		VaultID:     vaultId,
		Detect:      toGenDetect(c.Detect),
		Media:       toGenMedia(c.Media),
	}
	for _, m := range c.FileMapping {
		out.FileMapping = append(out.FileMapping, &gen.FileMapping{
			Source:                    optString(m.Source),
			Destination:               optString(m.Destination),
			OutputFileNameSuffix:      optString(m.OutputFileNameSuffix),
			EntitiesDestination:       optString(m.EntitiesDestination),
			ObjectEntitiesDestination: optString(m.ObjectEntitiesDestination),
		})
	}
	return out
}

func toGenDetect(s *Detect) *gen.Detect {
	if s == nil {
		return nil
	}
	out := &gen.Detect{
		Restrict:       s.Restrict,
		Skip:           s.Skip,
		ReturnEntities: optEnum[gen.DetectReturnEntities](s.ReturnEntities),
	}
	for _, e := range s.Entities {
		entity := &gen.DetectEntity{
			EntityType:           optEnum[gen.DetectEntityEntityType](e.EntityType),
			DeidentificationType: optEnum[gen.DetectEntityDeidentificationType](e.DeidentificationType),
			Destination:          optString(e.Destination),
			IncludePrefix:        optEnum[gen.DetectEntityIncludePrefix](e.IncludePrefix),
			PrefixOverride:       optString(e.PrefixOverride),
		}
		if e.Transformation != nil {
			entity.Transformation = &gen.EntityTransformation{}
			if shift := e.Transformation.ShiftDates; shift != nil {
				entity.Transformation.ShiftDates = &gen.TransformationShiftDates{MinDays: shift.MinDays, MaxDays: shift.MaxDays}
			}
		}
		out.Entities = append(out.Entities, entity)
	}
	for _, o := range s.ObjectEntities {
		out.ObjectEntities = append(out.ObjectEntities, &gen.DetectObjectEntity{
			EntityType:           optEnum[gen.DetectObjectEntityEntityType](o.EntityType),
			DeidentificationType: optEnum[gen.DetectObjectEntityDeidentificationType](o.DeidentificationType),
		})
	}
	for _, t := range s.CustomTypes {
		out.CustomTypes = append(out.CustomTypes, &gen.DetectCustomTypes{
			Label:                optString(t.Label),
			Match:                t.Match,
			DeidentificationType: optEnum[gen.DetectCustomTypesDeidentificationType](t.DeidentificationType),
			Destination:          optString(t.Destination),
			IncludePrefix:        optEnum[gen.DetectCustomTypesIncludePrefix](t.IncludePrefix),
			PrefixOverride:       optString(t.PrefixOverride),
		})
	}
	return out
}

func toGenMedia(m *Media) *gen.Media {
	if m == nil {
		return nil
	}
	out := &gen.Media{}
	if a := m.Audio; a != nil {
		out.Audio = &gen.AudioV2{
			OutputProcessedAudio: a.OutputProcessedAudio,
			OutputTranscription:  optEnum[gen.AudioV2OutputTranscription](a.OutputTranscription),
		}
		if b := a.Bleep; b != nil {
			out.Audio.Bleep = &gen.AudioV2Bleep{
				StartPadding: optFloat(b.StartPadding),
				StopPadding:  optFloat(b.StopPadding),
				Frequency:    b.Frequency,
				Gain:         b.Gain,
			}
		}
	}
	if d := m.Document; d != nil {
		out.Document = &gen.DocumentV2{}
		if p := d.Pdf; p != nil {
			out.Document.Pdf = &gen.DocumentV2Pdf{
				ProcessingMode: optEnum[gen.DocumentV2PdfProcessingMode](p.ProcessingMode),
				Density:        p.Density,
				MaxResolution:  p.MaxResolution,
			}
		}
	}
	if i := m.Image; i != nil {
		out.Image = &gen.ImageV2{
			OutputProcessedImage: i.OutputProcessedImage,
			OutputOcrText:        i.OutputOcrText,
			MaskingMethod:        optEnum[gen.ImageV2MaskingMethod](i.MaskingMethod),
		}
	}
	return out
}

// isNilRule reports whether rule is nil or a nil pointer to one of the rule types.
func isNilRule(rule RedactionRule) bool {
	switch r := rule.(type) {
	case nil:
		return true
	case *FlowRedactionLevel:
		return r == nil
	case *SkyRedactionLevel:
		return r == nil
	}
	return false
}

func (s FlowDeidentifySourceType) deidentifySource() string { return string(s) }
func (s SkyDeidentifySourceType) deidentifySource() string  { return string(s) }
func (s FlowReidentifySourceType) reidentifySource() string { return string(s) }
func (s SkyReidentifySourceType) reidentifySource() string  { return string(s) }

// deidentifySourceName returns source's wire value, or "" for nil, including a nil
// *FlowDeidentifySourceType or *SkyDeidentifySourceType.
func deidentifySourceName(source DeidentifyFileSource) string {
	switch s := source.(type) {
	case nil:
		return ""
	case *FlowDeidentifySourceType:
		if s == nil {
			return ""
		}
	case *SkyDeidentifySourceType:
		if s == nil {
			return ""
		}
	}
	return source.deidentifySource()
}

// reidentifySourceName returns source's wire value, or "" for nil, including a nil
// *FlowReidentifySourceType or *SkyReidentifySourceType.
func reidentifySourceName(source ReidentifyFileSource) string {
	switch s := source.(type) {
	case nil:
		return ""
	case *FlowReidentifySourceType:
		if s == nil {
			return ""
		}
	case *SkyReidentifySourceType:
		if s == nil {
			return ""
		}
	}
	return source.reidentifySource()
}

// resolveSource maps a wire value to the vault's source, whose wire value wire returns; one
// the vault does not know is nil.
func resolveSource[S comparable](value string, known []S, wire func(S) string) S {
	for _, source := range known {
		if wire(source) == value {
			return source
		}
	}
	var unknown S
	return unknown
}

func resolveDeidentifySource(value string, known []DeidentifyFileSource) DeidentifyFileSource {
	return resolveSource(value, known, DeidentifyFileSource.deidentifySource)
}

func resolveReidentifySource(value string, known []ReidentifyFileSource) ReidentifyFileSource {
	return resolveSource(value, known, ReidentifyFileSource.reidentifySource)
}

func (l FlowRedactionLevel) apiFields() (string, string, EntityType, RedactionType) {
	return l.TokenGroupName, l.RedactionPattern, "", ""
}

func (l SkyRedactionLevel) apiFields() (string, string, EntityType, RedactionType) {
	return "", "", l.EntityName, l.RedactionType
}

// toGenRedactionLevels sends each rule's fields as given. Nil rules, including nil
// *FlowRedactionLevel and *SkyRedactionLevel pointers, are skipped.
func toGenRedactionLevels(levels []RedactionRule) []*gen.RedactionLevel {
	var out []*gen.RedactionLevel
	for _, l := range levels {
		if isNilRule(l) {
			continue
		}
		tokenGroupName, redactionPattern, entityName, redactionType := l.apiFields()
		out = append(out, &gen.RedactionLevel{
			TokenGroupName:   optString(tokenGroupName),
			EntityName:       optEnum[gen.RedactionLevelEntityName](entityName),
			RedactionPattern: optString(redactionPattern),
			RedactionType:    optEnum[gen.RedactionLevelRedactionType](redactionType),
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// Responses
// ---------------------------------------------------------------------------

func fromGenDeidentifyString(r *gen.DeidentifyStringResponseV2) *DeidentifyStringResponse {
	out := &DeidentifyStringResponse{ProcessedText: deref(r.ProcessedText), Metrics: fromGenMetrics(r.Metrics), Entities: []DetectedEntity{}}
	for _, e := range r.Entities {
		if e == nil {
			continue
		}
		entity := DetectedEntity{Token: deref(e.Token), Value: deref(e.Value), EntityType: deref(e.EntityType), EntityScores: e.EntityScores}
		if entity.EntityScores == nil {
			entity.EntityScores = map[string]float64{}
		}
		if l := e.Location; l != nil {
			entity.Location = &EntityLocation{
				StartIndex: deref(l.StartIndex), EndIndex: deref(l.EndIndex),
				StartIndexProcessed: deref(l.StartIndexProcessed), EndIndexProcessed: deref(l.EndIndexProcessed),
			}
		}
		out.Entities = append(out.Entities, entity)
	}
	return out
}

// runStatuses are the run statuses this SDK knows.
var runStatuses = []DetectRunStatus{DetectRunStatusUnknown, DetectRunStatusFailed, DetectRunStatusSuccess, DetectRunStatusInProgress, DetectRunStatusQueued}

// guardrailsValidations are the guardrails results this SDK knows.
var guardrailsValidations = []GuardrailsValidation{GuardrailsValidationFailed, GuardrailsValidationPassed, GuardrailsValidationUnknown}

// knownOr returns value when it is one of known, and fallback otherwise (missing included).
func knownOr[T ~string](value string, known []T, fallback T) T {
	if slices.Contains(known, T(value)) {
		return T(value)
	}
	return fallback
}

// runStatus maps the API's run status; a missing or unknown one is UNKNOWN, which ends polling.
func runStatus(value string) DetectRunStatus {
	return knownOr(value, runStatuses, DetectRunStatusUnknown)
}

// fromGenRun fills in the run ID, which the API does not echo. A missing or unknown status
// is UNKNOWN; an output type that isn't one of sources (the SDK's accepted types) is empty.
func fromGenRun(runId string, r *gen.DetectRunsResponseV2, sources []DeidentifyFileSource) *GetRunResponse {
	return &GetRunResponse{
		RunId:      runId,
		Status:     runStatus(derefEnum(r.Status)),
		OutputType: resolveDeidentifySource(derefEnum(r.OutputType), sources),
		Output:     fromGenOutputs(r.Output),
		Metrics:    fromGenMetrics(r.Metrics),
		Message:    deref(r.Message),
	}
}

// fromGenOutputs never returns nil.
func fromGenOutputs(outputs []*gen.FileOutput) []FileOutput {
	out := []FileOutput{}
	for _, o := range outputs {
		if o == nil {
			continue
		}
		out = append(out, FileOutput{
			ProcessedFile:          deref(o.ProcessedFile),
			ProcessedFileType:      ProcessedFileType(derefEnum(o.ProcessedFileType)),
			ProcessedFileExtension: formatFromWire(derefEnum(o.ProcessedFileExtension)), // unknown: empty
		})
	}
	return out
}

// fromGenReidentifyOutputs never returns nil.
func fromGenReidentifyOutputs(outputs []*gen.FileOutput) []ReidentifyFileOutput {
	out := []ReidentifyFileOutput{}
	for _, o := range outputs {
		if o == nil {
			continue
		}
		out = append(out, ReidentifyFileOutput{
			ProcessedFile:          deref(o.ProcessedFile),
			ProcessedFileType:      ProcessedFileType(derefEnum(o.ProcessedFileType)),
			ProcessedFileExtension: reidentifyFormatFromWire(derefEnum(o.ProcessedFileExtension)), // unknown: empty
		})
	}
	return out
}

// reidentifyFormats maps each re-identify format's lower-case wire value (also its file
// extension) to its constant.
var reidentifyFormats = func() map[string]ReidentifyFileDataFormat {
	formats := map[string]ReidentifyFileDataFormat{}
	for _, f := range []ReidentifyFileDataFormat{
		ReidentifyFileDataFormatCSV, ReidentifyFileDataFormatDOCX, ReidentifyFileDataFormatJSON, ReidentifyFileDataFormatJSONL,
		ReidentifyFileDataFormatPPTX, ReidentifyFileDataFormatTXT, ReidentifyFileDataFormatXLSX, ReidentifyFileDataFormatXML,
	} {
		formats[strings.ToLower(string(f))] = f
	}
	return formats
}()

// reidentifyFormatFromWire returns the re-identify format for a wire value or extension in
// any case, or "" for one re-identification doesn't work with.
func reidentifyFormatFromWire(value string) ReidentifyFileDataFormat {
	return reidentifyFormats[strings.ToLower(value)]
}

func fromGenMetrics(m *gen.Metrics) *Metrics {
	if m == nil {
		return nil
	}
	return &Metrics{
		Size:           m.Size,
		WordCount:      m.WordCount,
		CharacterCount: m.CharacterCount,
		Slides:         m.Slides,
		Pages:          m.Pages,
		Duration:       m.Duration,
	}
}
