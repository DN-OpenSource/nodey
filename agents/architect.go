package agents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
)

type Flowchart struct {
	Overview    Overview     `json:"overview"`
	Nodes       []Node       `json:"nodes"`
	Connections []Connection `json:"connections"`
}

type Overview struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type Node struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // trigger, action, decision, end
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Title string `json:"title"`
	Notes string `json:"notes"` // Technical logic details
}

type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // out, yes, no
}

// GenerateFlowchart creates or updates a flowchart.
func GenerateFlowchart(client *openai.Client, requirements string, research string, currentFlow *Flowchart) (Flowchart, error) {
	sysPrompt := `You are a Flow Architect. Generate or Modify a JSON flowchart based on the requirements.
Rules:
1. Coordinates: Start at (100, 300). Vertical or Horizontal flow. Avoid overlapping.
2. Nodes: Must have unique IDs. Types: "start", "trigger", "action", "decision", "end".
   - "start": The entry point of the flow.
   - "trigger": The event that initiates a process logic.
   - "decision": Branching point (Requires 'yes' and 'no' connections).
   - "action": A process step.
   - "end": The final step.
3. Content: 
   - title: Short display name (e.g. "User Clicks").
   - notes: Technical details (e.g. "API call to /v1/auth").
4. Connections: valid "from" and "to" IDs.

If an Existing Flow is provided, MODIFY it to meet the new requirements. Do not start over unless asked.
Preserve existing IDs if possible.

Example Output Structure:
{
  "overview": {"title": "Example Flow", "summary": "A simple flow"},
  "nodes": [
     {"id": "1", "type": "start", "x": 100, "y": 300, "title": "Start", "notes": "Entry point"},
     {"id": "2", "type": "action", "x": 400, "y": 300, "title": "Process", "notes": "..."}
  ],
  "connections": [
     {"from": "1", "to": "2", "type": "out"}
  ]
}

CRITICAL: You MUST generate at least 2 nodes. Return strictly JSON.`

	input := fmt.Sprintf("Requirements: %s\n\nResearch: %s", requirements, research)
	if currentFlow != nil {
		currentJSON, _ := json.MarshalIndent(currentFlow, "", "  ")
		input += fmt.Sprintf("\n\nExisting Flowchart to Modify:\n%s", string(currentJSON))
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(sysPrompt),
		openai.UserMessage(input),
	}

	res, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModelGPT5Nano2025_08_07,
	})
	if err != nil {
		return Flowchart{}, err
	}

	var flow Flowchart
	if err := json.Unmarshal([]byte(cleanJSON(res.Choices[0].Message.Content)), &flow); err != nil {
		return Flowchart{}, err
	}
	if len(flow.Nodes) == 0 {
		return Flowchart{}, fmt.Errorf("generated flowchart has 0 nodes, likely invalid JSON or AI refusal")
	}
	return flow, nil
}
