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
				Transformation: &Transformation{ShiftDates: &ShiftDates{MinDays: 0, MaxDays: 30}},
			}},
			ObjectEntities: []ObjectEntity{{EntityType: ObjectEntityTypeFace, DeidentificationType: ObjectDeidentificationTypeRedact}},
			CustomTypes:    []CustomType{{Label: "EMP", Match: []string{"EMP-[0-9]+"}}},
			ReturnEntities: ReturnEntitiesTypeAll,
		},
		FileMapping: []FileMapping{{Source: "s3://in", Destination: "s3://out"}},
		Media: &Media{
			Audio:    &Audio{OutputProcessedAudio: true, OutputTranscription: OutputTranscriptionTypeTranscription, Bleep: &Bleep{StartPadding: 0.5, Frequency: 1000}},
			Document: &Document{Pdf: &Pdf{ProcessingMode: PdfProcessingModeOCR, Density: 300}},
			Image:    &Image{OutputOcrText: true, MaskingMethod: MaskingMethodBlur},
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
		t.Fatal("nil configuration should stay nil")
	}
}
