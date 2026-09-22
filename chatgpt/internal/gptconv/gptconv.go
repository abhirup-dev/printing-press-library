// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.

// Package gptconv models ChatGPT conversations: normalized timestamps across
// the API's mixed formats, user-anchor outlines (the same derivation the web
// UI's TableOfContentsSidebar performs client-side), section slicing,
// hydration statistics, transcript rendering, and completeness verification.
package gptconv

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Message mirrors the messages[] item of GET /backend-api/conversations/{id}.
type Message struct {
	ID         string          `json:"id"`
	Author     Author          `json:"author"`
	CreateTime json.Number     `json:"create_time"`
	UpdateTime *json.Number    `json:"update_time,omitempty"`
	Content    Content         `json:"content"`
	Status     string          `json:"status"`
	EndTurn    *bool           `json:"end_turn"`
	Weight     json.Number     `json:"weight"`
	Recipient  string          `json:"recipient"`
	Metadata   MessageMetadata `json:"metadata"`
}

type Author struct {
	Role string `json:"role"`
	Name string `json:"name"`
}

type Content struct {
	ContentType string          `json:"content_type"`
	Parts       json.RawMessage `json:"parts"`
}

type MessageMetadata struct {
	RequestID                        string `json:"request_id"`
	TurnID                           string `json:"turn_id"`
	ModelSlug                        string `json:"model_slug"`
	MessageType                      string `json:"message_type"`
	IsVisuallyHiddenFromConversation *bool  `json:"is_visually_hidden_from_conversation"`
}

