// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

package gptconv

import (
	"encoding/json"
	"testing"
	"time"
)

func msg(id, role string, ct float64, text, ctype string, endTurn *bool, hidden *bool) Message {
	m := Message{
		ID:         id,
		Author:     Author{Role: role},
		CreateTime: json.Number(jsonFloat(ct)),
		Content:    Content{ContentType: ctype, Parts: json.RawMessage(`["` + text + `"]`)},
		Status:     "finished_successfully",
	}
	m.Metadata.IsVisuallyHiddenFromConversation = hidden
	if role == "assistant" {
		m.EndTurn = endTurn
	}
	return m
}

func jsonFloat(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

func tp(b bool) *bool { return &b }

var fixture = []Message{
	msg("aaa1", "system", 100.0, "", "text", nil, tp(true)),
	msg("uuu1", "user", 100.5, "Hey, so I'm in the market for a new phone and want a really good camera this time around", "text", nil, nil),
	msg("sss1", "assistant", 101.0, "context", "model_editable_context", nil, nil),
	msg("ttt1", "assistant", 101.2, "thinking hard about phones", "thoughts", tp(false), nil),
	msg("rrr1", "assistant", 102.0, "Great question — here are the best camera phones.", "text", tp(true), nil),
	msg("ttool", "tool", 102.5, "web.run result", "text", nil, nil),
	msg("uuu2", "user", 200.0, "What about things like haptics and battery life", "text", nil, nil),
	msg("ttt2", "assistant", 200.5, "considering haptics", "reasoning_recap", tp(false), nil),
	msg("rrr2", "assistant", 201.0, "Haptics: XYZ is best.", "text", tp(true), nil),
}

func TestBuildOutline_UserAnchorsOnly(t *testing.T) {
	sections := BuildOutline(fixture)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections (one per visible user message), got %d", len(sections))
	}
	if sections[0].AnchorMessageID != "uuu1" || sections[1].AnchorMessageID != "uuu2" {
		t.Fatalf("anchors wrong: %+v", sections)
	}
	if sections[0].Index != 1 || sections[1].Index != 2 {
		t.Fatalf("1-based indexes wrong: %+v", sections)
	}
	if sections[0].MessageCount != 5 { // uuu1..ttool
		t.Fatalf("section 1 span wrong: %d", sections[0].MessageCount)
	}
	if sections[1].MessageCount != 3 { // uuu2..rrr2
		t.Fatalf("section 2 span wrong: %d", sections[1].MessageCount)
	}
	if sections[0].AssistantReplies != 1 || sections[1].AssistantReplies != 1 {
		t.Fatalf("assistant replies wrong: %+v", sections)
	}
	want := "Hey, so I'm in the market for a new phone and want a really good camera this…"
	if sections[0].PromptExcerpt != want {
		t.Fatalf("excerpt = %q, want %q", sections[0].PromptExcerpt, want)
	}
}

func TestBuildOutline_HiddenUserSkipped(t *testing.T) {
	withHidden := append(append([]Message{}, fixture...), msg("uhid", "user", 300.0, "hidden prompt", "text", nil, tp(true)))
	if got := len(BuildOutline(withHidden)); got != 2 {
		t.Fatalf("hidden user message must not create a section; got %d", got)
	}
}

func TestSliceBounds_Section(t *testing.T) {
	s, e, err := SliceBounds(fixture, 1, "", "")
	if err != nil || s != 1 || e != 6 {
		t.Fatalf("section 1 slice = [%d,%d) err=%v", s, e, err)
	}
	s, e, err = SliceBounds(fixture, 2, "", "")
	if err != nil || s != 6 || e != 9 {
		t.Fatalf("section 2 slice = [%d,%d) err=%v", s, e, err)
	}
	if _, _, err := SliceBounds(fixture, 3, "", ""); err == nil {
		t.Fatal("section 3 must be out of range")
	}
}

