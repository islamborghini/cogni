package bench

import (
	"strings"
	"testing"
)

func TestParseStreamJSONResultEvent(t *testing.T) {
	stream := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{"usage":{"input_tokens":10,"output_tokens":5}}}`,
		`{"type":"assistant","message":{"usage":{"input_tokens":7,"output_tokens":3}}}`,
		`{"type":"result","subtype":"success","result":"final answer","usage":{"input_tokens":17,"output_tokens":8}}`,
	}, "\n")
	got, err := parseStreamJSON(strings.NewReader(stream))
	if err != nil {
		t.Fatal(err)
	}
	if got.Result != "final answer" {
		t.Errorf("Result=%q, want %q", got.Result, "final answer")
	}
	if got.InputTokens != 17 || got.OutputTokens != 8 {
		t.Errorf("tokens=%d/%d, want 17/8", got.InputTokens, got.OutputTokens)
	}
}

func TestParseStreamJSONFallbackSumsAssistantUsage(t *testing.T) {
	// No result event — runner was killed mid-stream. Should sum
	// per-assistant usage so we still record approximate cost.
	stream := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`{"type":"assistant","message":{"usage":{"input_tokens":10,"output_tokens":5}}}`,
		`{"type":"assistant","message":{"usage":{"input_tokens":7,"output_tokens":3}}}`,
	}, "\n")
	got, err := parseStreamJSON(strings.NewReader(stream))
	if err != nil {
		t.Fatal(err)
	}
	if got.Result != "" {
		t.Errorf("Result=%q, want empty (no result event)", got.Result)
	}
	if got.InputTokens != 17 || got.OutputTokens != 8 {
		t.Errorf("fallback tokens=%d/%d, want 17/8", got.InputTokens, got.OutputTokens)
	}
}

func TestParseStreamJSONSkipsMalformedLines(t *testing.T) {
	stream := strings.Join([]string{
		`{"type":"system","subtype":"init"}`,
		`not json at all`,
		``,
		`{"type":"result","result":"ok","usage":{"input_tokens":3,"output_tokens":2}}`,
	}, "\n")
	got, err := parseStreamJSON(strings.NewReader(stream))
	if err != nil {
		t.Fatal(err)
	}
	if got.Result != "ok" || got.InputTokens != 3 || got.OutputTokens != 2 {
		t.Errorf("got %+v, want Result=ok tokens=3/2", got)
	}
}
