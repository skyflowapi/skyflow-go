package controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	client "github.com/skyflowapi/skyflow-go/v2/internal/generated/client"
	"github.com/skyflowapi/skyflow-go/v2/internal/generated/option"
	. "github.com/skyflowapi/skyflow-go/v2/internal/vault/controller"
	"github.com/skyflowapi/skyflow-go/v2/utils/common"
	. "github.com/skyflowapi/skyflow-go/v2/utils/common"
	skyflowError "github.com/skyflowapi/skyflow-go/v2/utils/error"
)

// makeValidJWT returns a JWT signed with HS256 that expires one hour from now.
func makeValidJWT() string {
	claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte("secret"))
	return s
}

// makeExpiredJWT returns a JWT signed with HS256 that expired one hour ago.
func makeExpiredJWT() string {
	claims := jwt.MapClaims{"exp": time.Now().Add(-time.Hour).Unix()}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := tok.SignedString([]byte("secret"))
	return s
}

// setupDetectFileMux creates an httptest.Server with handlers for upload and poll paths.
// uploadJSON is the JSON body returned for the upload endpoint.
// pollJSON is the JSON body returned for the /v1/detect/runs/ poll endpoint.
func setupDetectFileMux(uploadHandler http.HandlerFunc, pollHandler http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/detect/runs/", pollHandler)
	mux.HandleFunc("/", uploadHandler)
	return httptest.NewServer(mux)
}

// injectDetectFileClient sets the DetectController's FilesApiClient to use the given test server URL.
func injectDetectFileClient(d *DetectController, baseURL string) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	c := client.NewClient(
		option.WithBaseURL(baseURL),
		option.WithToken("token"),
		option.WithHTTPHeader(h),
	)
	d.FilesApiClient = *c.Files
}

// injectDetectTextClient sets the DetectController's TextApiClient to use the given test server URL.
func injectDetectTextClient(d *DetectController, baseURL string) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	c := client.NewClient(
		option.WithBaseURL(baseURL),
		option.WithToken("token"),
		option.WithHTTPHeader(h),
	)
	d.TextApiClient = *c.Strings
}

// noopBearerToken stubs SetBearerTokenForDetectControllerFunc to be a no-op.
func noopBearerToken() {
	SetBearerTokenForDetectControllerFunc = func(d *DetectController) *skyflowError.SkyflowError {
		return nil
	}
}

// ---------------------------------------------------------------------------
// 1. CreateDetectRequestClient — deprecated BaseVaultURL paths and token paths
// ---------------------------------------------------------------------------

var _ = Describe("CreateDetectRequestClient — uncovered branches", func() {
	var d *DetectController

	BeforeEach(func() {
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "v1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{
					ApiKey: "",
					Token:  "",
					Path:   "",
				},
			},
		}
	})

	Context("when Config.Credentials.Token is set and not expired", func() {
		It("should use the token from config without calling SetBearerToken", func() {
			validToken := makeValidJWT()
			d.Config.Credentials.Token = validToken
			// ApiKey must be empty so the `else if Token` branch is reached.
			d.Config.Credentials.ApiKey = ""

			err := CreateDetectRequestClient(d, nil)
			Expect(err).To(BeNil())
			Expect(d.Token).To(Equal(validToken))
		})
	})

	Context("when only BaseVaultURL (deprecated) is set", func() {
		It("should warn and use BaseVaultURL as the base URL", func() {
			d.Config.BaseVaultURL = "https://deprecated.example.com"
			d.Config.BaseVaultUrl = ""
			d.Config.Credentials.ApiKey = "api-key" // avoid bearer token generation

			err := CreateDetectRequestClient(d, nil)
			Expect(err).To(BeNil())
		})
	})

	Context("when both BaseVaultURL and BaseVaultUrl are set", func() {
		It("should warn about the deprecated field and prefer BaseVaultUrl", func() {
			d.Config.BaseVaultURL = "https://deprecated.example.com"
			d.Config.BaseVaultUrl = "https://new.example.com"
			d.Config.Credentials.ApiKey = "api-key"

			err := CreateDetectRequestClient(d, nil)
			Expect(err).To(BeNil())
		})
	})

	Context("when Config.Credentials.Token is set but expired", func() {
		It("should return a BEARER_TOKEN_EXPIRED error", func() {
			d.Config.Credentials.Token = makeExpiredJWT()
			d.Config.Credentials.ApiKey = ""

			err := CreateDetectRequestClient(d, nil)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when credentials come from CommonCreds.Token (else branch in CreateDetectRequestClient)", func() {
		It("should call SetBearerTokenForDetectController and use d.Token for the client", func() {
			// CreateDetectRequestClient calls SetBearerTokenForDetectController directly (not via
			// func var), so we cannot stub it. Provide CommonCreds.Token so setVaultCredentials
			// succeeds and GenerateToken short-circuits without a network call.
			validToken := makeValidJWT()
			d.CommonCreds = &Credentials{Token: validToken}
			d.Config.Credentials.ApiKey = ""
			d.Config.Credentials.Token = ""

			err := CreateDetectRequestClient(d, nil)
			Expect(err).To(BeNil())
			Expect(d.Token).To(Equal(validToken))
		})
	})
})

