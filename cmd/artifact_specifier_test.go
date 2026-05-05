package cmd

import "testing"

func TestParseImportURLSpecifier_PreservesPortAndPathWithMainArtifactSuffix(t *testing.T) {
	in := "http://localhost:8080/spec.yaml:true"
	url, main, secret := parseImportURLSpecifier(in)

	if url != "http://localhost:8080/spec.yaml" {
		t.Fatalf("url mismatch: got %q", url)
	}
	if main != true {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
	if secret != "" {
		t.Fatalf("secret mismatch: got %q", secret)
	}
}

func TestParseImportURLSpecifier_PreservesQueryAndFragment(t *testing.T) {
	in := "https://example.com:8443/spec.yaml?x=1#frag:false"
	url, main, secret := parseImportURLSpecifier(in)

	if url != "https://example.com:8443/spec.yaml?x=1#frag" {
		t.Fatalf("url mismatch: got %q", url)
	}
	if main != false {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
	if secret != "" {
		t.Fatalf("secret mismatch: got %q", secret)
	}
}

func TestParseImportURLSpecifier_WithSecretSuffix(t *testing.T) {
	in := "http://localhost:8080/spec.yaml:true:mySecret"
	url, main, secret := parseImportURLSpecifier(in)

	if url != "http://localhost:8080/spec.yaml" {
		t.Fatalf("url mismatch: got %q", url)
	}
	if main != true {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
	if secret != "mySecret" {
		t.Fatalf("secret mismatch: got %q", secret)
	}
}

func TestParseImportURLSpecifier_NoSuffixes_Unchanged(t *testing.T) {
	in := "http://localhost:8080/spec.yaml"
	url, main, secret := parseImportURLSpecifier(in)

	if url != in {
		t.Fatalf("url mismatch: got %q", url)
	}
	if main != true {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
	if secret != "" {
		t.Fatalf("secret mismatch: got %q", secret)
	}
}

func TestParseImportURLSpecifier_PortOnly_NoSuffixes_Unchanged(t *testing.T) {
	in := "http://localhost:8080/spec.yaml:1234"
	url, main, secret := parseImportURLSpecifier(in)

	if url != in {
		t.Fatalf("url mismatch: got %q", url)
	}
	if main != true {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
	if secret != "" {
		t.Fatalf("secret mismatch: got %q", secret)
	}
}

func TestParseImportFileSpecifier_SuffixBool(t *testing.T) {
	in := "./specs/openapi.yaml:false"
	path, main := parseImportFileSpecifier(in)
	if path != "./specs/openapi.yaml" {
		t.Fatalf("path mismatch: got %q", path)
	}
	if main != false {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
}

func TestParseImportFileSpecifier_NoSuffix_Unchanged(t *testing.T) {
	in := "./specs/openapi.yaml"
	path, main := parseImportFileSpecifier(in)
	if path != in {
		t.Fatalf("path mismatch: got %q", path)
	}
	if main != true {
		t.Fatalf("mainArtifact mismatch: got %v", main)
	}
}

