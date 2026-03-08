package runner

import (
	"bufio"
	"io"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// LogMsg represents a single line of log output
type LogMsg struct {
	Text string
}

// ExecFinishedMsg indicates a process has completed
type ExecFinishedMsg struct {
	Err error
}

// StartCommand runs a command with specific arguments in the specified directory
// and streams output to the given tea.Program instance.
func StartCommand(dir string, name string, args []string, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir

		// Create pipes for stdout and stderr
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return ExecFinishedMsg{Err: err}
		}

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			return ExecFinishedMsg{Err: err}
		}

		// Combine stdout and stderr read handling using simple goroutines
		streamOutput := func(r io.Reader) {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				text := scanner.Text()
				if p != nil {
					p.Send(LogMsg{Text: text})
				}
			}
		}

		if err := cmd.Start(); err != nil {
			return ExecFinishedMsg{Err: err}
		}

		// Stream asynchronously
		go streamOutput(stdoutPipe)
		go streamOutput(stderrPipe)

		// Wait synchronously in this tea.Cmd loop, which runs inside a goroutine by bubbletea
		err = cmd.Wait()

		// Instead of returning directly, we send the finished message so the UI can update
		if p != nil {
			p.Send(ExecFinishedMsg{Err: err})
		}

		return nil
	}
}
