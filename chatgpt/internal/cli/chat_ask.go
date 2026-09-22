// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source live

package cli

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"chatgpt-pp-cli/internal/cliutil"
	"chatgpt-pp-cli/internal/gptapi"
)

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type askResult struct {
	ConversationID  string `json:"conversation_id,omitempty"`
	MessageID       string `json:"message_id"`
	Model           string `json:"model"`
	Events          int    `json:"events"`
	Terminal        bool   `json:"message_stream_complete"`
	ResumeTokenSeen bool   `json:"resume_token_seen"`
	AssistantText   string `json:"assistant_text,omitempty"`
	Note            string `json:"note,omitempty"`
}

func newChatAskCmd(flags *rootFlags) *cobra.Command {
	var flagConversationID, flagModel, flagParent, flagEffort string
	var flagWait bool

	cmd := &cobra.Command{
		Use:   "ask [message]",
		Short: "EXPERIMENTAL: send a message (new chat or follow-up) and stream the response.",
		Long:  `Browser-gated experimental send surface. Implements the full protocol (sentinel requirements + sha3-512 proof-of-work, conduit prepare, POST /f/conversation SSE) and treats the typed message_stream_complete event as the only semantic terminal — [DONE] and connection close alone are never treated as completion. Turnstile: this CLI does not solve or bypass Turnstile; if the server requires a Turnstile token the command fails with an explicit error. Refresh your session import and retry, or use the web UI. Safety: only sends into the conversation you explicitly name with --conversation-id; omit it to start a new chat.`,
		Example: stringsTrimNl(`
  chatgpt-pp-cli ask "Summarize this thread in three bullets" --conversation-id 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --model gpt-5-6
  chatgpt-pp-cli ask "Reply with exactly: OK" --dry-run
`),
		Annotations: map[string]string{"mcp:read-only": "false", "pp:happy-args": "message=test;--dry-run"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "chat ask")
			}
			if len(args) < 1 || strings.TrimSpace(args[0]) == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("message text is required"))
			}
			if cliutil.IsAnyHarness() {
				return writeHarnessRefusal(cmd.OutOrStdout(), flags, "send a ChatGPT message")
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			gc, err := newGptClient(flags)
			if err != nil {
				return err
			}
			model := flagModel
			if model == "" {
				model = "gpt-5-6"
			}
			parent := flagParent
			if parent == "" && flagConversationID != "" {
				// Follow-up: parent is the conversation's current node.
				raw, err := gc.Get(ctx, "/backend-api/conversations/"+flagConversationID, nil)
				if err != nil {
					return fmt.Errorf("fetching conversation for parent: %w", err)
				}
				var j struct {
					CurrentNode string `json:"current_node"`
				}
				if err := json.Unmarshal(raw, &j); err != nil || j.CurrentNode == "" {
					return fmt.Errorf("resolving current_node for follow-up: %v", err)
				}
				parent = j.CurrentNode
			}
			if parent == "" {
				parent = newUUID() // new chat root
			}
			msgID := newUUID()
			body := map[string]any{
				"action": "next",
				"messages": []map[string]any{{
					"id":       msgID,
					"author":   map[string]string{"role": "user"},
					"content":  map[string]any{"content_type": "text", "parts": []string{args[0]}},
					"metadata": map[string]any{},
				}},
				"parent_message_id":    parent,
				"model":                model,
				"timezone_offset_min":  -330,
				"timezone":             "Asia/Calcutta",
				"conversation_mode":    map[string]string{"kind": "primary_assistant"},
				"system_hints":         []string{},
				"supports_buffering":   true,
				"supported_encodings":  []string{"gzip"},
				"client_prepare_state": "success",
				"thinking_effort":      "standard",
				"websocket_request_id": newUUID(),
			}
			if flagEffort != "" {
				body["thinking_effort"] = flagEffort
			}
			if flagConversationID != "" {
				body["conversation_id"] = flagConversationID
			}

			// 1) sentinel requirements (+PoW)
			req, proof, err := gc.Requirements(ctx)
			if err != nil {
				return fmt.Errorf("sentinel: %w", err)
			}
			// 2) conduit
			conduit, err := gc.ConduitPrepare(ctx)
			if err != nil {
				return fmt.Errorf("conduit: %w", err)
			}
			// 3) stream
			events, terminal, resume, err := gc.StreamSend(ctx, body, req.Token, proof, conduit)
			if err != nil {
				if strings.Contains(err.Error(), "403") {
					return fmt.Errorf("send rejected (403): %w — likely Turnstile-gated; this CLI never bypasses Turnstile. Refresh auth ('auth import') and retry, or use the web UI", err)
				}
				return err
			}
			res := askResult{
				MessageID:       msgID,
				Model:           model,
				Events:          len(events),
				Terminal:        terminal,
				ResumeTokenSeen: resume != "",
			}
			for _, ev := range events {
				var e struct {
					Message *struct {
						ID      string                `json:"id"`
						Author  struct{ Role string } `json:"author"`
						Content struct {
							ContentType string          `json:"content_type"`
							Parts       json.RawMessage `json:"parts"`
						} `json:"content"`
						EndTurn *bool `json:"end_turn"`
					} `json:"message"`
					ConversationID string `json:"conversation_id"`
				}
				if json.Unmarshal(ev.Raw, &e) != nil {
					continue
				}
				if e.ConversationID != "" {
					res.ConversationID = e.ConversationID
				}
				if e.Message != nil && e.Message.Author.Role == "assistant" && e.Message.Content.ContentType == "text" &&
					e.Message.EndTurn != nil && *e.Message.EndTurn {
					var parts []string
					_ = json.Unmarshal(e.Message.Content.Parts, &parts)
					res.AssistantText = strings.Join(parts, "")
				}
			}
			if !terminal {
				res.Note = "stream ended without message_stream_complete — response may be incomplete; treat as in-progress and check 'chat status'"
			}
			if flagWait && res.ConversationID != "" {
				_ = waitForConversation(ctx, gc, res.ConversationID, 60*time.Second)
			}
			return printJSONFiltered(cmd.OutOrStdout(), res, flags)
		},
	}
	cmd.Flags().StringVar(&flagConversationID, "conversation-id", "", "send into this conversation (omit to start a new chat)")
	cmd.Flags().StringVar(&flagModel, "model", "", "model slug (see models list; default gpt-5-6)")
	cmd.Flags().StringVar(&flagParent, "parent", "", "parent message id (defaults to the conversation's current_node)")
	cmd.Flags().StringVar(&flagEffort, "thinking-effort", "", "low, standard, or high")
	cmd.Flags().BoolVar(&flagWait, "wait", false, "after sending, wait for the stream to fully complete")
	return cmd
}

