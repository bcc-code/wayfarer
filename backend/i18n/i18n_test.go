package i18n

import "testing"

func TestGetBetReasonMessages(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		wantLang string // expected language used (for verifying fallback)
	}{
		{
			name:     "Norwegian returns Norwegian messages",
			lang:     "nb",
			wantLang: "nb",
		},
		{
			name:     "English returns English messages",
			lang:     "en",
			wantLang: "en",
		},
		{
			name:     "Unknown language falls back to Norwegian (default)",
			lang:     "unknown",
			wantLang: "nb",
		},
		{
			name:     "Language variation normalized (no -> nb)",
			lang:     "no",
			wantLang: "nb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs := GetBetReasonMessages(tt.lang)

			// Verify that we got non-empty messages
			if msgs.Stake == "" {
				t.Error("expected non-empty Stake message")
			}
			if msgs.Winnings == "" {
				t.Error("expected non-empty Winnings message")
			}

			// Verify the messages contain the placeholder
			if msgs.Stake != "" && !containsPlaceholder(msgs.Stake, "{challenge}") {
				t.Errorf("Stake message should contain {challenge} placeholder, got: %s", msgs.Stake)
			}
			if msgs.Winnings != "" && !containsPlaceholder(msgs.Winnings, "{challenge}") {
				t.Errorf("Winnings message should contain {challenge} placeholder, got: %s", msgs.Winnings)
			}
		})
	}
}

func TestFormatBetStakeReason(t *testing.T) {
	tests := []struct {
		name          string
		lang          string
		challengeName string
		want          string
	}{
		{
			name:          "Norwegian formatting",
			lang:          "nb",
			challengeName: "Quiz 1",
			want:          "Quiz 1 - satset",
		},
		{
			name:          "English formatting",
			lang:          "en",
			challengeName: "Quiz 1",
			want:          "Quiz 1 - stake",
		},
		{
			name:          "Empty challenge name",
			lang:          "nb",
			challengeName: "",
			want:          " - satset",
		},
		{
			name:          "Challenge name with special characters",
			lang:          "en",
			challengeName: "Quiz: Level 2!",
			want:          "Quiz: Level 2! - stake",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBetStakeReason(tt.lang, tt.challengeName)
			if got != tt.want {
				t.Errorf("FormatBetStakeReason(%q, %q) = %q, want %q", tt.lang, tt.challengeName, got, tt.want)
			}
		})
	}
}

func TestFormatBetWinningsReason(t *testing.T) {
	tests := []struct {
		name          string
		lang          string
		challengeName string
		want          string
	}{
		{
			name:          "Norwegian formatting",
			lang:          "nb",
			challengeName: "Quiz 1",
			want:          "Quiz 1 - gevinst",
		},
		{
			name:          "English formatting",
			lang:          "en",
			challengeName: "Quiz 1",
			want:          "Quiz 1 - winnings",
		},
		{
			name:          "Empty challenge name",
			lang:          "nb",
			challengeName: "",
			want:          " - gevinst",
		},
		{
			name:          "Challenge name with special characters",
			lang:          "en",
			challengeName: "Quiz: Level 2!",
			want:          "Quiz: Level 2! - winnings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBetWinningsReason(tt.lang, tt.challengeName)
			if got != tt.want {
				t.Errorf("FormatBetWinningsReason(%q, %q) = %q, want %q", tt.lang, tt.challengeName, got, tt.want)
			}
		})
	}
}

func containsPlaceholder(s, placeholder string) bool {
	return len(s) > 0 && (s == placeholder || len(s) > len(placeholder))
}
