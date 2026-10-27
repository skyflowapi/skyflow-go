package detect

import gen "github.com/skyflowapi/skyflow-go/common/internal/generated"

// Mapping between the public types and the generated wire types. The generated types use
// pointers for every optional field; an unset public value maps to nil so it is omitted.

func optString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func optBool(v bool) *bool {
	if !v {
		return nil
	}
	return &v
}

func optInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

func optFloat(v float32) *float64 {
	if v == 0 {
		return nil
	}
	f := float64(v)
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

// toGenConfiguration adds the vault ID the API requires on an inline configuration.
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
				minDays, maxDays := shift.MinDays, shift.MaxDays
				entity.Transformation.ShiftDates = &gen.TransformationShiftDates{MinDays: &minDays, MaxDays: &maxDays}
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
			OutputProcessedAudio: optBool(a.OutputProcessedAudio),
			OutputTranscription:  optEnum[gen.AudioV2OutputTranscription](a.OutputTranscription),
		}
		if b := a.Bleep; b != nil {
			out.Audio.Bleep = &gen.AudioV2Bleep{
				StartPadding: optFloat(b.StartPadding),
				StopPadding:  optFloat(b.StopPadding),
				Frequency:    optInt(b.Frequency),
				Gain:         optInt(b.Gain),
			}
		}
	}
	if d := m.Document; d != nil {
		out.Document = &gen.DocumentV2{}
		if p := d.Pdf; p != nil {
			out.Document.Pdf = &gen.DocumentV2Pdf{
				ProcessingMode: optEnum[gen.DocumentV2PdfProcessingMode](p.ProcessingMode),
				Density:        optInt(p.Density),
				MaxResolution:  optInt(p.MaxResolution),
			}
		}
	}
	if i := m.Image; i != nil {
		out.Image = &gen.ImageV2{
			OutputProcessedImage: optBool(i.OutputProcessedImage),
			OutputOcrText:        optBool(i.OutputOcrText),
			MaskingMethod:        optEnum[gen.ImageV2MaskingMethod](i.MaskingMethod),
		}
	}
	return out
}

func toGenRedactionLevels(levels []RedactionLevel) []*gen.RedactionLevel {
	var out []*gen.RedactionLevel
	for _, l := range levels {
		out = append(out, &gen.RedactionLevel{
			TokenGroupName:   optString(l.TokenGroupName),
			EntityName:       optEnum[gen.RedactionLevelEntityName](l.EntityName),
			RedactionPattern: optString(l.RedactionPattern),
			RedactionType:    optEnum[gen.RedactionLevelRedactionType](l.RedactionType),
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
		if l := e.Location; l != nil {
			entity.Location = EntityLocation{
				StartIndex: deref(l.StartIndex), EndIndex: deref(l.EndIndex),
				StartIndexProcessed: deref(l.StartIndexProcessed), EndIndexProcessed: deref(l.EndIndexProcessed),
			}
		}
		out.Entities = append(out.Entities, entity)
	}
	return out
}

// fromGenRun fills in the run ID, which the API does not echo. A missing status is UNKNOWN.
func fromGenRun(runId string, r *gen.DetectRunsResponseV2) *GetRunResponse {
	status := DetectRunStatus(derefEnum(r.Status))
	if status == "" {
		status = DetectRunStatusUnknown
	}
	return &GetRunResponse{
		RunId:      runId,
		Status:     status,
		OutputType: DataSourceType(derefEnum(r.OutputType)),
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
			ProcessedFileExtension: FileDataFormat(derefEnum(o.ProcessedFileExtension)),
		})
	}
	return out
}

func fromGenMetrics(m *gen.Metrics) *Metrics {
	if m == nil {
		return nil
	}
	return &Metrics{
		Size:           deref(m.Size),
		WordCount:      deref(m.WordCount),
		CharacterCount: deref(m.CharacterCount),
		Slides:         m.Slides,
		Pages:          m.Pages,
		Duration:       m.Duration,
	}
}
