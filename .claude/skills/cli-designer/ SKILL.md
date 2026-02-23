---
name: cli-designer
description: Create distinctive, production-grade terminal user interfaces (TUIs) in Go using the Charm ecosystem (bubbletea, bubbles, lipgloss). Use this skill whenever the user asks to build CLI applications, terminal UIs, interactive command-line tools, TUI dashboards, terminal forms, or any Go-based terminal interface. Also triggers for styling or beautifying existing bubbletea apps, designing terminal layouts, or building interactive prompts. If the user mentions bubbletea, bubbles, lipgloss, Charm, TUI, or wants to make a CLI tool look polished and distinctive, use this skill.
---

This skill guides creation of distinctive, production-grade terminal user interfaces using Go and the Charm ecosystem. The goal: TUIs that feel *crafted*, not slapped together. Every app should feel like someone who cares built it.

The user provides TUI requirements: a component, application, dashboard, form, or interactive tool to build. They may include context about the purpose, audience, or technical constraints.

## The Charm Stack

Before designing anything, know the tools:

| Package | Role | Import |
|---------|------|--------|
| **bubbletea** | Application framework (Elm Architecture: Model → Update → View) | `github.com/charmbracelet/bubbletea` |
| **bubbles** | Pre-built components (list, table, textinput, textarea, viewport, spinner, progress, paginator, filepicker, help, timer, stopwatch) | `github.com/charmbracelet/bubbles/*` |
| **lipgloss** | Styling engine (colors, borders, padding, margins, alignment, layout composition) | `github.com/charmbracelet/lipgloss` |
| **huh** | Declarative forms and prompts | `github.com/charmbracelet/huh` |
| **glamour** | Markdown rendering in the terminal | `github.com/charmbracelet/glamour` |
| **harmonica** | Spring-based animations | `github.com/charmbracelet/harmonica` |
| **log** | Styled logging | `github.com/charmbracelet/log` |
| **lipgloss/table** | Styled table rendering | `github.com/charmbracelet/lipgloss/table` |

Always use the latest stable versions. When in doubt, check `go doc` or the Charm GitHub repos.

## Design Thinking for the Terminal

Before writing code, understand the medium. A terminal is NOT a web browser. It's a constrained, text-driven canvas — and those constraints are what make great TUIs feel so satisfying. Embrace them.

### Commit to an Aesthetic Direction

Pick a vibe and commit hard:

- **Neo-retro terminal** — Phosphor greens, amber glows, scanline vibes. CRT energy.
- **Brutalist CLI** — Raw, dense, no decoration. Information as interface. Monospace poetry.
- **Soft/pastel** — Muted tones, rounded borders, generous whitespace. Cozy terminal.
- **Cyberpunk dashboard** — Neon accents on dark backgrounds, box-drawing art, data-dense layouts.
- **Minimal zen** — One or two colors, extreme restraint, every character earns its place.
- **Luxury dev tool** — The Stripe/Linear of CLIs. Refined typography, subtle color, impeccable spacing.
- **Playful/whimsical** — Emoji accents, fun spinners, personality in every message.
- **Editorial/structured** — Magazine-like sections, clear hierarchy, strong headers.

These are starting points, not a menu. Invent something new. The key is *intentionality* — every color, border, and spacing choice should serve the aesthetic.

### Terminal Design Principles

These are the things that separate a great TUI from a mediocre one:

1. **Respect the grid.** Every character is a cell. Alignment matters more here than anywhere. Off-by-one spacing errors are visually catastrophic in a terminal.

2. **Color is your most powerful tool — use it surgically.** You have far fewer pixels than the web. A single accent color against a muted palette does more work than a rainbow. Define a tight palette: 1 background, 1 foreground, 1-2 accents, 1 muted/dim. That's it for most apps.

3. **Borders and box-drawing characters are architecture.** They define regions, create hierarchy, and give structure. lipgloss provides `NormalBorder`, `RoundedBorder`, `ThickBorder`, `DoubleBorder`, `HiddenBorder`, and `BlockBorder`. Pick one and use it consistently — or define a custom `lipgloss.Border` for something unique.

4. **Whitespace is not wasted space.** Padding inside boxes, margins between sections, blank lines between logical groups — these create breathing room. Dense TUIs feel claustrophobic. Use `Padding()` and `Margin()` generously.

5. **Hierarchy through styling, not just layout.** Bold for primary, dim for secondary, color for interactive/actionable, strikethrough for disabled. lipgloss gives you `Bold()`, `Faint()`, `Italic()`, `Underline()`, `Strikethrough()`, and `Reverse()`.

6. **Responsive to terminal size.** Always handle `tea.WindowSizeMsg`. Your UI must adapt. Hard-coded widths are a bug, not a feature.

7. **Motion and feedback.** Spinners for async work. Progress bars for known-duration tasks. Immediate visual feedback on every keypress. The user should never wonder "did that work?"

