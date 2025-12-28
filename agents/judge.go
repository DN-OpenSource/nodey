package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
)

type JudgeResponse struct {
	Approved bool   `json:"approved"`
	Critique string `json:"critique"`
	Dissent  string `json:"dissent"` // If approved=false, explain why
}

func Judgement(client *openai.Client, flowchartJSON string, requirements string) (JudgeResponse, error) {
	sysPrompt := `You are a panel of 3 Senior Software Architects acting as Judges.
Review the provided Flowchart JSON against the Requirements.
Vote on whether it is valid, complete, and technically sound.

If UNANIMOUS APPROVAL: return "approved": true.
If ANY DISAGREEMENT or MAJOR ISSUES: return "approved": false and provide "critique" and "dissent".

Critique should be constructive.
Ensure nodes are not overlapping (check coordinates).
Ensure logic flows correctly. It MUST have a 'start' and 'end' node.`

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(sysPrompt),
		openai.UserMessage(fmt.Sprintf("Requirements: %s\n\nFlowchart JSON: %s", requirements, flowchartJSON)),
	}

	res, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModelGPT5Nano2025_08_07,
	})
	if err != nil {
		return JudgeResponse{}, err
	}

	var resp JudgeResponse
	if err := json.Unmarshal([]byte(cleanJSON(res.Choices[0].Message.Content)), &resp); err != nil {
		// Fallback if strict JSON fails, try to parse manually or return error
		// For now, assuming GPT4o follows instructions well.
		return JudgeResponse{}, fmt.Errorf("failed to parse judge response: %v", err)
	}

	// Normalize strings
	resp.Critique = strings.TrimSpace(resp.Critique)

	return resp, nil
}
