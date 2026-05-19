package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

const redactedValue = "[REDACTED]"

var sensitiveKeys = map[string]struct{}{
	"authorization":       {},
	"proxy-authorization": {},
	"access_token":        {},
	"refresh_token":       {},
	"id_token":            {},
	"client_secret":       {},
	"clientsecret":        {},
	"password":            {},
	"token":               {},
	"api_key":             {},
	"apikey":              {},
	"secret":              {},
	"cookie":              {},
	"set-cookie":          {},
	"x-api-key":           {},
}

var (
	urlEncodedSecretPattern = regexp.MustCompile(`(?i)(access_token|refresh_token|id_token|client_secret|clientsecret|password|token|api_key|apikey|secret)=([^&\s]+)`)
	authSchemePattern       = regexp.MustCompile(`(?i)\b(bearer|basic)\s+([A-Za-z0-9\-._~+/=]+)`)
)

// MaskSecret replaces non-empty values with a redaction marker.
func MaskSecret(value string) string {
	if value == "" {
		return ""
	}
	return redactedValue
}

// SanitizeHeaders returns a sanitized copy of the headers.
func SanitizeHeaders(headers http.Header) http.Header {
	if headers == nil {
		return nil
	}
	sanitized := make(http.Header, len(headers))
	for key, values := range headers {
		if isSensitiveKey(key) {
			redactedValues := make([]string, len(values))
			for i := range redactedValues {
				redactedValues[i] = redactedValue
			}
			sanitized[key] = redactedValues
			continue
		}
		copied := make([]string, len(values))
		copy(copied, values)
		sanitized[key] = copied
	}
	return sanitized
}

// SanitizeJSON sanitizes sensitive fields in JSON payloads.
func SanitizeJSON(data []byte) []byte {
	if len(bytes.TrimSpace(data)) == 0 {
		return data
	}

	var decoded interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return data
	}

	cleaned := sanitizeValue(decoded)
	encoded, err := json.Marshal(cleaned)
	if err != nil {
		return data
	}
	return encoded
}

// SanitizeMap sanitizes sensitive fields in generic maps.
func SanitizeMap(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	return sanitizeMap(data)
}

// SanitizeString attempts to redact secrets in string payloads.
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	if sanitized, ok := sanitizeHTTPDump(input); ok {
		return sanitized
	}

	trimmed := strings.TrimSpace(input)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		if sanitized := SanitizeJSON([]byte(input)); sanitized != nil {
			return string(sanitized)
		}
	}

	sanitized := sanitizeHeaderLines(input)
	sanitized = urlEncodedSecretPattern.ReplaceAllString(sanitized, "$1="+redactedValue)
	sanitized = authSchemePattern.ReplaceAllString(sanitized, "$1 "+redactedValue)
	return sanitized
}

func sanitizeHeaderLines(input string) string {
	lines := strings.Split(input, "\n")
	previousSensitive := false

	for i, line := range lines {
		trimmedLine := strings.TrimLeft(line, " \t")
		leadingWhitespace := line[:len(line)-len(trimmedLine)]

		if previousSensitive && trimmedLine != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			lines[i] = leadingWhitespace + redactedValue
			continue
		}

		previousSensitive = false
		separatorIndex := strings.Index(trimmedLine, ":")
		if separatorIndex == -1 {
			continue
		}

		keyPart := trimmedLine[:separatorIndex]
		key := strings.TrimSpace(keyPart)
		if !isSensitiveKey(key) {
			continue
		}

		previousSensitive = true
		lines[i] = leadingWhitespace + keyPart + ": " + redactedValue
	}

	return strings.Join(lines, "\n")
}

func sanitizeHTTPDump(input string) (string, bool) {
	separator := "\r\n\r\n"
	index := strings.Index(input, separator)
	lineSep := "\r\n"
	if index == -1 {
		separator = "\n\n"
		index = strings.Index(input, separator)
		if index == -1 {
			return "", false
		}
		lineSep = "\n"
	}

	headersPart := input[:index]
	bodyPart := input[index+len(separator):]

	headersNormalized := strings.ReplaceAll(headersPart, "\r\n", "\n")
	headersSanitized := sanitizeHeaderLines(headersNormalized)
	headersSanitized = strings.ReplaceAll(headersSanitized, "\n", lineSep)

	bodySanitized := sanitizeBody(bodyPart)
	combined := headersSanitized + separator + bodySanitized
	combined = urlEncodedSecretPattern.ReplaceAllString(combined, "$1="+redactedValue)
	combined = authSchemePattern.ReplaceAllString(combined, "$1 "+redactedValue)
	return combined, true
}

func sanitizeBody(body string) string {
	trimmed := strings.TrimSpace(body)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		sanitized := SanitizeJSON([]byte(body))
		return string(sanitized)
	}
	return urlEncodedSecretPattern.ReplaceAllString(body, "$1="+redactedValue)
}

func sanitizeValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return sanitizeMap(typed)
	case []interface{}:
		cleaned := make([]interface{}, len(typed))
		for i, item := range typed {
			cleaned[i] = sanitizeValue(item)
		}
		return cleaned
	default:
		return value
	}
}

func sanitizeMap(data map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{}, len(data))
	for key, value := range data {
		if isSensitiveKey(key) {
			cleaned[key] = redactedValue
			continue
		}
		cleaned[key] = sanitizeValue(value)
	}
	return cleaned
}

func isSensitiveKey(key string) bool {
	_, ok := sensitiveKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}
