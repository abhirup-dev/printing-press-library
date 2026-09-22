// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source auto

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"chatgpt-pp-cli/internal/gptconv"
)

type transcriptView struct {
	ConversationID string                `json:"conversation_id"`
	Title          string                `json:"title"`
	Created        string                `json:"created"`
	Updated        string                `json:"updated"`
	Model          string                `json:"default_model_slug,omitempty"`
	Slice          string                `json:"slice,omitempty"`
	MessageCount   int                   `json:"message_count"`
	Messages       []gptconv.Message     `json:"messages"`
	Verification   *gptconv.Verification `json:"verification,omitempty"`
}

func newNovelTranscriptCmd(flags *rootFlags) *cobra.Command {
	var flagSection int
	var flagFromAnchor, flagToAnchor, flagFormat, flagOut string
	var flagVerify, flagTree bool
	var flagShow []string

	cmd := &cobra.Command{
		Use:   "transcript [conversation-id]",
		Short: "Extract an entire conversation — complete, verified, JSON or readable markdown.",
		Long:  `Extracts a conversation completely: fetches the message array and completes it through the forward cursor until the tail matches current_node (the page_info flag can under-report truncation). --verify checks completeness and per-message statuses. Slice with --section N (outline section) or --from-anchor/--to-anchor message ids. Markdown output hides reasoning, tool output, and hidden system nodes unless requested via --show. Use 'transcript --tree' for the legacy mapping-tree shape. Do NOT use slicing flags for the whole conversation; run 'transcript <id>' bare.`,
		Example: strings.Trim(`
  chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --verify --json
  chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --section 3
  chatgpt--pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --format md -o thread.md
  chatgpt-pp-cli transcript 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --from-anchor bbb2162c --to-anchor bbb21702 --json
`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto", "pp:novel-scaffold": "true", "pp:happy-args": "conversation-id=6aaa52c7-daec-83ee-9c8a-e1635dbf397e;--json"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "transcript")
			}
			if len(args) < 1 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("conversation-id is required"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			show := parseShow(flagShow)
			if flagTree {
				return runTree(cmd, flags, ctx, args[0])
			}
			conv, err := fetchConversationComplete(ctx, flags, args[0])
			if err != nil {
				return err
			}
			msgs := conv.Messages
			sliceDesc := ""
			if flagSection > 0 || flagFromAnchor != "" || flagToAnchor != "" {
				s, e, err := gptconv.SliceBounds(msgs, flagSection, flagFromAnchor, flagToAnchor)
				if err != nil {
					_ = cmd.Usage()
					return usageErr(err)
				}
				msgs = msgs[s:e]
				sliceDesc = fmt.Sprintf("messages[%d:%d]", s, e)
			}
			view := transcriptView{
				ConversationID: conv.ConversationID,
				Title:          conv.Title,
				Created:        gptconv.ISO(conv.CreateTime),
				Updated:        gptconv.ISO(conv.UpdateTime),
				Model:          conv.DefaultModelSlug,
				Slice:          sliceDesc,
				MessageCount:   len(msgs),
				Messages:       msgs,
			}
			if flagVerify {
				v := gptconv.Verify(conv)
				view.Verification = &v
				if sliceDesc != "" {
					v.Notes = append(v.Notes, "verification covers the full conversation, not just the slice")
				}
			}
			if flagFormat == "md" {
				md := gptconv.RenderMarkdown(conv, msgs, show)
				if flagOut != "" {
					if err := os.WriteFile(flagOut, []byte(md), 0o644); err != nil {
						return fmt.Errorf("writing markdown: %w", err)
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s\n", flagOut)
					return nil
				}
				fmt.Fprint(cmd.OutOrStdout(), md)
				return nil
			}
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s — %d messages\n", conv.Title, len(msgs))
			for _, m := range msgs {
				if m.Author.Role == "assistant" && (m.Content.ContentType == "thoughts" || m.Content.ContentType == "reasoning_recap") && !show.ShowReasoning {
					continue
				}
				if m.Author.Role == "tool" && !show.ShowTools {
					continue
				}
				if m.Author.Role == "system" && !show.ShowHidden {
					continue
				}
				fmt.Fprintf(out, "[%s] %s: %s\n", gptconv.ISO(m.CreateTime), m.Author.Role, truncateStr(strings.ReplaceAll(m.Content.Text(), "\n", " ⏎ "), 100))
			}
			if view.Verification != nil {
				complete := "incomplete"
				if view.Verification.Complete {
					complete = "complete"
				}
				fmt.Fprintf(out, "\nverification: %s (tail_matches_current_node=%t)\n", complete, view.Verification.TailMatches)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&flagSection, "section", 0, "extract only this outline section (1-based; see 'outline')")
	cmd.Flags().StringVar(&flagFromAnchor, "from-anchor", "", "slice starting at this message id (inclusive)")
	cmd.Flags().StringVar(&flagToAnchor, "to-anchor", "", "slice ending at this message id (inclusive)")
	cmd.Flags().StringVar(&flagFormat, "format", "json", "json or md (readable markdown transcript)")
	cmd.Flags().StringVarP(&flagOut, "out", "o", "", "write markdown to this file instead of stdout")
	cmd.Flags().BoolVar(&flagVerify, "verify", false, "verify completeness against current_node and message statuses")
	cmd.Flags().BoolVar(&flagTree, "tree", false, "emit the legacy singular-endpoint mapping tree instead of the messages array")
	cmd.Flags().StringSliceVar(&flagShow, "show", nil, "extra content to include: reasoning, tools, hidden, code (csv)")
	return cmd
}

func parseShow(list []string) gptconv.RenderOptions {
	var o gptconv.RenderOptions
	for _, s := range list {
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "reasoning":
			o.ShowReasoning = true
		case "tools":
			o.ShowTools = true
		case "hidden":
			o.ShowHidden = true
		case "code":
			o.ShowCode = true
		}
	}
	return o
}

// runTree emits the singular-endpoint mapping tree (legacy compat read).
func runTree(cmd *cobra.Command, flags *rootFlags, ctx context.Context, id string) error {
	gc, err := newGptClient(flags)
	if err != nil {
		return err
	}
	raw, err := gc.Get(ctx, "/backend-api/conversation/"+id, nil)
	if err != nil {
		return fmt.Errorf("fetching mapping tree: %w", err)
	}
	var loose json.RawMessage = raw
	return printJSONFiltered(cmd.OutOrStdout(), loose, flags)
}