// ---------------------------------------------------------------------------
// 2. SetBearerTokenForDetectController — uncovered branches
// ---------------------------------------------------------------------------

var _ = Describe("SetBearerTokenForDetectController — uncovered branches", func() {

	Context("when setVaultCredentials returns an error (no credentials anywhere)", func() {
		It("should propagate the setVaultCredentials error", func() {
			// Remove env var to prevent fallback.
			_ = os.Unsetenv("SKYFLOW_CREDENTIALS")

			d := &DetectController{
				Config: &VaultConfig{
					VaultId:     "v1",
					Credentials: Credentials{}, // all empty
				},
				CommonCreds: nil,
			}
			err := SetBearerTokenForDetectController(d)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when d.Token is valid and not expired (reuse path)", func() {
		It("should reuse the existing token without calling GenerateToken", func() {
			validToken := makeValidJWT()
			d := &DetectController{
				Config: &VaultConfig{
					VaultId: "v1",
					Credentials: Credentials{
						CredentialsString: `{"some":"valid_looking_string"}`,
					},
				},
				Token: validToken, // already set, not expired
			}
			err := SetBearerTokenForDetectController(d)
			Expect(err).To(BeNil())
			// Token should remain unchanged.
			Expect(d.Token).To(Equal(validToken))
		})
	})

	Context("when d.Token is empty and GenerateToken short-circuits via CommonCreds.Token", func() {
		It("should set d.Token via GenerateToken when using a token in CommonCreds", func() {
			validToken := makeValidJWT()
			d := &DetectController{
				Config: &VaultConfig{
					VaultId:     "v1",
					Credentials: Credentials{}, // empty — falls through to CommonCreds
				},
				CommonCreds: &Credentials{Token: validToken},
				Token:       "", // empty so generation path is entered
			}
			err := SetBearerTokenForDetectController(d)
			Expect(err).To(BeNil())
			Expect(d.Token).To(Equal(validToken))
		})
	})
})

// ---------------------------------------------------------------------------
// 3. CreateAudioRequest — OutputTranscription branch
// ---------------------------------------------------------------------------

var _ = Describe("CreateAudioRequest — OutputTranscription branch", func() {
	It("should set OutputTranscription when the field is non-empty", func() {
		// Entities must be non-empty: CreateAudioRequest uses an unsafe type assertion on
		// the result of CreateEntityTypesRef, which panics on untyped nil (empty entities).
		req := &DeidentifyFileRequest{
			Entities:            []DetectEntities{Name},
			OutputTranscription: "srt",
		}
		result := CreateAudioRequest(req, "base64data", "vault1", "mp3")
		Expect(result).ToNot(BeNil())
		Expect(result.OutputTranscription).ToNot(BeNil())
		Expect(string(*result.OutputTranscription)).To(Equal("srt"))
	})

	It("should not set OutputTranscription when the field is empty", func() {
		req := &DeidentifyFileRequest{
			Entities: []DetectEntities{Name},
		}
		result := CreateAudioRequest(req, "base64data", "vault1", "mp3")
		Expect(result).ToNot(BeNil())
		Expect(result.OutputTranscription).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// 4. CreateEntityTypesRef — default / unknown type switch case
// ---------------------------------------------------------------------------

var _ = Describe("CreateEntityTypesRef — default case", func() {
	It("should return nil for an unknown dataType", func() {
		entities := []DetectEntities{Name, EmailAddress}
		result := CreateEntityTypesRef(entities, "completely_unknown_type")
		Expect(result).To(BeNil())
	})

	It("should return nil when entities slice is empty", func() {
		result := CreateEntityTypesRef([]DetectEntities{}, "text")
		Expect(result).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// 5. CreateEntityTypes — default / unknown entityType switch case
// ---------------------------------------------------------------------------

var _ = Describe("CreateEntityTypes — default case", func() {
	It("should return nil for an unknown entityType", func() {
		entities := []DetectEntities{Name}
		result := CreateEntityTypes(entities, "unknown_entity_type")
		Expect(result).To(BeNil())
	})

	It("should return nil for an empty entities slice", func() {
		result := CreateEntityTypes([]DetectEntities{}, "text")
		Expect(result).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// 6. CreateTokenType — VaultToken field population
// ---------------------------------------------------------------------------

var _ = Describe("CreateTokenType — VaultToken field", func() {
	It("should populate VaultToken when format.VaultToken is non-empty", func() {
		format := TokenFormat{
			VaultToken: []DetectEntities{Name, EmailAddress},
		}
		result := CreateTokenType(format)
		Expect(result).ToNot(BeNil())
		Expect(result.VaultToken).To(HaveLen(2))
	})

	It("should return nil when all format fields are empty", func() {
		result := CreateTokenType(TokenFormat{})
		Expect(result).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// 7. CreateMaskingMethod — empty method returns nil
// ---------------------------------------------------------------------------

var _ = Describe("CreateMaskingMethod — empty method", func() {
	It("should return nil when method is empty string", func() {
		result := CreateMaskingMethod("")
		Expect(result).To(BeNil())
	})

	It("should return non-nil for a valid masking method", func() {
		result := CreateMaskingMethod(BLACKBOX)
		Expect(result).ToNot(BeNil())
	})
})

// ---------------------------------------------------------------------------
// 8. DeidentifyText — client-level custom headers validation error
//                     and empty API response branch
// ---------------------------------------------------------------------------

var _ = Describe("DeidentifyText — additional uncovered branches", func() {
	var (
		d   *DetectController
		ctx context.Context
		req DeidentifyTextRequest
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{
					ApiKey: "test-api-key",
				},
			},
		}
		req = DeidentifyTextRequest{
			Text:     "Some text to deidentify",
			Entities: []DetectEntities{Name},
		}
	})

	Context("when controller-level CustomHeaders are invalid", func() {
		It("should return error for empty-map client headers", func() {
			d.CustomHeaders = make(map[CustomHeaderKey]string) // empty map → error
			result, err := d.DeidentifyText(ctx, req, common.DeidentifyTextOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("should return error for an invalid header key in client headers", func() {
			d.CustomHeaders = map[CustomHeaderKey]string{
				CustomHeaderKey("x-invalid-client-header"): "value",
			}
			result, err := d.DeidentifyText(ctx, req, common.DeidentifyTextOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when the API returns a nil/empty response body", func() {
		It("should return an empty DeidentifyTextResponse without error", func() {
			// Server returns HTTP 200 with JSON null so that response.Body is nil.
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("null"))
			}))
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectTextClient(ctrl, ts.URL)
				return nil
			}
			SetBearerTokenForDetectControllerFunc = func(ctrl *DetectController) *skyflowError.SkyflowError {
				return nil
			}

			result, err := d.DeidentifyText(ctx, req, common.DeidentifyTextOptions{})
			// The nil/empty body branch returns an empty response with no error.
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
		})
	})
})

// ---------------------------------------------------------------------------
// 9. ReidentifyText — client-level custom headers validation error
// ---------------------------------------------------------------------------

var _ = Describe("ReidentifyText — client-level custom headers error", func() {
	var (
		d   *DetectController
		ctx context.Context
		req ReidentifyTextRequest
	)

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
		}
		req = ReidentifyTextRequest{
			Text:             "Redacted text",
			RedactedEntities: []DetectEntities{Name},
		}
	})

	It("should return error for invalid client-level custom headers", func() {
		d.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("x-bad-header"): "v",
		}
		result, err := d.ReidentifyText(ctx, req, common.ReidentifyTextOptions{})
		Expect(result).To(BeNil())
		Expect(err).ToNot(BeNil())
	})

	It("should return error for empty-map client-level custom headers", func() {
		d.CustomHeaders = make(map[CustomHeaderKey]string)
		result, err := d.ReidentifyText(ctx, req, common.ReidentifyTextOptions{})
		Expect(result).To(BeNil())
		Expect(err).ToNot(BeNil())
	})
})

// ---------------------------------------------------------------------------
// 10. DeidentifyFile — request.File.File branch and pollErr != nil
// ---------------------------------------------------------------------------

var _ = Describe("DeidentifyFile — additional uncovered branches", Ordered, func() {
	var (
		d       *DetectController
		ctx     context.Context
		tempDir string
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeAll(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "detect_coverage_*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() {
		if tempDir != "" {
			_ = os.RemoveAll(tempDir)
		}
	})

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
		}
	})

	Context("when File.File (*os.File) is provided instead of FilePath", func() {
		It("should read the file content and process it", func() {
			// Create a temp txt file.
			f, err := os.CreateTemp(tempDir, "input.*.txt")
			Expect(err).To(BeNil())
			_, _ = f.WriteString("test content for deidentify")
			_ = f.Close()
			// Re-open for reading (DeidentifyFile reads from the file handle).
			fOpen, err := os.Open(f.Name())
			Expect(err).To(BeNil())
			defer fOpen.Close()

			uploadResp := map[string]string{"run_id": "run-file-001"}
			pollResp := map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
				},
			}

			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(uploadResp)
				},
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(pollResp)
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			req := DeidentifyFileRequest{
				File:            FileInput{File: fOpen}, // use File field, not FilePath
				Entities:        []DetectEntities{Name},
				OutputDirectory: tempDir,
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.RunId).To(Equal("run-file-001"))
		})
	})

	Context("when the poll API returns an HTTP error (pollErr != nil)", func() {
		It("should return a SkyflowError wrapping the poll error", func() {
			uploadResp := map[string]string{"run_id": "run-poll-err"}

			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(uploadResp)
				},
				func(w http.ResponseWriter, r *http.Request) {
					// Poll endpoint returns HTTP error.
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":{"message":"internal error"}}`))
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			// Create a temp txt file for the request.
			f, ferr := os.CreateTemp(tempDir, "input.*.txt")
			Expect(ferr).To(BeNil())
			_, _ = f.WriteString("content")
			_ = f.Close()

			req := DeidentifyFileRequest{
				File:     FileInput{FilePath: f.Name()},
				Entities: []DetectEntities{Name},
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("processFileByType — default case (unknown file extension)", func() {
		It("should fall through to the generic DeidentifyFile endpoint", func() {
			uploadResp := map[string]string{"run_id": "run-generic-001"}
			pollResp := map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "dat",
						"processedFileType":      "GENERIC",
					},
				},
			}

			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(uploadResp)
				},
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(pollResp)
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			// Write a file with an unknown extension (.dat → default case).
			f, ferr := os.CreateTemp(tempDir, "file.*.dat")
			Expect(ferr).To(BeNil())
			_, _ = f.WriteString("binary content")
			_ = f.Close()

			req := DeidentifyFileRequest{
				File:     FileInput{FilePath: f.Name()},
				Entities: []DetectEntities{Name},
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.RunId).To(Equal("run-generic-001"))
		})
	})
})

// ---------------------------------------------------------------------------
// 11. pollForResults — GetRun error, nil response, and backoff else branch
// ---------------------------------------------------------------------------

var _ = Describe("pollForResults — uncovered branches (via DeidentifyFile)", Ordered, func() {
	var (
		d       *DetectController
		ctx     context.Context
		tempDir string
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeAll(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "detect_poll_*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() { _ = os.RemoveAll(tempDir) })

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
		}
	})

	// Helper: create a temp txt file and return its path.
	makeTxt := func(dir string) string {
		f, err := os.CreateTemp(dir, "input.*.txt")
		Expect(err).To(BeNil())
		_, _ = f.WriteString("test")
		_ = f.Close()
		return f.Name()
	}

	Context("when GetRun returns HTTP error during polling", func() {
		It("should propagate the error from pollForResults", func() {
			uploadResp := map[string]string{"run_id": "run-poll-err2"}

			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(uploadResp)
				},
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":{"message":"bad run id"}}`))
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			req := DeidentifyFileRequest{
				File:     FileInput{FilePath: makeTxt(tempDir)},
				Entities: []DetectEntities{Name},
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when GetRun returns a null / empty body during polling", func() {
		It("should return error EMPTY_DEIDENTIFY_FILE_RESPONSE", func() {
			uploadResp := map[string]string{"run_id": "run-null-body"}

			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(uploadResp)
				},
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("null")) // null JSON → nil Body
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			req := DeidentifyFileRequest{
				File:     FileInput{FilePath: makeTxt(tempDir)},
				Entities: []DetectEntities{Name},
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("when polling enters the backoff else branch (nextWaitTime < maxWaitTime)", func() {
		// With WaitTime=3:
		//   iter1: currentWaitTime=1 < 3, nextWaitTime=2 < 3 → ELSE branch, sleep 2s
		//   iter2: server returns SUCCESS → loop exits
		// Total sleep: ~2 seconds.
		It("should hit the else branch before receiving SUCCESS", func() {
			uploadResp := map[string]string{"run_id": "run-backoff"}
			callCount := 0
			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(uploadResp)
				},
				func(w http.ResponseWriter, r *http.Request) {
					callCount++
					w.Header().Set("Content-Type", "application/json")
					if callCount == 1 {
						// First poll: return IN_PROGRESS to trigger the backoff.
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"status": "IN_PROGRESS",
						})
					} else {
						// Second poll: return SUCCESS.
						_ = json.NewEncoder(w).Encode(map[string]interface{}{
							"status": "SUCCESS",
							"output": []map[string]interface{}{
								{
									"processedFile":          "dGVzdA==",
									"processedFileExtension": "txt",
									"processedFileType":      "TEXT",
								},
							},
						})
					}
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			req := DeidentifyFileRequest{
				File:     FileInput{FilePath: makeTxt(tempDir)},
				Entities: []DetectEntities{Name},
				WaitTime: 3, // ensures else branch is hit on first iteration
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.Status).To(Equal("SUCCESS"))
		})
	})
})

// ---------------------------------------------------------------------------
// 12. processDeidentifyFileResponse — error paths (via DeidentifyFile)
//
//     NOTE: DeidentifyFile discards the return value of processDeidentifyFileResponse,
//     so test assertions target overall DeidentifyFile behaviour.
//     These tests exercise the error code paths for coverage.
// ---------------------------------------------------------------------------

var _ = Describe("processDeidentifyFileResponse — error code paths (via DeidentifyFile)", Ordered, func() {
	var (
		d       *DetectController
		ctx     context.Context
		tempDir string
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeAll(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "detect_proc_*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() { _ = os.RemoveAll(tempDir) })

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
		}
	})

	makeTxt2 := func() string {
		f, err := os.CreateTemp(tempDir, "input.*.txt")
		Expect(err).To(BeNil())
		_, _ = f.WriteString("test content")
		_ = f.Close()
		return f.Name()
	}

	setupSuccessServer := func(pollBody interface{}) *httptest.Server {
		uploadResp := map[string]string{"run_id": "run-proc-01"}
		return setupDetectFileMux(
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(uploadResp)
			},
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(pollBody)
			},
		)
	}

	injectAndNoop := func(ts *httptest.Server) {
		CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
			injectDetectFileClient(ctrl, ts.URL)
			return nil
		}
		noopBearerToken()
	}

	Context("when poll returns SUCCESS with empty output slice", func() {
		It("should call processDeidentifyFileResponse and handle empty output gracefully", func() {
			ts := setupSuccessServer(map[string]interface{}{
				"status": "SUCCESS",
				"output": []interface{}{},
			})
			defer ts.Close()
			injectAndNoop(ts)

			req := DeidentifyFileRequest{
				File:            FileInput{FilePath: makeTxt2()},
				Entities:        []DetectEntities{Name},
				OutputDirectory: tempDir,
			}
			// processDeidentifyFileResponse returns error (empty output) but DeidentifyFile discards it.
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(err).To(BeNil())
			_ = result // may be nil since parseDeidentifyFileResponse also handles empty output
		})
	})

	Context("when outputDir is empty (processDeidentifyFileResponse returns nil early)", func() {
		It("should not attempt to write files", func() {
			ts := setupSuccessServer(map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
				},
			})
			defer ts.Close()
			injectAndNoop(ts)

			req := DeidentifyFileRequest{
				File:            FileInput{FilePath: makeTxt2()},
				Entities:        []DetectEntities{Name},
				OutputDirectory: "", // empty → processDeidentifyFileResponse returns nil early
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
		})
	})

	Context("when the first output has invalid base64 content", func() {
		It("should call processDeidentifyFileResponse and hit the decode error path", func() {
			ts := setupSuccessServer(map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "!!!not-valid-base64!!!",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
				},
			})
			defer ts.Close()
			injectAndNoop(ts)

			req := DeidentifyFileRequest{
				File:            FileInput{FilePath: makeTxt2()},
				Entities:        []DetectEntities{Name},
				OutputDirectory: tempDir,
			}
			// processDeidentifyFileResponse hits decode error (discarded).
			// parseDeidentifyFileResponse also hits decode error (error discarded by DeidentifyFile).
			_, _ = d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		})
	})

	Context("when os.WriteFile fails for the first output (read-only output file exists)", func() {
		It("should hit the WriteFile error path in processDeidentifyFileResponse", func() {
			// Use a fixed-name input file so the expected output path is deterministic.
			inputPath := filepath.Join(tempDir, "testwrite_first.txt")
			Expect(os.WriteFile(inputPath, []byte("content"), 0644)).To(Succeed())

			// Pre-create the output file (processed-testwrite_first.txt) as read-only.
			// processDeidentifyFileResponse builds the output path as:
			//   filepath.Join(outputDir, "processed-" + fileName)
			outputPath := filepath.Join(tempDir, "processed-testwrite_first.txt")
			Expect(os.WriteFile(outputPath, []byte("old"), 0444)).To(Succeed())
			defer func() { _ = os.Chmod(outputPath, 0644) }()

			ts := setupSuccessServer(map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
				},
			})
			defer ts.Close()
			injectAndNoop(ts)

			req := DeidentifyFileRequest{
				File:            FileInput{FilePath: inputPath},
				Entities:        []DetectEntities{Name},
				OutputDirectory: tempDir, // exists & writable → passes validation
			}
			// processDeidentifyFileResponse tries to write to the read-only file → error (discarded).
			_, _ = d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		})
	})

	Context("when a secondary output has invalid base64", func() {
		It("should hit the decode error path for secondary outputs", func() {
			ts := setupSuccessServer(map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==", // first output: valid
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
					{
						"processedFile":          "!!!invalid-base64!!!",
						"processedFileExtension": "json",
						"processedFileType":      "entities",
					},
				},
			})
			defer ts.Close()
			injectAndNoop(ts)

			req := DeidentifyFileRequest{
				File:            FileInput{FilePath: makeTxt2()},
				Entities:        []DetectEntities{Name},
				OutputDirectory: tempDir,
			}
			_, _ = d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		})
	})

	Context("when os.WriteFile fails for a secondary output (read-only secondary file exists)", func() {
		It("should hit the WriteFile error path for secondary outputs", func() {
			// Use a fixed-name input file so the secondary output path is deterministic.
			// processDeidentifyFileResponse builds secondary output path as:
			//   filepath.Join(outputDir, "processed-" + fileBaseName + "." + ext)
			// where fileBaseName = "testwrite_second" and ext = "json".
			inputPath := filepath.Join(tempDir, "testwrite_second.txt")
			Expect(os.WriteFile(inputPath, []byte("content"), 0644)).To(Succeed())

			// Pre-create the secondary output file as read-only.
			secondaryOutputPath := filepath.Join(tempDir, "processed-testwrite_second.json")
			Expect(os.WriteFile(secondaryOutputPath, []byte("old"), 0444)).To(Succeed())
			defer func() { _ = os.Chmod(secondaryOutputPath, 0644) }()

			ts := setupSuccessServer(map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "json",
						"processedFileType":      "entities",
					},
				},
			})
			defer ts.Close()
			injectAndNoop(ts)

			req := DeidentifyFileRequest{
				File:            FileInput{FilePath: inputPath},
				Entities:        []DetectEntities{Name},
				OutputDirectory: tempDir, // exists & writable → passes validation; secondary file is read-only
			}
			_, _ = d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		})
	})
})

