package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/forge-tui/forge/internal/acp"
	"github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/config"
	forgeExec "github.com/forge-tui/forge/internal/exec"
	"github.com/forge-tui/forge/internal/messages"
	"github.com/forge-tui/forge/internal/session"
	"github.com/forge-tui/forge/internal/ui/header"
	"github.com/forge-tui/forge/internal/ui/input"
	paletteui "github.com/forge-tui/forge/internal/ui/palette"
	viewportui "github.com/forge-tui/forge/internal/ui/viewport"
)

// program is set in main() after tea.NewProgram so that exec goroutines
// can call program.Send() for streaming output messages.
var program *tea.Program

type rootModel struct {
	header      header.Model
	input       input.Model
	vp          viewportui.Model
	rb          *block.RingBuffer
	cancels     map[string]context.CancelFunc
	session     *session.Session
	termWidth   int
	termHeight  int
	paletteOpen bool
	palette     paletteui.Model
	acpClient   *acp.Client
	cfg         *config.Config
	shellOnly   bool
	quitting    bool
}

func newRootModel() rootModel {
	rb := block.NewRingBuffer()
	sess := session.New(rb)
	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}
	client := &acp.Client{}
	shellOnly := cfg.IsShellOnly()
	if !shellOnly {
		if err := client.Connect(cfg); err != nil {
			shellOnly = true
		}
	}
	m := rootModel{
		header:    header.New(sess.Cwd),
		input:     input.New(),
		vp:        viewportui.New(),
		rb:        rb,
		cancels:   make(map[string]context.CancelFunc),
		session:   sess,
		acpClient: client,
		cfg:       cfg,
		shellOnly: shellOnly,
	}
	if shellOnly {
		m.header.SetStatusHint("AI offline — add config: ~/.config/forge/config.toml")
	}
	return m
}

// startACPStream returns a tea.Cmd that streams an ACP prompt in a goroutine,
// sending ACPTokenMsg for each token and ACPDoneMsg when finished.
func startACPStream(p *tea.Program, client *acp.Client, blockID, prompt string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := client.SendPrompt(ctx, prompt, func(token string) {
			p.Send(messages.ACPTokenMsg{BlockID: blockID, Token: token})
		})
		return messages.ACPDoneMsg{BlockID: blockID, Err: err}
	}
}