## lipgloss Styling Patterns

### Define a Theme, Not Inline Styles

Always centralize your styling. Never scatter `lipgloss.NewStyle()` calls through your View function.

```go
// theme.go — define once, use everywhere
package main

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
    colorPrimary   = lipgloss.Color("#E0ADF7")
    colorSecondary = lipgloss.Color("#7B6B8D")
    colorAccent    = lipgloss.Color("#FF6F91")
    colorDim       = lipgloss.Color("#555555")
    colorBg        = lipgloss.Color("#1A1A2E")
    colorSuccess   = lipgloss.Color("#98C379")
    colorWarning   = lipgloss.Color("#E5C07B")
    colorError     = lipgloss.Color("#E06C75")
)

// Reusable styles
var (
    styleTitle = lipgloss.NewStyle().
        Bold(true).
        Foreground(colorPrimary).
        MarginBottom(1)

    styleSubtle = lipgloss.NewStyle().
        Foreground(colorDim)

    styleActive = lipgloss.NewStyle().
        Foreground(colorAccent).
        Bold(true)

    styleBox = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(colorSecondary).
        Padding(1, 2)

    styleStatusBar = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFFDF5")).
        Background(lipgloss.Color("#6124DF")).
        Padding(0, 1)
)
```

This is non-negotiable. A theme file is the single biggest quality-of-life improvement for any TUI codebase.

### Adaptive Colors for Light/Dark Terminals

Always consider that users might have light OR dark terminal backgrounds:

```go
var subtleColor = lipgloss.AdaptiveColor{
    Light: "#555555",
    Dark:  "#AAAAAA",
}
```

### Layout Composition

lipgloss's `JoinHorizontal`, `JoinVertical`, and `Place` functions are your layout engine:

```go
func (m model) View() string {
    // Compose panels side by side
    left := styleBox.Width(m.width/2 - 2).Render(m.leftContent())
    right := styleBox.Width(m.width/2 - 2).Render(m.rightContent())
    body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

    // Stack vertically: header → body → status
    return lipgloss.JoinVertical(lipgloss.Left,
        m.headerView(),
        body,
        m.statusBarView(),
    )
}
```

### lipgloss Table for Data Display

The `lipgloss/table` package is powerful — use it for any structured data:

```go
import "github.com/charmbracelet/lipgloss/table"

t := table.New().
    Border(lipgloss.RoundedBorder()).
    BorderStyle(lipgloss.NewStyle().Foreground(colorSecondary)).
    Headers("NAME", "STATUS", "UPTIME").
    Row("api-server", "● Running", "4d 12h").
    Row("worker-01", "● Running", "4d 12h").
    Row("scheduler", "○ Stopped", "—").
    StyleFunc(func(row, col int) lipgloss.Style {
        if row == table.HeaderRow {
            return lipgloss.NewStyle().Bold(true).Foreground(colorPrimary)
        }
        if row%2 == 0 {
            return lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
        }
        return lipgloss.NewStyle()
    })
```

## bubbletea Architecture Patterns

### The Model-Update-View Contract

Every bubbletea app implements `tea.Model`:

```go
type model struct {
    // State
    items    []item
    cursor   int
    selected map[int]struct{}

    // UI state
    width  int
    height int

    // Sub-models (bubbles components)
    list    list.Model
    spinner spinner.Model
    help    help.Model

    // Application state
    loading bool
    err     error
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        m.spinner.Tick,       // Start spinner animation
        tea.SetWindowTitle("My App"),
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}

func (m model) View() string {
    if m.loading {
        return m.loadingView()
    }
    return m.mainView()
}
```

### Multi-View State Machine

For apps with multiple screens, use an explicit state enum:

```go
type viewState int

const (
    viewList viewState = iota
    viewDetail
    viewEdit
    viewConfirm
)

type model struct {
    state viewState
    // ... sub-models for each view
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Route updates to the active view's sub-model
    switch m.state {
    case viewList:
        return m.updateList(msg)
    case viewDetail:
        return m.updateDetail(msg)
    // ...
    }
    return m, nil
}

func (m model) View() string {
    switch m.state {
    case viewList:
        return m.listView()
    case viewDetail:
        return m.detailView()
    // ...
    }
    return ""
}
```

### Async Operations

Commands are how you do I/O. Never block in Update:

```go
// Define a message type for the result
type fetchResultMsg struct {
    data []item
    err  error
}

// Return a Cmd that performs async work
func fetchItems(url string) tea.Cmd {
    return func() tea.Msg {
        resp, err := http.Get(url)
        if err != nil {
            return fetchResultMsg{err: err}
        }
        defer resp.Body.Close()
        var items []item
        json.NewDecoder(resp.Body).Decode(&items)
        return fetchResultMsg{data: items}
    }
}

// Handle the result in Update
case fetchResultMsg:
    m.loading = false
    if msg.err != nil {
        m.err = msg.err
        return m, nil
    }
    m.items = msg.data
```