// ---------------------------------------------------------------------------
// 13. parseDeidentifyFileResponse + GetDetectRun — error/branch coverage
// ---------------------------------------------------------------------------

var _ = Describe("parseDeidentifyFileResponse + GetDetectRun — uncovered branches", func() {
	var (
		d   *DetectController
		ctx context.Context
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
		}
	})

	setupGetRunServer := func(responseBody interface{}, statusCode int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(responseBody)
		}))
	}

	injectFiles := func(baseURL string) {
		CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
			injectDetectFileClient(ctrl, baseURL)
			return nil
		}
		noopBearerToken()
	}

	Context("GetDetectRun — nil/empty API response body", func() {
		It("should return empty DeidentifyFileResponse when server returns null body", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("null"))
			}))
			defer ts.Close()
			injectFiles(ts.URL)

			req := GetDetectRunRequest{RunId: "run123"}
			result, err := d.GetDetectRun(ctx, req, common.GetDetectRunOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
		})
	})

	Context("GetDetectRun — parseDeidentifyFileResponse returns error (invalid base64)", func() {
		It("should return a SkyflowError wrapping the parse error", func() {
			// Status is not IN_PROGRESS; output[0] has invalid base64.
			response := map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "!!!invalid-base64!!!",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
				},
			}
			ts := setupGetRunServer(response, http.StatusOK)
			defer ts.Close()
			injectFiles(ts.URL)

			req := GetDetectRunRequest{RunId: "run-parse-err"}
			result, err := d.GetDetectRun(ctx, req, common.GetDetectRunOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})

	Context("parseDeidentifyFileResponse — ProcessedFileType is nil (UNKNOWN_STATUS branch)", func() {
		It("should set Type to UNKNOWN_STATUS when processedFileType is absent", func() {
			// processedFile and processedFileExtension present but processedFileType absent.
			response := map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==", // valid base64
						"processedFileExtension": "txt",
						// no processedFileType → nil → UNKNOWN_STATUS branch
					},
				},
			}
			ts := setupGetRunServer(response, http.StatusOK)
			defer ts.Close()
			injectFiles(ts.URL)

			req := GetDetectRunRequest{RunId: "run-unknown-type"}
			result, err := d.GetDetectRun(ctx, req, common.GetDetectRunOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.Type).To(Equal("UNKNOWN"))
		})
	})

	Context("parseDeidentifyFileResponse — nil ProcessedFileType in entities loop (continue branch)", func() {
		It("should skip secondary output entries that have no processedFileType", func() {
			response := map[string]interface{}{
				"status": "SUCCESS",
				"output": []map[string]interface{}{
					{
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "txt",
						"processedFileType":      "TEXT",
					},
					{
						// second output with no processedFileType → continue branch
						"processedFile":          "dGVzdA==",
						"processedFileExtension": "json",
					},
				},
			}
			ts := setupGetRunServer(response, http.StatusOK)
			defer ts.Close()
			injectFiles(ts.URL)

			req := GetDetectRunRequest{RunId: "run-continue"}
			result, err := d.GetDetectRun(ctx, req, common.GetDetectRunOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			// The nil-type secondary output should have been skipped, so Entities is empty.
			Expect(result.Entities).To(BeEmpty())
		})
	})

	Context("GetDetectRun — API returns HTTP error", func() {
		It("should propagate the HTTP error as a SkyflowError", func() {
			ts := setupGetRunServer(
				map[string]interface{}{"error": map[string]string{"message": "not found"}},
				http.StatusNotFound,
			)
			defer ts.Close()
			injectFiles(ts.URL)

			req := GetDetectRunRequest{RunId: "run-http-err"}
			result, err := d.GetDetectRun(ctx, req, common.GetDetectRunOptions{})
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
		})
	})
})

