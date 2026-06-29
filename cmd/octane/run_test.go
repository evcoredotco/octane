package main

import "testing"

func TestParseParameterOverrides(t *testing.T) {
	t.Parallel()

	got, err := parseParameterOverrides([]string{
		"valid_idTag=AABBCC",
		"meterStart=0",
	})
	if err != nil {
		t.Fatalf("parseParameterOverrides: %v", err)
	}

	if got["valid_idTag"] != "AABBCC" {
		t.Errorf("valid_idTag: want %q, got %q", "AABBCC", got["valid_idTag"])
	}

	if got["meterStart"] != "0" {
		t.Errorf("meterStart: want %q, got %q", "0", got["meterStart"])
	}
}

func TestParseParameterOverridesRejectsMalformedValue(t *testing.T) {
	t.Parallel()

	_, err := parseParameterOverrides([]string{"valid_idTag"})
	if err == nil {
		t.Fatal("parseParameterOverrides: want malformed value error, got nil")
	}
}
