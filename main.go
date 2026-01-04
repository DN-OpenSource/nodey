package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"github.com/dhirajnikam/nodey/agents"
	"github.com/dhirajnikam/nodey/generator"
)

// -- Styles --
var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Background(lipgloss.Color("#25A065")).Padding(0, 1)
	agentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	logStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D7D7D"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F44336"))
)

// -- States --
type state int

const (
	stateInput state = iota
	stateHistory
	stateAnalyzing
	stateAnswering
	stateResearching
	stateArchitecting
	stateJudging
	stateGenerating
	stateDone
)

// -- Messages --
type analysisMsg agents.AnalystResponse
type researchMsg string
type architectMsg agents.Flowchart
type judgeMsg agents.JudgeResponse
type generationMsg struct {
	err      error
	filename string
}
type errMsg struct{ err error }

// -- Model --
type model struct {
	client *openai.Client

	state   state
	spinner spinner.Model

	// Input
	textInput textarea.Model
	prompt    string
	history   []string // logs of what happened

	// History / File Loading
	files        []string
	cursor       int
	selectedFile string
	loadedFlow   *agents.Flowchart // The flow we are editing

	// Analysis
	questions       []string
	currentQIndex   int
	answers         []string
	analysisSummary string

	// Research
	researchData string

	// Architecture
	flowchart agents.Flowchart
	revision  int

	// Judgement
	critique string

	// Output
	finalPath string
	err       error
}

