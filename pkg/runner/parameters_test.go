package runner

import "testing"

func TestBindStepParameters_ReplacesDeclaredPlaceholders(t *testing.T) {
	t.Parallel()

	got, err := bindStepParameters(
		`station "CP01" sends Authorize with idTag "{valid_idTag}"`,
		[]string{"valid_idTag"},
		map[string]string{"valid_idTag": "AABBCC"},
	)
	if err != nil {
		t.Fatalf("bindStepParameters: %v", err)
	}

	want := `station "CP01" sends Authorize with idTag "AABBCC"`
	if got != want {
		t.Errorf("bound step: want %q, got %q", want, got)
	}
}

func TestBindStepParameters_ReturnsErrorForMissingValue(t *testing.T) {
	t.Parallel()

	_, err := bindStepParameters(
		"connector {connectorId} is ready",
		[]string{"connectorId"},
		map[string]string{},
	)
	if err == nil {
		t.Fatal("bindStepParameters: want missing parameter error, got nil")
	}
}
