package utils

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMaskSecret(t *testing.T) {
	if MaskSecret("") != "" {
		t.Fatalf("expected empty string to stay empty")
	}
	if MaskSecret("value") != redactedValue {
		t.Fatalf("expected value to be redacted")
	}
}

func TestSanitizeHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer token")
	headers.Set("X-Api-Key", "key")
	headers.Set("Content-Type", "application/json")

	sanitized := SanitizeHeaders(headers)
	if sanitized.Get("Authorization") != redactedValue {
		t.Fatalf("expected authorization to be redacted")
	}
	if sanitized.Get("X-Api-Key") != redactedValue {
		t.Fatalf("expected api key to be redacted")
	}
	if sanitized.Get("Content-Type") != "application/json" {
		t.Fatalf("expected content-type to be preserved")
	}
	if headers.Get("Authorization") == redactedValue {
		t.Fatalf("expected original headers to remain unchanged")
	}
}

func TestSanitizeJSONNested(t *testing.T) {
	payload := []byte(`{"access_token":"abc","nested":{"refresh_token":"def","safe":"ok"},"list":[{"id_token":"ghi"},{"value":"ok"}]}`)
	sanitized := SanitizeJSON(payload)

	var decoded map[string]interface{}
	if err := json.Unmarshal(sanitized, &decoded); err != nil {
		t.Fatalf("failed to unmarshal sanitized json: %v", err)
	}

	if decoded["access_token"] != redactedValue {
		t.Fatalf("expected access_token to be redacted")
	}
	if decoded["nested"].(map[string]interface{})["refresh_token"] != redactedValue {
		t.Fatalf("expected refresh_token to be redacted")
	}
	if decoded["nested"].(map[string]interface{})["safe"] != "ok" {
		t.Fatalf("expected safe value to be preserved")
	}
	list := decoded["list"].([]interface{})
	if list[0].(map[string]interface{})["id_token"] != redactedValue {
		t.Fatalf("expected id_token to be redacted")
	}
}

func TestSanitizeJSONCaseInsensitive(t *testing.T) {
	payload := []byte(`{"Access_Token":"abc","Client_Secret":"def"}`)
	sanitized := SanitizeJSON(payload)

	var decoded map[string]interface{}
	if err := json.Unmarshal(sanitized, &decoded); err != nil {
		t.Fatalf("failed to unmarshal sanitized json: %v", err)
	}

	if decoded["Access_Token"] != redactedValue {
		t.Fatalf("expected Access_Token to be redacted")
	}
	if decoded["Client_Secret"] != redactedValue {
		t.Fatalf("expected Client_Secret to be redacted")
	}
}

func TestSanitizeJSONMalformed(t *testing.T) {
	payload := []byte(`{"access_token":`) // malformed
	sanitized := SanitizeJSON(payload)
	if string(sanitized) != string(payload) {
		t.Fatalf("expected malformed json to remain unchanged")
	}
}

func TestSanitizeStringHeadersAndForm(t *testing.T) {
	input := "Authorization: Bearer abc\nX-Api-Key: key\nContent-Type: text/plain\n\nclient_secret=secret&grant_type=password"
	sanitized := SanitizeString(input)

	if !containsLine(sanitized, "Authorization: "+redactedValue) {
		t.Fatalf("expected authorization header redacted")
	}
	if !containsLine(sanitized, "X-Api-Key: "+redactedValue) {
		t.Fatalf("expected api key header redacted")
	}
	if !containsLine(sanitized, "Content-Type: text/plain") {
		t.Fatalf("expected content-type preserved")
	}
	if !containsLine(sanitized, "client_secret="+redactedValue+"&grant_type=password") {
		t.Fatalf("expected form secret redacted")
	}
}

func TestSanitizeStringBasicAuth(t *testing.T) {
	input := "Authorization: Basic dGVzdDp0ZXN0"
	sanitized := SanitizeString(input)
	if sanitized != "Authorization: "+redactedValue {
		t.Fatalf("expected basic auth to be redacted")
	}
}

func containsLine(input, needle string) bool {
	return len(input) > 0 && strings.Contains(input, needle)
}