### Delegating to Sub-Models

When using bubbles components, always propagate messages:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    // Handle global keys first
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
    }

    // Then delegate to sub-models
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)

    m.spinner, cmd = m.spinner.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}
```

## Common UI Recipes

### Status Bar

```go
func (m model) statusBarView() string {
    left := styleStatusBar.Render(" ● Connected ")
    right := styleStatusBar.Render(" Press ? for help ")
    gap := strings.Repeat(" ", max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right)))
    mid := lipgloss.NewStyle().
        Background(lipgloss.Color("#3C3C5C")).
        Render(gap)
    return left + mid + right
}
```

### Keybinding Help

Always include a help view — it's table stakes for any serious TUI:

```go
import "github.com/charmbracelet/bubbles/help"
import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
    Up    key.Binding
    Down  key.Binding
    Enter key.Binding
    Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Up, k.Down, k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down},
        {k.Enter, k.Quit},
    }
}

var keys = keyMap{
    Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
    Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
    Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
    Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
```

### Error States

Don't ignore errors — style them:

```go
func (m model) errorView() string {
    icon := lipgloss.NewStyle().Foreground(colorError).Render("✗")
    title := lipgloss.NewStyle().Bold(true).Foreground(colorError).Render("Error")
    msg := lipgloss.NewStyle().Foreground(colorDim).Render(m.err.Error())

    content := fmt.Sprintf("%s %s\n\n%s\n\n%s",
        icon, title, msg,
        styleSubtle.Render("Press any key to retry, q to quit"))

    return lipgloss.Place(m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        styleBox.Render(content))
}
```

### Loading States

```go
func (m model) loadingView() string {
    content := fmt.Sprintf("%s Loading...", m.spinner.View())
    return lipgloss.Place(m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        styleSubtle.Render(content))
}
```

## CLI Aesthetics Anti-Patterns

NEVER do these:

- **No visual hierarchy.** Walls of unstyled text with no color, no bold, no spacing. If your View() returns plain `fmt.Sprintf` output, you've failed.
- **Hard-coded dimensions.** Using fixed widths instead of responding to `tea.WindowSizeMsg`. The terminal can be any size — your app must adapt.
- **Ignoring the help system.** Every interactive app needs discoverable keybindings. Use the `help` bubble.
- **Rainbow vomit.** Using 8+ colors with no coherent palette. A TUI with too many colors looks worse than one with none. Stick to your theme.
- **Walls of borders.** Boxing every single element. Borders create regions — use them to group, not to decorate every line.
- **No loading/error states.** If your app does I/O, it needs a spinner for loading and a styled error view for failures. Blanking out or panicking is not UX.
- **Blocking in Update.** Never do synchronous I/O in Update. Use `tea.Cmd` for anything that touches the network, filesystem, or takes time.
- **Ignoring vim-style keys.** Power users expect `j/k` for up/down, `g/G` for top/bottom, `/` for search. If you're building a list or navigation, support these alongside arrow keys.

## Implementation Checklist

Before shipping any TUI, verify:

- [ ] Responds to `tea.WindowSizeMsg` and adapts layout
- [ ] Has a centralized theme (colors, styles) — not inline styles
- [ ] Includes help keybindings (at minimum `?` to toggle help)
- [ ] Handles `ctrl+c` and `q` for quitting gracefully
- [ ] Loading states with spinners for any async work
- [ ] Error states that are styled and actionable (retry/quit)
- [ ] Consistent border style throughout
- [ ] Color palette is tight (max 5-6 intentional colors)
- [ ] Padding and margins create visual breathing room
- [ ] Alt-screen mode enabled for full-screen apps (`tea.WithAltScreen()`)
- [ ] Mouse support if appropriate (`tea.WithMouseCellMotion()`)
- [ ] Clean exit — no leftover terminal state artifacts

## Project Structure

For non-trivial TUI apps, organize like this:

```
app/
├── app.go           # Program entry, tea.NewProgram setup
├── model.go         # Main model struct, Init/Update/View
├── update.go        # Update logic (if model.go gets large)
├── views.go         # View rendering functions
├── theme.go         # All lipgloss styles and color definitions
├── keys.go          # Key bindings and help keymap
├── commands.go      # tea.Cmd functions for I/O
└── components/      # Custom reusable sub-models
    ├── statusbar.go
    └── panel.go
```

Keep `View()` functions clean — they should read like a layout description, not contain styling logic. If a view function is longer than ~40 lines, break it into sub-view methods.

## Remember

The terminal is an intimate medium. People stare at their terminals all day. A well-designed TUI earns trust and affection in a way that a web UI rarely does. Don't waste that opportunity with lazy defaults. Every character, every color, every border is a choice — make it deliberately.