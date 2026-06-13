package handler

import "testing"

func TestParseIncludeCompleted(t *testing.T) {
	t.Parallel()

	trueCases := []string{"true", "TRUE", "1", "yes", "YES", " Yes "}
	for _, raw := range trueCases {
		got, err := parseIncludeCompleted(raw)
		if err != nil {
			t.Errorf("parseIncludeCompleted(%q): unexpected error: %v", raw, err)
		}
		if !got {
			t.Errorf("parseIncludeCompleted(%q) = false, want true", raw)
		}
	}

	falseCases := []string{"false", "FALSE", "0", "no", "NO", " no "}
	for _, raw := range falseCases {
		got, err := parseIncludeCompleted(raw)
		if err != nil {
			t.Errorf("parseIncludeCompleted(%q): unexpected error: %v", raw, err)
		}
		if got {
			t.Errorf("parseIncludeCompleted(%q) = true, want false", raw)
		}
	}

	emptyCases := []string{"", "   "}
	for _, raw := range emptyCases {
		got, err := parseIncludeCompleted(raw)
		if err != nil {
			t.Errorf("parseIncludeCompleted(%q): unexpected error: %v", raw, err)
		}
		if got {
			t.Errorf("parseIncludeCompleted(%q) = true, want false", raw)
		}
	}

	invalidCases := []string{"maybe", "2", "t", "f", "on", "off"}
	for _, raw := range invalidCases {
		_, err := parseIncludeCompleted(raw)
		if err == nil {
			t.Errorf("parseIncludeCompleted(%q): expected error", raw)
		} else if err.Error() != includeCompletedAllowedMsg {
			t.Errorf("parseIncludeCompleted(%q) error = %q, want %q", raw, err.Error(), includeCompletedAllowedMsg)
		}
	}
}
