package detect

import (
	"encoding/json"
	"testing"
)

func TestConfigurationMapsToWireJSON(t *testing.T) {
	config := &DetectConfiguration{
		Name: "cfg",
		Detect: &Detect{
			Entities: []Entity{{
				EntityType: EntityTypeDOB, DeidentificationType: DeidentificationTypeEntityOnly, IncludePrefix: IncludePrefixNo,
				Transformation: &Transformation{ShiftDates: &ShiftDates{MinDays: ptr(0), MaxDays: ptr(30)}},
			}},
			ObjectEntities: []ObjectEntity{{EntityType: ObjectEntityTypeFace, DeidentificationType: ObjectDeidentificationTypeRedact}},
			CustomTypes:    []CustomType{{Label: "EMP", Match: []string{"EMP-[0-9]+"}}},
			ReturnEntities: ReturnEntitiesTypeAll,
		},
		FileMapping: []FileMapping{{Source: "s3://in", Destination: "s3://out"}},
		Media: &Media{
			Audio:    &Audio{OutputProcessedAudio: ptr(true), OutputTranscription: OutputTranscriptionTypeTranscription, Bleep: &Bleep{StartPadding: ptr[float32](0.5), Frequency: ptr(1000)}},
			Document: &Document{Pdf: &Pdf{ProcessingMode: PdfProcessingModeOCR, Density: ptr(300)}},
			Image:    &Image{OutputOcrText: ptr(true), MaskingMethod: MaskingMethodBlur},
		},
	}
	raw, err := json.Marshal(toGenConfiguration(config, "vault1"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	_ = json.Unmarshal(raw, &got)

	want := `{"name":"cfg","vaultId":"vault1",
		"detect":{"entities":[{"entityType":"DOB","deidentificationType":"ENTITY_ONLY","includePrefix":"NO","transformation":{"shiftDates":{"minDays":0,"maxDays":30}}}],
			"objectEntities":[{"entityType":"FACE","deidentificationType":"REDACT"}],
			"customTypes":[{"label":"EMP","match":["EMP-[0-9]+"]}],"returnEntities":"ALL"},
		"fileMapping":[{"source":"s3://in","destination":"s3://out"}],
		"media":{"audio":{"outputProcessedAudio":true,"outputTranscription":"TRANSCRIPTION","bleep":{"startPadding":0.5,"frequency":1000}},
			"document":{"pdf":{"processingMode":"OCR","density":300}},
			"image":{"outputOcrText":true,"maskingMethod":"BLUR"}}}`
	var expected map[string]interface{}
	if err := json.Unmarshal([]byte(want), &expected); err != nil {
		t.Fatal(err)
	}
	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(expected)
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("got  %s\nwant %s", gotJSON, wantJSON)
	}
	if toGenConfiguration(nil, "v") != nil {
		t.Fatal("a configuration ID alone sends no configuration")
	}
}

func ptr[T any](v T) *T { return &v }

// Media options are pointers so that zero and false can be sent, unlike nil.
func TestMediaOptionsSendZeroValues(t *testing.T) {
	media := &Media{
		Audio:    &Audio{OutputProcessedAudio: ptr(false), Bleep: &Bleep{StartPadding: ptr[float32](0), StopPadding: ptr[float32](0), Gain: ptr(0)}},
		Document: &Document{Pdf: &Pdf{Density: ptr(0)}},
		Image:    &Image{OutputProcessedImage: ptr(false), OutputOcrText: ptr(false)},
	}
	raw, err := json.Marshal(toGenConfiguration(&DetectConfiguration{Media: media}, "v").Media)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"audio":{"outputProcessedAudio":false,"bleep":{"startPadding":0,"stopPadding":0,"gain":0}},"document":{"pdf":{"density":0}},"image":{"outputProcessedImage":false,"outputOcrText":false}}`
	if string(raw) != want {
		t.Fatalf("got  %s\nwant %s", raw, want)
	}
	if raw, err = json.Marshal(toGenConfiguration(&DetectConfiguration{Media: &Media{Audio: &Audio{Bleep: &Bleep{}}}}, "v").Media); err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"audio":{"bleep":{}}}` {
		t.Fatalf("unset options should be omitted: %s", raw)
	}
}

// The source names are "" for nil, including a nil pointer, and a pointer gives the value
// it points to.
func TestSourceNames(t *testing.T) {
	var nilFlow *FlowDeidentifySourceType
	var nilSky *SkyDeidentifySourceType
	skyflowID := SkyDeidentifySourceTypeSkyflowID
	for source, want := range map[DeidentifyFileSource]string{nil: "", nilFlow: "", nilSky: "", &skyflowID: "SKYFLOW_ID", FlowDeidentifySourceTypeBase64: "BASE64"} {
		if got := deidentifySourceName(source); got != want {
			t.Errorf("deidentifySourceName(%#v) = %q, want %q", source, got, want)
		}
	}
	var nilFlowRe *FlowReidentifySourceType
	var nilSkyRe *SkyReidentifySourceType
	presigned := FlowReidentifySourceTypePresignedURL
	for source, want := range map[ReidentifyFileSource]string{nil: "", nilFlowRe: "", nilSkyRe: "", &presigned: "PRESIGNED_URL", SkyReidentifySourceTypeBase64: "BASE64"} {
		if got := reidentifySourceName(source); got != want {
			t.Errorf("reidentifySourceName(%#v) = %q, want %q", source, got, want)
		}
	}
}