func newModel() model {
	ti := textarea.New()
	ti.Placeholder = "Describe the flow you need (e.g. 'User Login Process')...\nPress Ctrl+S to submit."
	ti.Focus()
	ti.CharLimit = 0 // No limit
	ti.SetHeight(3)
	ti.ShowLineNumbers = false

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Initialize OpenAI Client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println(errorStyle.Render("Error: OPENAI_API_KEY environment variable not set."))
		fmt.Println("Please run: export OPENAI_API_KEY='your-key-here'")
		os.Exit(1)
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return model{
		client:    &client,
		state:     stateInput,
		spinner:   s,
		textInput: ti,
		history:   []string{},
		revision:  0,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Allow quitting only if not waiting for API, or force quit
			return m, tea.Quit
		}

		// Global 'Open' handler if in Done state
		if m.state == stateDone && msg.String() == "o" {
			openFileInOS(m.finalPath)
			return m, nil
		}

		// Input Handling
		if m.state == stateInput {
			switch msg.String() {
			case "ctrl+s":
				m.prompt = m.textInput.Value()
				if strings.TrimSpace(m.prompt) == "" {
					return m, nil
				}
				m.textInput.Reset()
				m.history = append(m.history, "User: "+m.prompt)
				m.state = stateAnalyzing
				return m, tea.Batch(m.spinner.Tick, analyzeCmd(m.client, m.prompt, m.history))

			// Press Ctrl+L or some key to load history
			case "ctrl+l":
				files, err := getJSONFiles()
				if err == nil && len(files) > 0 {
					m.files = files
					m.state = stateHistory
					m.cursor = 0
				} else {
					m.history = append(m.history, "System: No saved flowcharts found.")
				}
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		// History Selection
		if m.state == stateHistory {
			switch msg.Type {
			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
				}
			case tea.KeyDown:
				if m.cursor < len(m.files)-1 {
					m.cursor++
				}
			case tea.KeyEnter:
				m.selectedFile = m.files[m.cursor]
				// Load it
				data, err := os.ReadFile(m.selectedFile)
				if err == nil {
					var flow agents.Flowchart
					if err := json.Unmarshal(data, &flow); err == nil {
						m.loadedFlow = &flow
						m.history = append(m.history, fmt.Sprintf("System: Loaded %s. Enter changes below:", m.selectedFile))
						m.state = stateInput
						m.textInput.Placeholder = "What changes do you want to make?\nPress Ctrl+S to submit."
						return m, nil
					}
				}
				m.state = stateInput
				m.history = append(m.history, "System: Failed to load file.")
				return m, nil

			case tea.KeyEsc:
				m.state = stateInput
				return m, nil

			// Open HTML from history
			case tea.KeyRunes:
				if msg.String() == "o" {
					m.selectedFile = m.files[m.cursor]
					// Convert .json to .html
					htmlPath := strings.Replace(m.selectedFile, ".json", ".html", 1)
					openFileInOS(htmlPath)
					return m, nil
				}
			}
			return m, nil
		}

		// Answering Questions Handling
		if m.state == stateAnswering {
			switch msg.String() {
			case "ctrl+s":
				answer := m.textInput.Value()
				if strings.TrimSpace(answer) == "" {
					return m, nil
				}
				m.history = append(m.history, fmt.Sprintf("Q: %s\nA: %s", m.questions[m.currentQIndex], answer))
				m.answers = append(m.answers, answer)
				m.textInput.Reset()

				m.currentQIndex++
				if m.currentQIndex >= len(m.questions) {
					// All answered
					m.state = stateResearching
					return m, tea.Batch(m.spinner.Tick, researchCmd(m.client, m.prompt+" "+strings.Join(m.answers, " "), m.history))
				}
				// Next question
				m.textInput.Placeholder = "Your answer..."
				return m, nil
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

	case spinner.TickMsg:
		if m.state != stateInput && m.state != stateAnswering && m.state != stateDone {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case analysisMsg:
		// Result from Analyst
		if msg.Status == "valid" {
			m.analysisSummary = msg.Summary
			m.history = append(m.history, "Analyst: Request is valid. "+msg.Summary)
			m.state = stateResearching
			return m, researchCmd(m.client, m.prompt, m.history) // Start research immediately
		} else if msg.Status == "needs_info" {
			m.history = append(m.history, fmt.Sprintf("Analyst: Need info - %s", msg.Reason))
			m.questions = msg.Questions
			m.currentQIndex = 0
			m.state = stateAnswering
			m.textInput.Placeholder = "Answer the question...\nPress Ctrl+S to submit."
			m.textInput.Focus()
			return m, textarea.Blink
		} else {
			m.state = stateInput // Go back to input
			m.err = fmt.Errorf("Analyst rejected: %s", msg.Reason)
			m.history = append(m.history, "Analyst: Rejected - "+msg.Reason)
			return m, nil
		}

	case researchMsg:
		m.researchData = string(msg)
		m.history = append(m.history, "Researcher: Found relevant patterns and data.")
		m.state = stateArchitecting
		// Pass loadedFlow if it exists
		return m, architectCmd(m.client, m.prompt+" "+m.analysisSummary, m.researchData, m.loadedFlow)

	case architectMsg:
		m.flowchart = agents.Flowchart(msg)
		m.history = append(m.history, fmt.Sprintf("Architect: Drafted flow with %d nodes.", len(m.flowchart.Nodes)))
		m.state = stateJudging
		return m, judgeCmd(m.client, m.flowchart, m.prompt)

	case judgeMsg:
		if msg.Approved {
			m.history = append(m.history, "Judges: Unanimous Approval.")
			m.state = stateGenerating
			return m, generateCmd(m.flowchart)
		}
		// Not approved
		m.revision++
		if m.revision > 3 {
			// Fail safe
			m.history = append(m.history, "Judges: Forced approval after max revisions.")
			m.state = stateGenerating
			return m, generateCmd(m.flowchart)
		}
		m.critique = msg.Critique + " " + msg.Dissent
		m.history = append(m.history, fmt.Sprintf("Judges: Critique - %s. Sending back to Architect.", msg.Critique))
		m.state = stateArchitecting
		// Re-architect with critique
		reqs := fmt.Sprintf("%s. Feedback: %s", m.prompt, m.critique)
		return m, architectCmd(m.client, reqs, m.researchData, m.loadedFlow)

	case generationMsg:
		if msg.err != nil {
			m.err = msg.err
			m.history = append(m.history, "Generator: Failed - "+msg.err.Error())
			m.state = stateInput
		} else {
			m.finalPath = msg.filename
			m.history = append(m.history, "Generator: Success! Saved to "+m.finalPath)
			m.state = stateDone
		}

	case errMsg:
		m.err = msg.err
		m.history = append(m.history, "Error: "+msg.err.Error())
		m.state = stateInput
	}

	return m, nil
}

func (m model) View() string {
	header := titleStyle.Render(" Nodey ") + "\n\n"

	// History Log
	logView := ""
	start := 0
	if len(m.history) > 10 {
		start = len(m.history) - 10
	}
	for _, entry := range m.history[start:] {
		logView += logStyle.Render("• "+entry) + "\n"
	}
	logView += "\n"

	// Content area
	content := ""
	switch m.state {
	case stateInput:
		prefix := ""
		if m.loadedFlow != nil {
			prefix = agentStyle.Render(fmt.Sprintf("[Editing %s]", m.selectedFile)) + "\n"
		}
		content = prefix + "Describe your desired flowchart (Ctrl+L to Load):\n" + m.textInput.View()
		if m.err != nil {
			content += "\n\n" + errorStyle.Render(m.err.Error())
		}

	case stateHistory:
		content = "Select a file to load:\n\n"
		for i, file := range m.files {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			content += fmt.Sprintf("%s %s\n", cursor, file)
		}

		// Footer for History
		content += "\n" + logStyle.Render("--------------------------------------")
		content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true).Render("[ o ] Open in Browser") + "   " + logStyle.Render("[ Enter ] Edit Flow") + "   " + logStyle.Render("[ Esc ] Cancel")

	case stateAnalyzing:
		content = fmt.Sprintf("%s Analyst is thinking...", m.spinner.View())

	case stateAnswering:
		q := m.questions[m.currentQIndex]
		prog := fmt.Sprintf("(%d/%d)", m.currentQIndex+1, len(m.questions))
		content = fmt.Sprintf("%s Agent needs clarification:\n\n%s %s\n\n%s", m.spinner.View(), agentStyle.Render(q), prog, m.textInput.View())

	case stateResearching:
		content = fmt.Sprintf("%s Researcher is gathering data...", m.spinner.View())

	case stateArchitecting:
		prefix := "Architect"
		if m.revision > 0 {
			prefix = fmt.Sprintf("Architect (Revision %d)", m.revision)
		}
		content = fmt.Sprintf("%s %s is designing the layout...", m.spinner.View(), prefix)

	case stateJudging:
		content = fmt.Sprintf("%s Judges are reviewing the draft...", m.spinner.View())

	case stateGenerating:
		content = fmt.Sprintf("%s Generating HTML artifact...", m.spinner.View())

	case stateDone:
		cwd, _ := os.Getwd()
		absPath := filepath.Join(cwd, m.finalPath)
		link := fmt.Sprintf("file://%s", absPath)
		content = fmt.Sprintf("%s\n\nFile saved to:\n%s\n\n%s\n%s",
			agentStyle.Render("Process Complete!"),
			link,
			lipgloss.NewStyle().Bold(true).Render("Press 'o' to open in browser"),
			logStyle.Render("(or Cmd+Click the link above)"),
		)
	}

	// Fancy Box for Content
	contentBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#04B575")).
		Padding(1, 2).
		Width(60).
		Render(content)

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, logView, contentBox, "\n\n"+logStyle.Render("q: quit")))
}

