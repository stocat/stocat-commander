package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"stocat-commander/runner"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateProjectList sessionState = iota
	stateCommandList
	stateCommandLog
	stateSetupLog
)

type Model struct {
	state         sessionState
	projects      []Project
	projectCursor int
	commandCursor int
	workspaceDir  string

	// For logging view
	viewport     viewport.Model
	logs         []string
	filteredLogs []string
	searchInput  textinput.Model
	isSearching  bool
	searchQuery  string

	// Progress indicator
	spinner spinner.Model

	// App reference
	Program *tea.Program

	// Status flags
	isRunning bool
	err       error

	// Window dimensions
	width  int
	height int
}

func InitialModel(workspaceDir string) Model {
	projs := GetProjects(workspaceDir)

	vp := viewport.New(0, 0)
	vp.SetContent("No logs yet...")

	ti := textinput.New()
	ti.Placeholder = "Search logs... (Enter to confirm, Esc to cancel)"
	ti.CharLimit = 156
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{
			"⠋ █░░░░░░░░░",
			"⠙ ██░░░░░░░░",
			"⠹ ███░░░░░░░",
			"⠸ ████░░░░░░",
			"⠼ █████░░░░░",
			"⠴ ██████░░░░",
			"⠦ ███████░░░",
			"⠧ ████████░░",
			"⠇ █████████░",
			"⠏ ██████████",
			"⠋ ░█████████",
			"⠙ ░░████████",
			"⠹ ░░░███████",
			"⠸ ░░░░██████",
			"⠼ ░░░░░█████",
			"⠴ ░░░░░░████",
			"⠦ ░░░░░░░███",
			"⠧ ░░░░░░░░██",
			"⠇ ░░░░░░░░░█",
			"⠏ ░░░░░░░░░░",
		},
		FPS: time.Second / 15,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Bold(true)

	return Model{
		state:         stateProjectList,
		projects:      projs,
		projectCursor: 0,
		commandCursor: 0,
		workspaceDir:  workspaceDir,
		viewport:      vp,
		spinner:       s,
		logs:          []string{},
		filteredLogs:  []string{},
		searchInput:   ti,
		isSearching:   false,
		searchQuery:   "",
		isRunning:     false,
		width:         80, // Safe defaults
		height:        24,
	}
}

// highlightSearchTerm surrounds the case-insensitive matches of query inside text with a lipgloss style
func highlightSearchTerm(text, query string) string {
	if query == "" {
		return text
	}

	// Fast path case-insensitive search by using strings.Index on lowercased versions
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	qLen := len(query)

	var sb strings.Builder
	lastIdx := 0

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FDE047")).Bold(true) // Yellow background

	for {
		idx := strings.Index(lowerText[lastIdx:], lowerQuery)
		if idx == -1 {
			sb.WriteString(text[lastIdx:])
			break
		}

		realIdx := lastIdx + idx
		sb.WriteString(text[lastIdx:realIdx])
		sb.WriteString(style.Render(text[realIdx : realIdx+qLen]))

		lastIdx = realIdx + qLen
	}

	return sb.String()
}

// updateLogView helper logic
func (m Model) updateLogView() Model {
	if m.searchQuery == "" {
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
	} else {
		m.filteredLogs = []string{}
		for _, l := range m.logs {
			if strings.Contains(strings.ToLower(l), strings.ToLower(m.searchQuery)) {
				m.filteredLogs = append(m.filteredLogs, highlightSearchTerm(l, m.searchQuery))
			}
		}
		if len(m.filteredLogs) == 0 {
			m.viewport.SetContent("No matching logs found.")
		} else {
			m.viewport.SetContent(strings.Join(m.filteredLogs, "\n"))
		}
	}
	return m
}