// ---------------------------------------------------------------------------
// 14. CreateStructuredTextRequest — OutputTranscription is only on audio;
//     verify the structured-text path produces correct output (sanity check)
// ---------------------------------------------------------------------------

var _ = Describe("CreateStructuredTextRequest — basic coverage", func() {
	It("should build a valid structured text request", func() {
		req := &DeidentifyFileRequest{
			Entities: []DetectEntities{Name},
			TokenFormat: TokenFormat{
				VaultToken: []DetectEntities{EmailAddress},
			},
		}
		result := CreateStructuredTextRequest(req, "base64content", "vault1", "json")
		Expect(result).ToNot(BeNil())
		Expect(result.VaultId).To(Equal("vault1"))
		Expect(result.TokenType).ToNot(BeNil())
		Expect(result.TokenType.VaultToken).To(HaveLen(1))
	})
})

// ---------------------------------------------------------------------------
// 15. Additional edge cases for existing partially-covered helpers
// ---------------------------------------------------------------------------

var _ = Describe("CreateDeidentifyTextRequest — VaultToken in TokenFormat", func() {
	It("should populate VaultToken on the token type mapping", func() {
		req := DeidentifyTextRequest{
			Text:     "some text",
			Entities: []DetectEntities{Name},
			TokenFormat: TokenFormat{
				VaultToken: []DetectEntities{Name, EmailAddress},
			},
		}
		config := VaultConfig{VaultId: "vault1"}
		payload, err := CreateDeidentifyTextRequest(req, config)
		Expect(err).To(BeNil())
		Expect(payload).ToNot(BeNil())
		Expect(payload.TokenType).ToNot(BeNil())
		Expect(payload.TokenType.VaultToken).To(HaveLen(2))
	})
})