func TestSliceBounds_Anchors(t *testing.T) {
	s, e, err := SliceBounds(fixture, 0, "uuu1", "uuu2")
	if err != nil || s != 1 || e != 7 {
		t.Fatalf("anchor slice = [%d,%d) err=%v", s, e, err)
	}
	if _, _, err := SliceBounds(fixture, 0, "nope", ""); err == nil {
		t.Fatal("unknown from-anchor must error")
	}
}

func TestComputeStats(t *testing.T) {
	conv := &Conversation{ConversationID: "c1", Messages: fixture}
	st := ComputeStats(conv)
	if st.TotalMessages != len(fixture) || st.UserTurnCount != 2 || st.AssistantMessages != 5 ||
		st.AssistantTurnCount != 2 || st.ToolMessages != 1 || st.SystemMessages != 1 || st.HiddenMessages != 1 {
		t.Fatalf("stats wrong: %+v", st)
	}
	if st.SpanSeconds != 101 { // 100.0 → 201.0
		t.Fatalf("span = %d, want 101", st.SpanSeconds)
	}
	if st.ContentBytes == 0 {
		t.Fatal("content bytes must be non-zero")
	}
}

func TestVerify_TailAndStatus(t *testing.T) {
	conv := &Conversation{Messages: fixture, CurrentNode: "rrr2"}
	v := Verify(conv)
	if !v.Complete || !v.TailMatches {
		t.Fatalf("expected complete: %+v", v)
	}
	conv2 := &Conversation{Messages: fixture, CurrentNode: "other"}
	v2 := Verify(conv2)
	if v2.Complete || v2.TailMatches {
		t.Fatalf("mismatched tail must not verify complete: %+v", v2)
	}
	bad := fixture
	bad[len(bad)-1].Status = "in_progress"
	conv3 := &Conversation{Messages: bad, CurrentNode: "rrr2"}
	v3 := Verify(conv3)
	if v3.Complete || len(v3.NonFinishedMessages) != 1 {
		t.Fatalf("unfinished message must fail verify: %+v", v3)
	}
}

func TestParseISOTimestamp_MixedFormats(t *testing.T) {
	iso := ParseISOTimestamp("2026-09-16T08:26:49.075010Z", nil)
	if iso.IsZero() || iso.Year() != 2026 {
		t.Fatalf("ISO parse failed: %v", iso)
	}
	n := json.Number("1789547209.07501")
	unix := ParseISOTimestamp("", &n)
	if unix.IsZero() || unix.Unix() != 1789547209 {
		t.Fatalf("unix parse failed: %v", unix)
	}
	if d := iso.Sub(unix); d > time.Second || d < -time.Second {
		t.Fatalf("mixed formats should normalize to same instant: %v vs %v (%v)", iso, unix, d)
	}
}

func TestRenderMarkdown_DefaultsHideNoise(t *testing.T) {
	conv := &Conversation{Title: "T", ConversationID: "cid", DefaultModelSlug: "gpt-5-6", CurrentNode: "rrr2", Messages: fixture}
	md := RenderMarkdown(conv, fixture, RenderOptions{})
	for _, want := range []string{"# T", "## 👤 You", "## 🤖 Assistant", "Haptics: XYZ is best.", "model: `gpt-5-6`"} {
		if !contains(md, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
	for _, not := range []string{"thinking hard", "web.run result", "context", "hidden prompt"} {
		if contains(md, not) {
			t.Errorf("markdown should hide %q by default", not)
		}
	}
	mdAll := RenderMarkdown(conv, fixture, RenderOptions{ShowReasoning: true, ShowTools: true, ShowHidden: true})
	for _, want := range []string{"thinking hard", "web.run result", "**reasoning**"} {
		if !contains(mdAll, want) {
			t.Errorf("show-all markdown missing %q", want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

func TestContentText_Multimodal(t *testing.T) {
	c := Content{ContentType: "text", Parts: json.RawMessage(`["hello ",{"content_type":"image_asset_pointer"}]`)}
	if got := c.Text(); got != "hello [image_asset_pointer]" {
		t.Fatalf("multimodal text = %q", got)
	}
}