// SetProgram is used to inject the tea.Program pointer after initialization
func (m *Model) SetProgram(p *tea.Program) {
	m.Program = p
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.isSearching {
			switch msg.String() {
			case "enter":
				m.isSearching = false
				m.searchQuery = m.searchInput.Value()
				m = m.updateLogView()
				return m, nil
			case "esc":
				m.isSearching = false
				m.searchInput.SetValue(m.searchQuery)
				return m, nil
			default:
				m.searchInput, cmd = m.searchInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == stateProjectList || !m.isRunning {
				return m, tea.Quit
			}
		case "up", "k", "pgup", "u":
			if m.state == stateCommandLog || m.state == stateSetupLog {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			} else if m.state == stateProjectList && m.projectCursor > 0 {
				m.projectCursor--
			} else if m.state == stateCommandList && m.commandCursor > 0 {
				m.commandCursor--

				// Skip separator when moving up
				proj := m.projects[m.projectCursor]
				if proj.Commands[m.commandCursor].IsSeparator && m.commandCursor > 0 {
					m.commandCursor--
				}
			}
		case "down", "j", "pgdown", "d":
			if m.state == stateCommandLog || m.state == stateSetupLog {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			} else if m.state == stateProjectList && m.projectCursor < len(m.projects)-1 {
				m.projectCursor++
			} else if m.state == stateCommandList {
				proj := m.projects[m.projectCursor]
				if m.commandCursor < len(proj.Commands)-1 {
					m.commandCursor++

					// Skip separator when moving down
					if proj.Commands[m.commandCursor].IsSeparator && m.commandCursor < len(proj.Commands)-1 {
						m.commandCursor++
					}
				}
			}
		case "/":
			if m.state == stateCommandLog || m.state == stateSetupLog {
				m.isSearching = true
				m.searchInput.Focus()
				return m, nil
			}
		case "enter":
			if m.state == stateProjectList {
				proj := m.projects[m.projectCursor]

				// Check if project path exists
				if _, err := os.Stat(proj.Path); os.IsNotExist(err) {
					m.state = stateSetupLog
					m.logs = []string{fmt.Sprintf("Project %s not found.", proj.Name), fmt.Sprintf("Cloning from: %s", proj.RepoURL)}
					m.viewport.SetContent(strings.Join(m.logs, "\n"))
					m.viewport.GotoBottom()

					m.isRunning = true
					m.err = nil
					// Run git clone
					cmds = append(cmds, runner.StartCommand(m.workspaceDir, "git", []string{"clone", proj.RepoURL}, m.Program))
				} else {
					m.state = stateCommandList
					m.commandCursor = 0 // Reset command selection
				}
			} else if m.state == stateCommandList {
				proj := m.projects[m.projectCursor]
				cmdToRun := proj.Commands[m.commandCursor]

				// Determine running directory
				runDir := proj.Path
				if cmdToRun.DirName != "" {
					runDir = filepath.Join(m.workspaceDir, cmdToRun.DirName)
				}

				if cmdToRun.Interactive {
					// Open interactive terminal app (e.g. less)
					c := exec.Command(cmdToRun.Exec, cmdToRun.Args...)
					c.Dir = runDir

					m.state = stateCommandLog
					m.isRunning = true
					m.err = nil
					m.logs = []string{fmt.Sprintf("Opened interactive viewer: %s", cmdToRun.Name)}
					m.viewport.SetContent(strings.Join(m.logs, "\n"))

					cmd = tea.ExecProcess(c, func(err error) tea.Msg {
						return runner.ExecFinishedMsg{Err: err}
					})
					cmds = append(cmds, cmd)
				} else {
					// Async background execution
					m.state = stateCommandLog
					cmdStr := cmdToRun.Exec + " " + strings.Join(cmdToRun.Args, " ")
					m.logs = []string{fmt.Sprintf("Running: %s in %s", cmdStr, runDir)}
					m.viewport.SetContent(strings.Join(m.logs, "\n"))
					m.viewport.GotoBottom()

					m.isRunning = true
					m.err = nil
					cmds = append(cmds, runner.StartCommand(runDir, cmdToRun.Exec, cmdToRun.Args, m.Program))
				}
			}
		case "esc", "backspace":
			if m.state == stateCommandList {
				m.state = stateProjectList
			} else if m.state == stateCommandLog {
				m.state = stateCommandList
				m.searchQuery = ""
				m.searchInput.SetValue("")
			} else if m.state == stateSetupLog {
				m.state = stateProjectList
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = max(1, msg.Width-10)
		m.viewport.Height = max(1, msg.Height-14)

	case runner.LogMsg:
		m.logs = append(m.logs, msg.Text)

		isAtBottom := m.viewport.AtBottom()

		if m.searchQuery != "" {
			if strings.Contains(strings.ToLower(msg.Text), strings.ToLower(m.searchQuery)) {
				m.filteredLogs = append(m.filteredLogs, highlightSearchTerm(msg.Text, m.searchQuery))
				m.viewport.SetContent(strings.Join(m.filteredLogs, "\n"))
			}
		} else {
			m.viewport.SetContent(strings.Join(m.logs, "\n"))
		}

		if isAtBottom {
			m.viewport.GotoBottom()
		}

	case runner.ExecFinishedMsg:
		m.isRunning = false
		if msg.Err != nil {
			m.logs = append(m.logs, fmt.Sprintf("Error: %v", msg.Err))
			m.err = msg.Err
		} else {
			m.logs = append(m.logs, "Process Finished Successfully. (Press ESC to return)")
		}
		m.viewport.SetContent(strings.Join(m.logs, "\n"))
		m.viewport.GotoBottom()
	}

	// Update viewport if we are looking at logs
	if m.state == stateCommandLog || m.state == stateSetupLog {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	header := TitleStyle.Render(" 🚀 Stocat Commander ")

	var content string

	switch m.state {
	case stateProjectList:
		content = m.renderProjectList()
	case stateCommandList:
		content = m.renderCommandList()
	case stateCommandLog:
		content = m.renderCommandLog()
	case stateSetupLog:
		content = m.renderSetupLog()
	}

	// Help or status text at the bottom
	var statusText string
	switch m.state {
	case stateProjectList:
		statusText = "↑/↓: Navigate • Enter: Select (will clone if missing) • q: Quit"
	case stateCommandList:
		statusText = "↑/↓: Navigate • Enter: Run • esc: Back • q: Quit"
	case stateCommandLog, stateSetupLog:
		if m.isRunning {
			statusText = "Running... • esc: Back to Menu (Run in background) • ctrl+c: Quit"
		} else {
			statusText = "Done • esc: Back • q: Quit"
		}
	}

	footer := StatusStyle.Render(statusText)

	return AppStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			"\n",
			content,
			"\n",
			footer,
		),
	)
}

func (m Model) renderProjectList() string {
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Render("Select a Project:") + "\n\n")

	for i, proj := range m.projects {
		cursor := "  "
		style := ItemStyle
		if m.projectCursor == i {
			cursor = SelectedItemStyle.String()
			style = SelectedItemStyle
		}

		s.WriteString(cursor + style.Render(proj.Name))

		// Indicate if missing
		if _, err := os.Stat(proj.Path); os.IsNotExist(err) {
			s.WriteString(lipgloss.NewStyle().Foreground(colorError).Render(" (Not Cloned)"))
		}

		s.WriteString(" " + DescStyle.Render(proj.Description) + "\n")
	}

	return ListStyle.Width(max(1, m.width-8)).Height(max(1, m.height-12)).Render(s.String())
}