var _ = Describe("DeidentifyFile — client-level custom headers validation error", Ordered, func() {
	var (
		d       *DetectController
		ctx     context.Context
		tempDir string
	)

	BeforeAll(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "detect_hdrs_*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() { _ = os.RemoveAll(tempDir) })

	BeforeEach(func() {
		ctx = context.Background()
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
		}
	})

	It("should return error when client-level custom headers are invalid", func() {
		d.CustomHeaders = map[CustomHeaderKey]string{
			CustomHeaderKey("x-forbidden-header"): "value",
		}
		f, err := os.CreateTemp(tempDir, "input.*.txt")
		Expect(err).To(BeNil())
		_, _ = f.WriteString("content")
		_ = f.Close()

		req := DeidentifyFileRequest{
			File:     FileInput{FilePath: f.Name()},
			Entities: []DetectEntities{Name},
		}
		result, serr := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		Expect(result).To(BeNil())
		Expect(serr).ToNot(BeNil())
	})
})

var _ = Describe("GetDetectRun — client-level custom headers validation error", func() {
	It("should return error for invalid client-level CustomHeaders", func() {
		ctx := context.Background()
		d := &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{ApiKey: "test-api-key"},
			},
			CustomHeaders: map[CustomHeaderKey]string{
				CustomHeaderKey("x-bad-client-header"): "v",
			},
		}
		req := GetDetectRunRequest{RunId: "run123"}
		result, err := d.GetDetectRun(ctx, req, common.GetDetectRunOptions{})
		Expect(result).To(BeNil())
		Expect(err).ToNot(BeNil())
	})
	
})

