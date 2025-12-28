# Nodey Implementation Details & Architecture Guide

This document provides a deep dive into the internal workings of **Nodey**, the Multi-Agent Flowchart Builder. It is intended for developers who want to understand the code structure, the agentic workflow, or contribute to the project.

## 1. System Architecture

Nodey follows the **ELM Architecture** pattern provided by the **Bubble Tea** framework. This ensures a clean separation of State (Model), Logic (Update), and UI (View).

### State Machine (`main.go`)

The application moves strictly between the following states:

1.  **`stateInput`**: The resting state. Waiting for user text input.
    *   *Transitions to*: `stateAnalyzing` (on Enter), `stateHistory` (on Ctrl+L).
2.  **`stateAnalyzing`**: The **Analyst Agent** reviews the prompt.
    *   *Transitions to*: `stateResearching` (if clear), `stateAnswering` (if ambiguous).
3.  **`stateAnswering`**: The user answers clarifying questions from the Analyst.
    *   *Transitions to*: `stateResearching` (once all questions are answered).
4.  **`stateResearching`**: The **Researcher Agent** fetches domain knowledge.
    *   *Transitions to*: `stateArchitecting`.
5.  **`stateArchitecting`**: The **Architect Agent** designs the JSON graph structure.
    *   *Transitions to*: `stateJudging`.
6.  **`stateJudging`**: The **Judge Agent** critiques the graph.
    *   *Transitions to*: `stateGenerating` (if approved), `stateArchitecting` (if rejected, with feedback).
7.  **`stateGenerating`**: The system writes the HTML/JSON artifacts.
    *   *Transitions to*: `stateDone`.
8.  **`stateDone`**: Final success screen. Allows opening the file or quitting.
    *   *Transitions to*: `stateInput` (on Esc/Reset).

## 2. The AI Agents (`agents/` package)

Each agent is a specialized wrapper around an LLM call with a specific System Prompt.

### A. The Architect (`architect.go`)
*   **Role**: Converts natural language + research context into a strict JSON format.
*   **Input**: User Prompt, Research Summary, (Optional) Previous Flowchart JSON.
*   **Output**: `Flowchart` struct (Nodes list, Connections list).
*   **Logic**: It decides the Node Types (`start`, `action`, `decision`, `end`) and the logic flow.

### B. The Judge (`judge.go`)
*   **Role**: Quality Assurance.
*   **Input**: The drafted JSON Flowchart.
*   **Output**: Boolean `Approved`, String `Critique`.
*   **Logic**: Checks for infinite loops, orphaned nodes, illogical paths, or missing requirements.

### C. The Analyst (`analyst.go`)
*   **Role**: Front-line filter.
*   **Input**: Raw user text.
*   **Output**: Status `valid` vs `needs_info`, plus clarifying questions.

## 3. The Generator (`generator/html.go`)

This package is responsible for turning the abstract JSON data into a visual `HTML` file.

*   **Technology**:
    *   **Dagre.js**: Used for deterministic, hierarchical graph layout. We send the *nodes* and *connections* to the browser, and the embedded Javascript calculates the exact X/Y coordinates on load.
    *   **SVG**: Draws the connection lines with smooth Bezier curves or orthogonal paths.
    *   **CSS Objects**: Nodes are rendered as distinct HTML `div` elements with CSS styling for shadows, borders, and interaction.

### "Premium" UI Features Implemented
*   **Glassmorphism Sidebar**: The inspector panel uses `backdrop-filter: blur()`.
*   **Infinite Grid**: CSS `background-image` with linear gradients creates a scalable grid pattern.
*   **Smart Anchors**: The Javascript heuristically decides whether a line should exit from the Bottom or Right of a node to minimize crossing.

## 4. History & Iteration

Nodey is not just one-shot. It saves the *state* of the flow.

*   **Autosave**: Every generation saves a `_flow.json` file.
*   **Loading**: The `stateHistory` view scans the directory for these JSON files.
*   **Modification**: When a user loads a flow and prompts a change, the **Architect** receives the *entire existing JSON* as context. It is instructed to "Edit the existing flow" rather than creating from scratch, preserving IDs and layout where possible.

## 5. Deployment

The app is compiled into a single binary.
*   **Ref**: `release_to_homebrew.md` for distribution instructions.
*   **Requirements**: The end user only needs the binary; all HTML/CSS/JS assets are embedded in the final HTML file string.

---

**Codebase maintained by Antigravity (Google Deepmind)**
