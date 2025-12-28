# How to Use FlowBuilder 🌊

## 🆕 New Features
- **Load & Edit**: You can now load previous flowcharts and iteratively update them!
- **Premium UI**: The generated HTML now features an infinite canvas, dragging, and a beautiful dark/light hybrid theme.
- **Node Inspector**: Click any node to see the technical details in a sidebar.

## 🚀 workflow

### 1. Create a New Flow
Run the app:
```bash
./flowbuilder
```
Enter your prompt:
> "User Registration flow with email verification"

### 2. Edit an Existing Flow
1. Run the app.
2. Press **`Ctrl+L`** to open the Load Menu.
3. Select a previously saved `.json` file using Up/Down arrows and Enter.
4. You will see `[Editing flowchart.json]`.
5. Enter your modification request:
> "Add a check for 'User Banned' before sending email"

The Architect will take your existing flow and surgically insert the new logic while preserving the rest!

### 3. View the Result
Open `flowchart.html` in your browser.
- **Pan**: Click and drag empty space.
- **Zoom**: Mouse wheel.
- **Move Nodes**: Drag nodes to rearrange them (Shift-drag if needed).
- **Inspect**: Click a node to view its AI-generated technical notes.

## 🛠️ Installation (Homebrew)
See `release_to_homebrew.md` for instructions on how to package this for distribution.