// -- Commands --

func analyzeCmd(client *openai.Client, input string, history []string) tea.Cmd {
	return func() tea.Msg {
		res, err := agents.AnalyzeRequest(client, input, history)
		if err != nil {
			return analysisMsg{Status: "error", Reason: err.Error()}
		}
		return analysisMsg(res)
	}
}

func researchCmd(client *openai.Client, topic string, history []string) tea.Cmd {
	return func() tea.Msg {
		res, err := agents.Research(client, topic, history)
		if err != nil {
			return researchMsg("Research failed but continuing...")
		}
		return researchMsg(res)
	}
}

func architectCmd(client *openai.Client, reqs, research string, currentFlow *agents.Flowchart) tea.Cmd {
	return func() tea.Msg {
		res, err := agents.GenerateFlowchart(client, reqs, research, currentFlow)
		if err != nil {
			return errMsg{err}
		}
		return architectMsg(res)
	}
}

func judgeCmd(client *openai.Client, fc agents.Flowchart, reqs string) tea.Cmd {
	return func() tea.Msg {
		// Serialize flow to json for the judge
		jsonBytes, err := json.Marshal(fc)
		if err != nil {
			// Fallback string representation
			return judgeMsg{Approved: true}
		}

		res, err := agents.Judgement(client, string(jsonBytes), reqs)
		if err != nil {
			return judgeMsg{Approved: true} // Fail open on API error so user gets result
		}
		return judgeMsg(res)
	}
}

func generateCmd(fc agents.Flowchart) tea.Cmd {
	return func() tea.Msg {
		// Create a sanitized filename from the title
		title := fc.Overview.Title
		if title == "" {
			title = "untitled_flow"
		}

		// Simple sanitization
		safeTitle := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return '_'
		}, title)

		// Add timestamp to ensure uniqueness/history
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_%s_flow.html", safeTitle, timestamp)

		err := generator.GenerateHTML(fc, filename)
		// Return the filename in the msg so we can show it
		if err == nil {
			return generationMsg{err: nil, filename: filename}
		}
		return generationMsg{err: err}
	}
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// Helper to find json files
func getJSONFiles() ([]string, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var jsons []string
	for _, f := range files {
		// Look for files ending in _flow.json to keep it clean
		if !f.IsDir() && strings.HasSuffix(f.Name(), "_flow.json") {
			jsons = append(jsons, f.Name())
		} else if strings.EqualFold(f.Name(), "flowchart.json") {
			// Also include the legacy/default one
			jsons = append(jsons, f.Name())
		}
	}
	return jsons, nil
}

func openFileInOS(filename string) {
	cwd, _ := os.Getwd()
	absPath := filepath.Join(cwd, filename)

	// Mac specific 'open'
	cmd := exec.Command("open", absPath)
	_ = cmd.Run()
}