// Text joins string parts (multimodal objects render as [image]/[audio] tags).
func (c Content) Text() string {
	var parts []json.RawMessage
	if err := json.Unmarshal(c.Parts, &parts); err != nil {
		return ""
	}
	var sb strings.Builder
	for _, p := range parts {
		var s string
		if err := json.Unmarshal(p, &s); err == nil {
			sb.WriteString(s)
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal(p, &obj); err == nil {
			if ct, _ := obj["content_type"].(string); ct != "" {
				fmt.Fprintf(&sb, "[%s]", ct)
				continue
			}
		}
		sb.WriteString("[non-text]")
	}
	return sb.String()
}

// Conversation is the detail response.
type Conversation struct {
	Title            string      `json:"title"`
	CreateTime       json.Number `json:"create_time"`
	UpdateTime       json.Number `json:"update_time"`
	ConversationID   string      `json:"conversation_id"`
	DefaultModelSlug string      `json:"default_model_slug"`
	Messages         []Message   `json:"messages"`
	CurrentNode      string      `json:"current_node"`
	PageInfo         *PageInfo   `json:"page_info"`
}

type PageInfo struct {
	StartCursor     string `json:"start_cursor"`
	EndCursor       string `json:"end_cursor"`
	HasPreviousPage bool   `json:"has_previous_page"`
	HasNextPage     bool   `json:"has_next_page"`
}

// Time normalizes create_time (unix float) to time.Time. Zero on absence.
func Time(n json.Number) time.Time {
	if n == "" {
		return time.Time{}
	}
	f, err := n.Float64()
	if err != nil {
		return time.Time{}
	}
	sec := int64(f)
	return time.Unix(sec, int64((f-float64(sec))*1e9)).UTC()
}

// ISO renders a normalized timestamp; "" stays "".
func ISO(n json.Number) string {
	t := Time(n)
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

// ParseISOTimestamp parses list-item timestamps (ISO-8601 strings) or unix
// floats — the two formats the API mixes between list and detail responses.
func ParseISOTimestamp(s string, n *json.Number) time.Time {
	if n != nil && *n != "" {
		if t := Time(*n); !t.IsZero() {
			return t
		}
	}
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// Section is one outline entry — one per user message, matching the web UI's
// TOC: label = verbatim prompt excerpt, anchor = user message id.
type Section struct {
	Index            int         `json:"index"`
	AnchorMessageID  string      `json:"anchor_message_id"`
	PromptExcerpt    string      `json:"prompt_excerpt"`
	CreateTime       string      `json:"create_time"`
	CreateTimeRaw    json.Number `json:"create_time_raw"`
	MessageCount     int         `json:"message_count"` // messages from this anchor to the next
	FirstMessageID   string      `json:"first_message_id"`
	LastMessageID    string      `json:"last_message_id"`
	AssistantReplies int         `json:"assistant_replies"`
}

// ExcerptLimit is the TOC label truncation length (word boundary).
const ExcerptLimit = 80

// Excerpt renders a verbatim, single-line prompt excerpt.
func Excerpt(text string, limit int) string {
	line := strings.TrimSpace(strings.SplitN(text, "\n", 2)[0])
	if limit <= 0 {
		limit = ExcerptLimit
	}
	runes := []rune(line)
	if len(runes) <= limit {
		return line
	}
	cut := string(runes[:limit])
	if i := strings.LastIndexAny(cut, " \t"); i > limit/2 {
		cut = cut[:i]
	}
	return cut + "…"
}

// BuildOutline derives sections from user messages in chronological order.
// Hidden user messages (is_visually_hidden_from_conversation) are skipped,
// mirroring the web TOC. Requires messages sorted ascending by create_time.
func BuildOutline(msgs []Message) []Section {
	sorted := make([]Message, len(msgs))
	copy(sorted, msgs)
	sort.SliceStable(sorted, func(i, j int) bool {
		ti, tj := Time(sorted[i].CreateTime), Time(sorted[j].CreateTime)
		if ti.Equal(tj) {
			return sorted[i].ID < sorted[j].ID
		}
		return ti.Before(tj)
	})
	var anchors []int
	for i, m := range sorted {
		if m.Author.Role != "user" {
			continue
		}
		if m.Metadata.IsVisuallyHiddenFromConversation != nil && *m.Metadata.IsVisuallyHiddenFromConversation {
			continue
		}
		anchors = append(anchors, i)
	}
	sections := make([]Section, 0, len(anchors))
	for si, ai := range anchors {
		end := len(sorted)
		if si+1 < len(anchors) {
			end = anchors[si+1]
		}
		span := sorted[ai:end]
		replies := 0
		for _, m := range span {
			if m.Author.Role == "assistant" && m.Content.ContentType == "text" && m.EndTurn != nil && *m.EndTurn {
				replies++
			}
		}
		m := sorted[ai]
		sections = append(sections, Section{
			Index:            si + 1,
			AnchorMessageID:  m.ID,
			PromptExcerpt:    Excerpt(m.Content.Text(), ExcerptLimit),
			CreateTime:       ISO(m.CreateTime),
			CreateTimeRaw:    m.CreateTime,
			MessageCount:     len(span),
			FirstMessageID:   m.ID,
			LastMessageID:    span[len(span)-1].ID,
			AssistantReplies: replies,
		})
	}
	return sections
}

// SliceBounds resolves a slice to [start,end) message indexes (ascending
// order) from --section N, --from-anchor id, --to-anchor id.
func SliceBounds(msgs []Message, section int, fromAnchor, toAnchor string) (int, int, error) {
	outline := BuildOutline(msgs)
	if section > 0 {
		if section > len(outline) {
			return 0, 0, fmt.Errorf("section %d out of range: outline has %d sections", section, len(outline))
		}
		start := indexOfMessage(msgs, outline[section-1].AnchorMessageID)
		end := len(msgs)
		if section < len(outline) {
			end = indexOfMessage(msgs, outline[section].AnchorMessageID)
		}
		if start < 0 || end < start {
			return 0, 0, fmt.Errorf("section %d anchors not present in message list", section)
		}
		return start, end, nil
	}
	start, end := 0, len(msgs)
	if fromAnchor != "" {
		if i := indexOfMessage(msgs, fromAnchor); i >= 0 {
			start = i
		} else {
			return 0, 0, fmt.Errorf("--from-anchor %s not found in conversation", fromAnchor)
		}
	}
	if toAnchor != "" {
		if i := indexOfMessage(msgs, toAnchor); i > start {
			end = i + 1
		} else {
			return 0, 0, fmt.Errorf("--to-anchor %s not found after --from-anchor", toAnchor)
		}
	}
	return start, end, nil
}

func indexOfMessage(msgs []Message, id string) int {
	for i, m := range msgs {
		if m.ID == id {
			return i
		}
	}
	return -1
}

// Stats are the per-conversation hydration counts the list API omits.
type Stats struct {
	ConversationID     string `json:"conversation_id"`
	UserTurnCount      int    `json:"user_turn_count"`
	AssistantTurnCount int    `json:"assistant_turn_count"` // end_turn assistant messages
	AssistantMessages  int    `json:"assistant_messages"`   // all assistant nodes
	ToolMessages       int    `json:"tool_messages"`
	SystemMessages     int    `json:"system_messages"`
	HiddenMessages     int    `json:"hidden_messages"`
	TotalMessages      int    `json:"total_message_count"`
	SpanSeconds        int64  `json:"span_seconds"`
	ContentBytes       int    `json:"content_bytes"`
	FirstTime          string `json:"first_time"`
	LastTime           string `json:"last_time"`
}

// ComputeStats derives hydration stats from a conversation body.
func ComputeStats(conv *Conversation) Stats {
	s := Stats{ConversationID: conv.ConversationID}
	var first, last time.Time
	for _, m := range conv.Messages {
		s.TotalMessages++
		switch m.Author.Role {
		case "user":
			if m.Metadata.IsVisuallyHiddenFromConversation == nil || !*m.Metadata.IsVisuallyHiddenFromConversation {
				s.UserTurnCount++
			}
		case "assistant":
			s.AssistantMessages++
			if m.EndTurn != nil && *m.EndTurn {
				s.AssistantTurnCount++
			}
		case "tool":
			s.ToolMessages++
		case "system":
			s.SystemMessages++
		}
		if m.Metadata.IsVisuallyHiddenFromConversation != nil && *m.Metadata.IsVisuallyHiddenFromConversation {
			s.HiddenMessages++
		}
		if t := Time(m.CreateTime); !t.IsZero() {
			if first.IsZero() || t.Before(first) {
				first = t
			}
			if t.After(last) {
				last = t
			}
		}
	}
	if !first.IsZero() && !last.IsZero() {
		s.SpanSeconds = int64(last.Sub(first).Seconds())
		s.FirstTime = first.Format(time.RFC3339)
		s.LastTime = last.Format(time.RFC3339)
	}
	if b, err := json.Marshal(conv.Messages); err == nil {
		s.ContentBytes = len(b)
	}
	return s
}

// Verification reports completeness checks for extraction.
type Verification struct {
	Complete            bool           `json:"complete"`
	LastMessageID       string         `json:"last_message_id"`
	CurrentNode         string         `json:"current_node"`
	TailMatches         bool           `json:"tail_matches_current_node"`
	NonFinishedMessages []string       `json:"non_finished_messages"`
	CountByRole         map[string]int `json:"count_by_role"`
	Notes               []string       `json:"notes,omitempty"`
}

// Verify checks the extracted message list against the conversation's
// current_node and statuses.
func Verify(conv *Conversation) Verification {
	v := Verification{
		CurrentNode:   conv.CurrentNode,
		CountByRole:   map[string]int{},
		LastMessageID: "",
	}
	msgs := conv.Messages
	if len(msgs) > 0 {
		v.LastMessageID = msgs[len(msgs)-1].ID
	}
	v.TailMatches = v.LastMessageID != "" && v.LastMessageID == conv.CurrentNode
	for _, m := range msgs {
		v.CountByRole[m.Author.Role]++
		if m.Status != "" && m.Status != "finished_successfully" {
			v.NonFinishedMessages = append(v.NonFinishedMessages, fmt.Sprintf("%s(%s,%s)", m.ID, m.Author.Role, m.Status))
		}
	}
	v.NonFinishedMessages = orEmpty(v.NonFinishedMessages)
	v.Complete = v.TailMatches && len(v.NonFinishedMessages) == 0
	if !v.TailMatches {
		v.Notes = append(v.Notes, "last message does not match current_node — extraction may be truncated; re-run with cursor completion or --full-refresh")
	}
	if len(v.NonFinishedMessages) > 0 {
		v.Notes = append(v.Notes, fmt.Sprintf("%d messages not in finished_successfully state", len(v.NonFinishedMessages)))
	}
	v.Notes = orEmpty(v.Notes)
	return v
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// RenderOptions control transcript markdown rendering.
type RenderOptions struct {
	ShowReasoning bool
	ShowTools     bool
	ShowHidden    bool
	ShowCode      bool // assistant code content_type blocks (default: inline in text)
}

// RenderMarkdown renders the conversation (or slice) as readable markdown with
// timestamps, roles, and fenced code blocks.
func RenderMarkdown(conv *Conversation, msgs []Message, opts RenderOptions) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", conv.Title)
	fmt.Fprintf(&sb, "- Conversation ID: `%s`\n", conv.ConversationID)
	if t := ISO(conv.CreateTime); t != "" {
		fmt.Fprintf(&sb, "- Created: %s\n", t)
	}
	if t := ISO(conv.UpdateTime); t != "" {
		fmt.Fprintf(&sb, "- Updated: %s\n", t)
	}
	if conv.DefaultModelSlug != "" {
		fmt.Fprintf(&sb, "- Default model: `%s`\n", conv.DefaultModelSlug)
	}
	fmt.Fprintf(&sb, "- Messages: %d\n\n---\n\n", len(msgs))
	for _, m := range msgs {
		hidden := m.Metadata.IsVisuallyHiddenFromConversation != nil && *m.Metadata.IsVisuallyHiddenFromConversation
		switch m.Author.Role {
		case "system":
			if !opts.ShowHidden {
				continue
			}
			fmt.Fprintf(&sb, "### [system%s] %s\n\n```\n%s\n```\n\n", markHidden(hidden), ISO(m.CreateTime), m.Content.Text())
			continue
		case "tool":
			if !opts.ShowTools {
				continue
			}
			fmt.Fprintf(&sb, "### [tool%s → %s] %s\n\n```\n%s\n```\n\n", markHidden(hidden), m.Recipient, ISO(m.CreateTime), truncate(m.Content.Text(), 2000))
			continue
		case "assistant":
			switch m.Content.ContentType {
			case "thoughts", "reasoning_recap":
				if !opts.ShowReasoning {
					continue
				}
				fmt.Fprintf(&sb, "> **reasoning** (%s, %s)\n> %s\n\n", m.Content.ContentType, ISO(m.CreateTime), strings.ReplaceAll(truncate(m.Content.Text(), 2000), "\n", "\n> "))
				continue
			case "model_editable_context":
				if !opts.ShowHidden {
					continue
				}
				fmt.Fprintf(&sb, "### [context%s] %s\n\n```\n%s\n```\n\n", markHidden(hidden), ISO(m.CreateTime), truncate(m.Content.Text(), 1000))
				continue
			}
			fmt.Fprintf(&sb, "## 🤖 Assistant — %s\n\n%s\n\n", ISO(m.CreateTime), m.Content.Text())
			if m.Metadata.ModelSlug != "" {
				fmt.Fprintf(&sb, "*model: `%s`*\n\n", m.Metadata.ModelSlug)
			}
		case "user":
			if hidden && !opts.ShowHidden {
				continue
			}
			fmt.Fprintf(&sb, "## 👤 You — %s\n\n%s\n\n", ISO(m.CreateTime), m.Content.Text())
		default:
			fmt.Fprintf(&sb, "### [%s%s] %s\n\n%s\n\n", m.Author.Role, markHidden(hidden), ISO(m.CreateTime), truncate(m.Content.Text(), 2000))
		}
	}
	return sb.String()
}

func markHidden(hidden bool) string {
	if hidden {
		return "/hidden"
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