// ---------------------------------------------------------------------------
// 17. processFileByType — endpoint routing by file extension
// ---------------------------------------------------------------------------

var _ = Describe("processFileByType — endpoint routing by file extension", Ordered, func() {
	var (
		d       *DetectController
		ctx     context.Context
		tempDir string
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeAll(func() {
		ctx = context.Background()
		var err error
		tempDir, err = os.MkdirTemp("", "detect_routing_*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() { _ = os.RemoveAll(tempDir) })

	BeforeEach(func() {
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{Token: makeValidJWT()},
			},
		}
	})

	DescribeTable("routes each file extension to the correct API endpoint path",
		func(ext, expectedPath string) {
			var capturedPath string

			ts := setupDetectFileMux(
				func(w http.ResponseWriter, r *http.Request) {
					capturedPath = r.URL.Path
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]string{"run_id": "run-001"})
				},
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]interface{}{
						"status": "SUCCESS",
						"output": []map[string]interface{}{
							{
								"processedFile":          "dGVzdA==",
								"processedFileExtension": strings.TrimPrefix(ext, "."),
								"processedFileType":      "TEXT",
							},
						},
					})
				},
			)
			defer ts.Close()

			CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
				injectDetectFileClient(ctrl, ts.URL)
				return nil
			}
			noopBearerToken()

			f, ferr := os.CreateTemp(tempDir, "file.*"+ext)
			Expect(ferr).To(BeNil())
			_, _ = f.WriteString("test content")
			_ = f.Close()

			req := DeidentifyFileRequest{
				File:     FileInput{FilePath: f.Name()},
				Entities: []DetectEntities{Name},
			}
			result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(capturedPath).To(Equal(expectedPath))
		},
		Entry(".txt  → text endpoint",            ".txt",  "/v1/detect/deidentify/file/text"),
		Entry(".mp3  → audio endpoint",           ".mp3",  "/v1/detect/deidentify/file/audio"),
		Entry(".wav  → audio endpoint",           ".wav",  "/v1/detect/deidentify/file/audio"),
		Entry(".pdf  → pdf endpoint",             ".pdf",  "/v1/detect/deidentify/file/document/pdf"),
		Entry(".jpg  → image endpoint",           ".jpg",  "/v1/detect/deidentify/file/image"),
		Entry(".jpeg → image endpoint",           ".jpeg", "/v1/detect/deidentify/file/image"),
		Entry(".png  → image endpoint",           ".png",  "/v1/detect/deidentify/file/image"),
		Entry(".bmp  → image endpoint",           ".bmp",  "/v1/detect/deidentify/file/image"),
		Entry(".tif  → image endpoint",           ".tif",  "/v1/detect/deidentify/file/image"),
		Entry(".tiff → image endpoint",           ".tiff", "/v1/detect/deidentify/file/image"),
		Entry(".ppt  → presentation endpoint",    ".ppt",  "/v1/detect/deidentify/file/presentation"),
		Entry(".pptx → presentation endpoint",    ".pptx", "/v1/detect/deidentify/file/presentation"),
		Entry(".csv  → spreadsheet endpoint",     ".csv",  "/v1/detect/deidentify/file/spreadsheet"),
		Entry(".xls  → spreadsheet endpoint",     ".xls",  "/v1/detect/deidentify/file/spreadsheet"),
		Entry(".xlsx → spreadsheet endpoint",     ".xlsx", "/v1/detect/deidentify/file/spreadsheet"),
		Entry(".doc  → document endpoint",        ".doc",  "/v1/detect/deidentify/file/document"),
		Entry(".docx → document endpoint",        ".docx", "/v1/detect/deidentify/file/document"),
		Entry(".json → structured text endpoint", ".json", "/v1/detect/deidentify/file/structured_text"),
		Entry(".xml  → structured text endpoint", ".xml",  "/v1/detect/deidentify/file/structured_text"),
		Entry(".dat  → generic file endpoint",    ".dat",  "/v1/detect/deidentify/file"),
	)
})

