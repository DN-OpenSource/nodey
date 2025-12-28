package agents

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
)

// Research simulates a web search or knowledge expansion step.
func Research(client *openai.Client, topic string, history []string) (string, error) {
	// In a real app, this would call a search API.
	// Here, we use the LLM to "hallucinate" accurate info based on its training data,
	// formatted as if it found search results.

	sysPrompt := `You are an expert Researcher. 
The user needs detailed information about a topic to build a flowchart.
Provide a comprehensive summary of the steps, edge cases, and best practices for the requested process.
Format it as a clear research report.`

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(sysPrompt),
		openai.UserMessage(fmt.Sprintf("Research Topic: %s", topic)),
	}

	res, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    openai.ChatModelGPT5Nano2025_08_07,
	})
	if err != nil {
		return "", err
	}

	return res.Choices[0].Message.Content, nil
}