// waitForConversation polls until the conversation's current_node message is
// finished (bounded by timeout).
func waitForConversation(ctx context.Context, gc *gptapi.Client, id string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		raw, err := gc.Get(ctx, "/backend-api/conversations/"+id, nil)
		if err == nil {
			var j struct {
				CurrentNode string `json:"current_node"`
				Messages    []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
				} `json:"messages"`
			}
			if json.Unmarshal(raw, &j) == nil {
				for _, m := range j.Messages {
					if m.ID == j.CurrentNode && m.Status == "finished_successfully" {
						return true
					}
				}
			}
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(2 * time.Second):
		}
	}
	return false
}

func newChatStatusCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [conversation-id]",
		Short: "EXPERIMENTAL: report a conversation's streaming state (stream_status + current node state).",
		Example: stringsTrimNl(`
  chatgpt-pp-cli chat status 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --json
`),
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "live", "pp:happy-args": "conversation-id=6aaa52c7-daec-83ee-9c8a-e1635dbf397e;--json"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "chat status")
			}
			if len(args) < 1 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("conversation-id is required"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			gc, err := newGptClient(flags)
			if err != nil {
				return err
			}
			raw, err := gc.Get(ctx, "/backend-api/conversations/"+args[0], nil)
			if err != nil {
				return fmt.Errorf("fetching conversation: %w", err)
			}
			var j struct {
				CurrentNode string `json:"current_node"`
				Messages    []struct {
					ID     string                `json:"id"`
					Status string                `json:"status"`
					Role   struct{ Role string } `json:"author"`
				} `json:"messages"`
			}
			if err := json.Unmarshal(raw, &j); err != nil {
				return fmt.Errorf("parsing conversation: %w", err)
			}
			state := "idle"
			currentStatus := ""
			for _, m := range j.Messages {
				if m.ID == j.CurrentNode {
					currentStatus = m.Status
				}
			}
			if currentStatus != "" && currentStatus != "finished_successfully" {
				state = "streaming:" + currentStatus
			}
			return printJSONFiltered(cmd.OutOrStdout(), map[string]any{
				"conversation_id":     args[0],
				"state":               state,
				"current_node":        j.CurrentNode,
				"current_node_status": currentStatus,
				"message_count":       len(j.Messages),
			}, flags)
		},
	}
	return cmd
}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		chatCmd, _, err := root.Find([]string{"chat"})
		if err == nil && chatCmd != nil {
			addNovelCommandIfAbsent(chatCmd, newChatAskCmd(flags))
			addNovelCommandIfAbsent(chatCmd, newChatStatusCmd(flags))
		}
	})
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		authCmd, _, err := root.Find([]string{"auth"})
		if err == nil && authCmd != nil {
			addNovelCommandIfAbsent(authCmd, newAuthImportCmd(flags))
		}
	})
}
