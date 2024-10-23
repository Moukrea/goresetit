package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87C1FF"))

	warning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA07A"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF616E")).
		Bold(true)

	logoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF86C8")).
		Bold(true)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF86C8")).
		Bold(true).
		MarginLeft(2)

	inputStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#87C1FF")).
		MarginLeft(2)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA07A")).
		Bold(true).
		MarginLeft(2)
)

// ASCII art generated using: https://patorjk.com/software/taag/#p=display&f=Thick&t=GoresetIT
const Logo = `
.d88b                               w   888 88888 
8P www .d8b. 8d8b .d88b d88b .d88b w8ww  8    8   
8b  d8 8' .8 8P   8.dP' ` + "`" + `Yb. 8.dP'  8    8    8   
` + "`" + `Y88P' ` + "`" + `Y8P' 8    ` + "`" + `Y88P Y88P ` + "`" + `Y88P  Y8P 888   8   

`

// For mocking in tests
var newTeaProgram = tea.NewProgram

func ShowLogo() {
	println(logoStyle.Render(Logo))
	println()
}

// CommitModel handles the commit message input
type CommitModel struct {
	TextInput textinput.Model
	Err       error
	Done      bool
}

func InitialCommitModel() CommitModel {
	ti := textinput.New()
	ti.Placeholder = "Initial commit"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return CommitModel{
		TextInput: ti,
		Err:       nil,
		Done:      false,
	}
}

func (m CommitModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CommitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.TextInput.Value() != "" {
				m.Done = true
				return m, tea.Quit
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case error:
		m.Err = msg
		return m, nil
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m CommitModel) View() string {
	if m.Err != nil {
		return titleStyle.Render("Error: ") + m.Err.Error() + "\n"
	}

	var s string
	s += titleStyle.Render("Enter the initial commit message:\n\n")
	s += inputStyle.Render(m.TextInput.View()) + "\n\n"
	s += inputStyle.Render("(Press Enter to confirm or Esc/Ctrl+C to cancel)") + "\n"

	return s
}

// ConfirmModel handles the confirmation prompt
type ConfirmModel struct {
	Question string
	Done     bool
	Answer   bool
}

func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.Done = true
			m.Answer = true
			return m, tea.Quit
		case "n", "N", "q", "Q", "esc", "ctrl+c":
			m.Done = true
			m.Answer = false
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ConfirmModel) View() string {
	var s string
	s += warningStyle.Render("⚠️  WARNING ⚠️\n\n")
	s += titleStyle.Render(m.Question) + "\n\n"
	s += inputStyle.Render("Press 'y' to continue or 'n' to cancel") + "\n"
	return s
}

func PromptCommitMessage() (string, error) {
	p := newTeaProgram(InitialCommitModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	model, ok := m.(CommitModel)
	if !ok {
		return "", nil
	}

	if !model.Done {
		return "", nil
	}

	return model.TextInput.Value(), nil
}

func PromptConfirmation(dryRun bool) (bool, error) {
	var question string
	if dryRun {
		question = "GoresetIT will simulate squashing all commits on main branch (DRY RUN).\n" +
			"This operation will perform all local operations but won't push any changes.\n" +
			"Are you sure you want to continue?"
	} else {
		question = "GoresetIT will squash all commits on main branch.\n" +
			"THIS IS A DESTRUCTIVE OPERATION AND CANNOT BE UNDONE!\n" +
			"Are you sure you want to continue?"
	}

	model := ConfirmModel{
		Question: question,
	}

	p := newTeaProgram(model)
	m, err := p.Run()
	if err != nil {
		return false, err
	}

	finalModel, ok := m.(ConfirmModel)
	if !ok {
		return false, nil
	}

	return finalModel.Answer, nil
}