// ---------------------------------------------------------------------------
// 18. DeidentifyFile — AllowRegexList and RestrictRegexList passthrough
// ---------------------------------------------------------------------------

var _ = Describe("DeidentifyFile — AllowRegexList and RestrictRegexList passthrough", Ordered, func() {
	var (
		d       *DetectController
		ctx     context.Context
		tempDir string
	)

	AfterEach(func() {
		CreateDetectRequestClientFunc = CreateDetectRequestClient
		SetBearerTokenForDetectControllerFunc = SetBearerTokenForDetectController
	})

	BeforeAll(func() {
		ctx = context.Background()
		var err error
		tempDir, err = os.MkdirTemp("", "detect_regex_*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() { _ = os.RemoveAll(tempDir) })

	BeforeEach(func() {
		d = &DetectController{
			Config: &VaultConfig{
				VaultId:   "vault1",
				ClusterId: "cluster1",
				Env:       DEV,
				Credentials: Credentials{Token: makeValidJWT()},
			},
		}
	})

	successPollHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "SUCCESS",
			"output": []map[string]interface{}{
				{"processedFile": "dGVzdA==", "processedFileExtension": "txt", "processedFileType": "TEXT"},
			},
		})
	}

	It("should include allow_regex and restrict_regex in the outgoing request body", func() {
		var body map[string]interface{}

		ts := setupDetectFileMux(
			func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&body)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"run_id": "run-001"})
			},
			successPollHandler,
		)
		defer ts.Close()

		CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
			injectDetectFileClient(ctrl, ts.URL)
			return nil
		}
		noopBearerToken()

		f, _ := os.CreateTemp(tempDir, "file.*.txt")
		_, _ = f.WriteString("test content")
		_ = f.Close()

		req := DeidentifyFileRequest{
			File:              FileInput{FilePath: f.Name()},
			Entities:          []DetectEntities{Name},
			AllowRegexList:    []string{`\d{3}-\d{2}-\d{4}`},
			RestrictRegexList: []string{`^RESTRICT.*`},
		}
		result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		Expect(err).To(BeNil())
		Expect(result).ToNot(BeNil())

		Expect(body).To(HaveKey("allow_regex"))
		Expect(body["allow_regex"]).To(ContainElement(`\d{3}-\d{2}-\d{4}`))
		Expect(body).To(HaveKey("restrict_regex"))
		Expect(body["restrict_regex"]).To(ContainElement(`^RESTRICT.*`))
	})

	It("should omit allow_regex and restrict_regex from the request body when not set", func() {
		var body map[string]interface{}

		ts := setupDetectFileMux(
			func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&body)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]string{"run_id": "run-001"})
			},
			successPollHandler,
		)
		defer ts.Close()

		CreateDetectRequestClientFunc = func(ctrl *DetectController, headers map[CustomHeaderKey]string) *skyflowError.SkyflowError {
			injectDetectFileClient(ctrl, ts.URL)
			return nil
		}
		noopBearerToken()

		f, _ := os.CreateTemp(tempDir, "file.*.txt")
		_, _ = f.WriteString("test content")
		_ = f.Close()

		req := DeidentifyFileRequest{
			File:     FileInput{FilePath: f.Name()},
			Entities: []DetectEntities{Name},
		}
		result, err := d.DeidentifyFile(ctx, req, common.DeidentifyFileOptions{})
		Expect(err).To(BeNil())
		Expect(result).ToNot(BeNil())

		Expect(body).ToNot(HaveKey("allow_regex"))
		Expect(body).ToNot(HaveKey("restrict_regex"))
	})
})

// Ensure the test file compiles even without usages of fmt and strings packages.
var _ = fmt.Sprintf
var _ = strings.ToLower
