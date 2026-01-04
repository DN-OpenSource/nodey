# Nodey: The AI-Powered Flowchart Architect

**Nodey** is a terminal-based intelligent coding assistant that designs, validates, and generates beautiful, interactive flowcharts from natural language descriptions. It leverages a Multi-Agent AI System to ensure your workflows are logical, comprehensive, and visually stunning.

## 🚀 Features

*   **Multi-Agent Intelligence**:
    *   🕵️ **Analyst**: Clarifies your requirements and asks the right questions.
    *   📚 **Researcher**: Gathers context and best practices for your specific topic.
    *   🏗️ **Architect**: Designs the flowchart structure (Nodes, Decisions, Connections).
    *   ⚖️ **Judge**: Critiques the Flow to ensure logic and quality before generation.
*   **Premium UI Output**: Generates a high-quality, interactive HTML file with:
    *   **Auto-Layout**: Perfect spacing using Dagre.js (no manual dragging needed).
    *   **Infinite Canvas**: Pan and Zoom controls.
    *   **Smart Inspection**: Click any node to see detailed technical notes in a glassmorphic sidebar.
*   **Iterative Design**:
    *   Load previous flows from history.
    *   Ask Nodey to "Add a step here" or "Change logic there" and it understands the existing context.
*   **Terminal Native**:
    *   Built with Bubble Tea/Lip Gloss for a stunning TUI (Terminal User Interface).
    *   History management (`Ctrl+L`).
    *   Direct "Open in Browser" shortcuts.

---

## 🛠️ Installation

### Prerequisites
*   **Homebrew** (Package manager for macOS)
*   **OpenAI API Key**

### Install via Homebrew
```bash
brew tap DN-OpenSource/tap
brew install nodey
```

---

## ⚡ Usage

### Quick Start
1.  **Set your API Key**:
    ```bash
    export OPENAI_API_KEY="sk-..."
    ```
2.  **Run Nodey**:
    ```bash
    ./nodey
    ```
3.  **Describe your Flow**:
    > "Create a user registration flow with email verification and 2FA."

### Interactive Commands
| Key | Action | Context |
| :--- | :--- | :--- |
| `Enter` | Submit prompt / Select file | Input / History |
| `Ctrl+L` | **Open History** | Input |
| `o` | **Open in Browser** | History / Done Screen |
| `Esc` | Cancel / Back | History |
| `q` / `Ctrl+C` | Quit | Anywhere |

### Edit Existing Flows
1. Press `Ctrl+L` to view your saved history.
2. Select a flow and press `Enter`.
3. Nodey will load the context. Simply type your modification:
    > "Add a 'Forgot Password' branch after the login failure decision."

---

## 📂 Output

Nodey generates two files for every session (timestamped to prevent overwrites):
1.  `Title_YYYYMMDD_HHMMSS_flow.html`: The visual, interactive flowchart.
2.  `Title_YYYYMMDD_HHMMSS_flow.json`: The raw data used for history and iterative editing.

---

## 🏗️ Architecture

Nodey is built in **Go** and uses the **ELM Architecture (Model-View-Update)** via the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

### Directory Structure
*   `main.go`: Entry point. Handles the TUI state machine and user input.
*   `agents/`: Contains the logic for the specific AI agents (Architect, Judge, etc.).
*   `generator/`: Handles the HTML/JS generation logic (Dagre.js integration).
*   `release_to_homebrew.md`: Internal guide for distribution.

### Tech Stack
*   **Language**: Go (Golang)
*   **TUI**: Bubble Tea, Lip Gloss, Bubbles
*   **AI**: OpenAI API (GPT-4o/GPT-3.5-turbo)
*   **Frontend**: HTML5, CSS3, Dagre.js (Graph Layout)

---

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a PR for any bug fixes or feature additions.

## 📄 License

MIT License. See [LICENSE](LICENSE) file for details.
# nodey