func (m Model) renderSetupLog() string {
	proj := m.projects[m.projectCursor]

	title := lipgloss.NewStyle().Bold(true).Foreground(colorHighlight).Render(fmt.Sprintf(" Setup > Cloning %s ", proj.Name))

	var statusBanner string
	if m.isRunning {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FFFF")).
			Bold(true).
			Width(max(1, m.width-8)).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("\n%s\n\n현재 저장소를 클론하고 있습니다...", m.spinner.View()))
	} else if m.err != nil {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#EF4444")).
			Bold(true).
			Width(max(1, m.width-8)).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			Render("\n❌\n\n클론 실패 (ERROR) - 아래 로그를 확인해주세요.")
	} else {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#10B981")).
			Bold(true).
			Width(max(1, m.width-8)).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			Render("\n✅\n\n클론 완료! (ESC)를 눌러 돌아가세요.")
	}

	var searchBar string
	if m.isSearching {
		searchBar = "\n" + m.searchInput.View()
	} else if m.searchQuery != "" {
		searchBar = fmt.Sprintf("\nSearch filter active: '%s' (press / to change, Esc to clear by going back)", m.searchQuery)
	} else {
		searchBar = "\nPress / to search logs | ↑/↓ to scroll"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		statusBanner,
		OutputStyle.Width(max(1, m.width-8)).Height(max(1, m.height-16)).Render(m.viewport.View()),
		searchBar,
	)
}

