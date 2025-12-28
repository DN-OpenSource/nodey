package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
)

type AnalystResponse struct {
	Status    string   `json:"status"` // "valid", "needs_info", "invalid"
	Reason    string   `json:"reason"`
	Questions []string `json:"questions"`
	Summary   string   `json:"summary"` // Refined understanding of the request
}

// AnalyzeRequest uses the OpenAI API to evaluate the user's input.
func AnalyzeRequest(client *openai.Client, input string, history []string) (AnalystResponse, error) {
	// Construct the prompt
	sysPrompt := `You are an expert Requirements Analyst for a Flowchart Builder.
Your job is to analyze the user's request and determine if it's sufficient to build a flowchart.

Return a JSON object with:
- "status": "valid" (ready to build), "needs_info" (ambiguous/incomplete), or "invalid" (nonsense/unrelated).
- "reason": A short explanation of your decision.
- "questions": A list of 1-3 specific questions if status is "needs_info". Empty otherwise.
- "summary": A professional summary of the requirements so far.

Example:
Input: "Order flow"
Response: {"status": "needs_info", "reason": "Too vague", "questions": ["What triggers the order?", "Are there approval steps?"], "summary": "User wants an order process."}
`

	// Build messages including history if needed (simplified for now)
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(sysPrompt),
		openai.UserMessage(fmt.Sprintf("User Input: %s", input)),
	}

	// Make the call
	res, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModelGPT5Nano2025_08_07,
	})
	if err != nil {
		return AnalystResponse{}, err
	}

	content := res.Choices[0].Message.Content
	var response AnalystResponse
	if err := json.Unmarshal([]byte(cleanJSON(content)), &response); err != nil {
		return AnalystResponse{}, fmt.Errorf("failed to parse JSON: %v, content: %s", err, content)
	}

	return response, nil
}