func (m rootModel) Init() tea.Cmd {
	return tea.Batch(m.header.Init(), m.input.Init(), m.vp.Init())
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// When the palette is open, route all key events exclusively to it.
	if m.paletteOpen {
		if _, ok := msg.(tea.KeyMsg); ok {
			raw, cmd := m.palette.Update(msg)
			m.palette = raw.(paletteui.Model)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		raw, c := m.header.Update(msg)
		m.header = raw.(header.Model)
		cmds = append(cmds, c)
		raw2, c2 := m.input.Update(msg)
		m.input = raw2.(input.Model)
		cmds = append(cmds, c2)
		raw3, c3 := m.vp.Update(msg)
		m.vp = raw3.(viewportui.Model)
		cmds = append(cmds, c3)
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.quitting {
				return m, tea.Quit
			}
			if len(m.cancels) > 0 {
				// Cancel all running blocks; append ^C marker to their output.
				for blockID, cancel := range m.cancels {
					program.Send(messages.ExecOutputMsg{BlockID: blockID, Data: []byte("^C")})
					cancel()
					delete(m.cancels, blockID)
				}
				return m, nil
			}
			return m, tea.Quit
		case "q":
			if len(m.cancels) > 0 {
				m.quitting = true
				return m, nil
			}
			return m, tea.Quit
		case "y":
			if m.quitting {
				return m, tea.Quit
			}
		case "n":
			if m.quitting {
				m.quitting = false
				return m, nil
			}
		case "f":
			if m.shellOnly {
				m.header.SetStatusHint("AI offline — add config: ~/.config/forge/config.toml")
				return m, nil
			}
			if m.vp.FocusedIdx() >= 0 {
				focused := m.vp.Blocks()[m.vp.FocusedIdx()].Block()
				if focused.State == block.StateFailed {
					blockID := fmt.Sprintf("%d", focused.ID)
					last20 := focused.Output
					if len(last20) > 20 {
						last20 = last20[len(last20)-20:]
					}
					prompt := fmt.Sprintf("Command: %s\nError output:\n%s\nFix this command.",
						focused.Command, strings.Join(last20, "\n"))
					m.vp.InitAICard(blockID)
					return m, startACPStream(program, m.acpClient, blockID, prompt)
				}
			}
		}

	case messages.OpenPaletteMsg:
		all := m.rb.All()
		history := make([]string, 0, 20)
		for i := len(all) - 1; i >= 0 && len(history) < 20; i-- {
			if all[i].Command != "" {
				history = append(history, all[i].Command)
			}
		}
		m.palette = paletteui.New(history, m.termWidth, m.termHeight)
		m.paletteOpen = true
		return m, nil

	case messages.ClosePaletteMsg:
		m.paletteOpen = false
		return m, nil

	case messages.CopyLastOutputMsg:
		m.paletteOpen = false
		return m, nil

	case messages.FixWithAIMsg:
		m.paletteOpen = false
		if m.shellOnly {
			m.header.SetStatusHint("AI offline — add config: ~/.config/forge/config.toml")
			return m, nil
		}
		if m.vp.FocusedIdx() >= 0 {
			focused := m.vp.Blocks()[m.vp.FocusedIdx()].Block()
			if focused.State == block.StateFailed {
				blockID := fmt.Sprintf("%d", focused.ID)
				last20 := focused.Output
				if len(last20) > 20 {
					last20 = last20[len(last20)-20:]
				}
				prompt := fmt.Sprintf("Command: %s\nError output:\n%s\nFix this command.",
					focused.Command, strings.Join(last20, "\n"))
				m.vp.InitAICard(blockID)
				return m, startACPStream(program, m.acpClient, blockID, prompt)
			}
		}
		return m, nil

	case messages.SubmitMsg:
		m.paletteOpen = false
		if msg.IsAIPrompt {
			stripped := strings.TrimLeft(msg.Input, "/@")
			b := block.Block{
				Command:   msg.Input,
				Dir:       m.session.Cwd,
				State:     block.StateSuccess,
				StartedAt: time.Now(),
			}
			b = m.rb.Add(b)
			m.vp.AppendBlock(b)
			blockID := fmt.Sprintf("%d", b.ID)
			m.vp.InitAICard(blockID)
			return m, startACPStream(program, m.acpClient, blockID, stripped)
		}
		if !msg.IsAIPrompt {
			handled, sessionCmd := m.session.Handle(msg.Input)
			if handled {
				if sessionCmd != nil {
					cmds = append(cmds, sessionCmd)
				}
				return m, tea.Batch(cmds...)
			}
			b := block.Block{
				Command:   msg.Input,
				Dir:       m.session.Cwd,
				State:     block.StateRunning,
				StartedAt: time.Now(),
			}
			b = m.rb.Add(b)
			m.vp.AppendBlock(b)
			blockID := fmt.Sprintf("%d", b.ID)
			ctx, cancel := context.WithCancel(context.Background())
			m.cancels[blockID] = cancel
			return m, forgeExec.Start(program, blockID, b.Command, b.Dir, ctx)
		}

	case messages.ExecOutputMsg:
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.ExecDoneMsg:
		if cancel, ok := m.cancels[msg.BlockID]; ok {
			cancel()
			delete(m.cancels, msg.BlockID)
		}
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.ACPTokenMsg:
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.ACPDoneMsg:
		if msg.Err != nil {
			m.shellOnly = true
			m.header.SetStatusHint("AI offline — add config: ~/.config/forge/config.toml")
			m.vp.SetAIError(msg.BlockID, "AI error — try again")
			return m, nil
		}
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.AcceptAIMsg:
		// Dismiss the AI card, then shell-exec the suggested command.
		rawVp, c := m.vp.Update(messages.DismissAIMsg{})
		m.vp = rawVp.(viewportui.Model)
		cmds = append(cmds, c)
		b := block.Block{
			Command:   msg.Command,
			Dir:       m.session.Cwd,
			State:     block.StateRunning,
			StartedAt: time.Now(),
		}
		b = m.rb.Add(b)
		m.vp.AppendBlock(b)
		blockID := fmt.Sprintf("%d", b.ID)
		ctx, cancel := context.WithCancel(context.Background())
		m.cancels[blockID] = cancel
		cmds = append(cmds, forgeExec.Start(program, blockID, b.Command, b.Dir, ctx))
		return m, tea.Batch(cmds...)

	case messages.DismissAIMsg:
		raw, c := m.vp.Update(msg)
		m.vp = raw.(viewportui.Model)
		return m, c

	case messages.ViewportClearMsg:
		m.vp = viewportui.New()
		return m, nil

	case messages.CwdChangedMsg:
		m.session.Cwd = msg.Cwd
		raw, c := m.header.Update(msg)
		m.header = raw.(header.Model)
		return m, c
	}

	// Delegate remaining messages to input and viewport.
	raw, c := m.input.Update(msg)
	m.input = raw.(input.Model)
	cmds = append(cmds, c)

	raw2, c2 := m.vp.Update(msg)
	m.vp = raw2.(viewportui.Model)
	cmds = append(cmds, c2)

	return m, tea.Batch(cmds...)
}

func (m rootModel) View() string {
	if m.paletteOpen {
		overlay := m.palette.View()
		return lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, overlay)
	}
	inputView := m.input.View()
	if m.quitting {
		inputView = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#eab308")).
			Render("Quit? (y/n)")
	}
	return m.header.View() + "\n" + m.vp.View() + "\n" + inputView
}

func main() {
	m := newRootModel()
	program = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