func (m Model) renderCommandList() string {
	proj := m.projects[m.projectCursor]

	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("Select Command for '%s':", proj.Name)) + "\n\n")

	var runDir string // Extract selected command run dir

	for i, cmd := range proj.Commands {
		if cmd.IsSeparator {
			s.WriteString("\n") // Add empty line spacing
			continue
		}

		cursor := "  "
		style := ItemStyle
		if m.commandCursor == i {
			cursor = SelectedItemStyle.String()
			style = SelectedItemStyle

			// Highlighted command's directory
			runDir = proj.Path
			if cmd.DirName != "" {
				runDir = filepath.Join(m.workspaceDir, cmd.DirName)
			}
		}

		s.WriteString(cursor + style.Render(cmd.Name) + " - " + DescStyle.Render(cmd.Description) + "\n")
	}

	listWidth := max(1, (m.width-12)/2)
	detailWidth := max(1, m.width-12-listWidth)

	listBlock := ListStyle.Width(listWidth).Height(max(1, m.height-12)).Render(s.String())

	detailBlock := DetailStyle.Width(detailWidth).Height(max(1, m.height-12)).Render(
		fmt.Sprintf("Project: %s\nTarget Path: %s\n\nPress Enter to run.", proj.Name, runDir),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, listBlock, detailBlock)
}

func (m Model) renderCommandLog() string {
	proj := m.projects[m.projectCursor]
	cmd := proj.Commands[m.commandCursor]

	title := lipgloss.NewStyle().Bold(true).Foreground(colorHighlight).Render(fmt.Sprintf(" %s > %s ", proj.Name, cmd.Name))

	var statusBanner string
	if m.isRunning {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FFFF")). // Bright Cyan
			Bold(true).
			Width(max(1, m.width-8)).
			Height(5). // BIG format
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("\n🚀 %s 🚀\n\n현재 명령어를 열심히 실행 중입니다...", m.spinner.View()))
	} else if m.err != nil {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#EF4444")). // Red
			Bold(true).
			Width(max(1, m.width-8)).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			Render(fmt.Sprintf("\n💥 ❌ 💥\n\n실행 실패 (ERROR): %v", m.err))
	} else {
		statusBanner = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#10B981")). // Green
			Bold(true).
			Width(max(1, m.width-8)).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			Render("\n🎉 ✅ 🎉\n\n실행이 성공적으로 완료되었습니다! (ESC)를 눌러 돌아가세요.")
	}

	var searchBar string
	if m.isSearching {
		searchBar = "\n" + m.searchInput.View()
	} else if m.searchQuery != "" {
		searchBar = fmt.Sprintf("\nSearch filter active: '%s' (press / to change, Esc to clear by going back)", m.searchQuery)
	} else {
		searchBar = "\nPress / to search logs | ↑/↓ to scroll"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		statusBanner,
		OutputStyle.Width(max(1, m.width-8)).Height(max(1, m.height-16)).Render(m.viewport.View()),
		searchBar,
	)
}

func getDefaultWorkspace() string {
	cwd, _ := os.Getwd()
	// Generally we assume stocat-commander is in the workspace directory
	// so the parent directory is the workspace
	parent := cwd + "/.."
	if ws := os.Getenv("WORKSPACE"); ws != "" {
		parent = ws
	}
	return parent
}
