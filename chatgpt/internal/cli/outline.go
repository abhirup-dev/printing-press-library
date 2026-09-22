// Copyright 2026 dev-abhirup-sc and contributors. Licensed under Apache-2.0. See LICENSE.
// pp:data-source auto

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"chatgpt-pp-cli/internal/gptconv"
)

type outlineView struct {
	ConversationID string            `json:"conversation_id"`
	Title          string            `json:"title"`
	SectionCount   int               `json:"section_count"`
	TotalMessages  int               `json:"total_message_count"`
	Sections       []gptconv.Section `json:"sections"`
}

func newNovelOutlineCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outline [conversation-id]",
		Short: "Map a long thread into sections at each of your prompts, with anchors and timestamps.",
		Long:  `Derives the conversation outline the same way ChatGPT's own table of contents does: one section per user message, in chronological order. Each section carries the verbatim prompt excerpt, the anchor message id (usable with 'transcript --from-anchor/--to-anchor'), timestamps, and the message range. Use this command to map a long thread's structure. Do NOT use it for full content; use 'transcript' instead.`,
		Example: strings.Trim(`
  chatgpt-pp-cli outline 6aaa52c7-daec-83ee-9c8a-e1635dbf397e
  chatgpt-pp-cli outline 6aaa52c7-daec-83ee-9c8a-e1635dbf397e --json --select sections.prompt_excerpt
`, "\n"),
		Annotations: map[string]string{"mcp:read-only": "true", "pp:data-source": "auto", "pp:novel-scaffold": "true", "pp:happy-args": "conversation-id=6aaa52c7-daec-83ee-9c8a-e1635dbf397e;--json"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "outline")
			}
			if len(args) < 1 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("conversation-id is required"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			conv, err := fetchConversationComplete(ctx, flags, args[0])
			if err != nil {
				return err
			}
			view := outlineView{
				ConversationID: conv.ConversationID,
				Title:          conv.Title,
				TotalMessages:  len(conv.Messages),
				Sections:       gptconv.BuildOutline(conv.Messages),
			}
			view.SectionCount = len(view.Sections)
			if !wantsHumanTable(cmd.OutOrStdout(), flags) {
				return printJSONFiltered(cmd.OutOrStdout(), view, flags)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s — %d sections, %d messages\n\n", conv.Title, view.SectionCount, view.TotalMessages)
			for _, s := range view.Sections {
				fmt.Fprintf(out, "%2d. %s\n    %s → %s  (%d msgs, %d replies)  anchor %s\n",
					s.Index, s.PromptExcerpt, s.CreateTime, "", s.MessageCount, s.AssistantReplies, shortID(s.AnchorMessageID))
			}
			return nil
		},
	}
	return cmd
}

// fetchConversationComplete fetches a conversation and completes it through
// the forward cursor when the primary response is truncated.
func fetchConversationComplete(ctx context.Context, flags *rootFlags, id string) (*gptconv.Conversation, error) {
	gc, err := newGptClient(flags)
	if err != nil {
		return nil, err
	}
	raw, err := gc.Get(ctx, "/backend-api/conversations/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching conversation: %w", err)
	}
	var conv gptconv.Conversation
	if err := json.Unmarshal(raw, &conv); err != nil {
		return nil, fmt.Errorf("parsing conversation: %w", err)
	}
	// Completion loop: the primary response can be truncated while claiming
	// has_next_page=false. Walk the forward cursor until the tail matches
	// current_node (or stops growing).
	for i := 0; i < 100; i++ {
		if conv.CurrentNode != "" && len(conv.Messages) > 0 {
			last := conv.Messages[len(conv.Messages)-1]
			if last.ID == conv.CurrentNode {
				break
			}
		} else if len(conv.Messages) == 0 {
			break
		}
		lastID := conv.Messages[len(conv.Messages)-1].ID
		q := url.Values{}
		q.Set("after", lastID)
		raw2, err := gc.Get(ctx, "/backend-api/conversations/"+id+"/messages", q)
		if err != nil {
			return nil, fmt.Errorf("completing conversation: %w", err)
		}
		var page struct {
			Messages []gptconv.Message `json:"messages"`
		}
		if err := json.Unmarshal(raw2, &page); err != nil {
			return nil, fmt.Errorf("parsing messages page: %w", err)
		}
		if len(page.Messages) == 0 {
			break
		}
		conv.Messages = append(conv.Messages, page.Messages...)
	}
	return &conv, nil
}
