package onboarding

import (
	"strings"
	"testing"
)

func TestMentionParsing(t *testing.T) {
	// Realistic raw text from Slack for "@katana-onboarder onboard Test User test.user@global.komatsu"
	raw := "<@U09ABCDEF> onboard Test User <mailto:test.user@global.komatsu|test.user@global.komatsu>"
	text := strings.TrimSpace(unwrapSlackEntities(raw))

	fields := strings.Fields(text)
	if len(fields) == 0 || strings.ToLower(fields[0]) != "onboard" {
		t.Fatalf("expected first token 'onboard', got text %q", text)
	}
	rest := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	engineers, err := parseEngineers(rest)
	if err != nil {
		t.Fatalf("parseEngineers(%q) error: %v", rest, err)
	}
	if len(engineers) != 1 || engineers[0].Email != "test.user@global.komatsu" {
		t.Fatalf("unexpected engineers: %+v", engineers)
	}
}

func TestMentionSlackAddParsing(t *testing.T) {
	raw := "<@U09ABCDEF> slack-add Test User <mailto:test.user@global.komatsu|test.user@global.komatsu> to <#C123|ext-program-katana>, #ext-program-katana-toolchain"
	text := strings.TrimSpace(unwrapSlackEntities(raw))

	fields := strings.Fields(text)
	if strings.ToLower(fields[0]) != "slack-add" {
		t.Fatalf("expected 'slack-add', got text %q", text)
	}
	rest := strings.TrimSpace(strings.TrimPrefix(text, fields[0]))
	users, channelsRaw, ok := splitOnTo(rest)
	if !ok {
		t.Fatalf("splitOnTo failed for %q", rest)
	}
	if _, err := parseEngineers(users); err != nil {
		t.Fatalf("parseEngineers(%q) error: %v", users, err)
	}
	channels := parseChannels(channelsRaw)
	if len(channels) != 2 || channels[0] != "#ext-program-katana" {
		t.Fatalf("unexpected channels: %+v", channels)
	}
}
