package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Task represents a single step in a multi-step operation.
type Task struct {
	Label string
	Run   func() error
}

// stepDoneMsg signals that a task step completed.
type stepDoneMsg struct {
	err error
}

type model struct {
	tasks   []Task
	current int
	done    []bool
	errs    []error
	spinner spinner.Model
	err     error
}

func newModel(tasks []Task) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = Cyan
	return model{
		tasks:   tasks,
		done:    make([]bool, len(tasks)),
		errs:    make([]error, len(tasks)),
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runCurrent())
}

func (m model) runCurrent() tea.Cmd {
	i := m.current
	if i >= len(m.tasks) {
		return tea.Quit
	}
	task := m.tasks[i]
	return func() tea.Msg {
		err := task.Run()
		return stepDoneMsg{err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.err = fmt.Errorf("interrupted")
			return m, tea.Quit
		}
	case stepDoneMsg:
		m.done[m.current] = true
		m.errs[m.current] = msg.err
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.current++
		if m.current >= len(m.tasks) {
			return m, tea.Quit
		}
		return m, m.runCurrent()
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	for i, task := range m.tasks {
		if i < m.current {
			// completed
			if m.errs[i] != nil {
				b.WriteString(fmt.Sprintf("%s %s\n", Red.Render("✗"), task.Label))
			} else {
				b.WriteString(fmt.Sprintf("%s %s\n", Green.Render("✓"), task.Label))
			}
		} else if i == m.current {
			if m.done[i] {
				if m.errs[i] != nil {
					b.WriteString(fmt.Sprintf("%s %s\n", Red.Render("✗"), task.Label))
				} else {
					b.WriteString(fmt.Sprintf("%s %s\n", Green.Render("✓"), task.Label))
				}
			} else {
				b.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), task.Label))
			}
		}
		// future tasks not shown until reached
	}
	return b.String()
}

// RunTasks runs a sequence of tasks with a spinner UI.
// Each task is displayed with a spinner while running and a checkmark when done.
// Returns the first error encountered, or nil if all tasks succeed.
func RunTasks(tasks []Task) error {
	if len(tasks) == 0 {
		return nil
	}

	m := newModel(tasks)
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("ui error: %w", err)
	}

	final := result.(model)
	return final.err
}
