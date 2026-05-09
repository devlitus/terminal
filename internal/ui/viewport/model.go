package viewport

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	bubblesviewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
	blockui "github.com/forge-tui/forge/internal/ui/block"
)

// Model is the scrollable container that stacks command block views vertically.
// It owns focus state and maps keyboard navigation to BlockFocusedMsg events.
type Model struct {
	blocks       []blockui.Model
	blockIDs     []uint64 // parallel to blocks; tracks each block's ID for focus events
	vp           bubblesviewport.Model
	focusedIdx   int
	autoScroll   bool
	width        int
	height       int
	statusNotice string
}

// New returns a Model with no blocks, no focus, and autoScroll enabled.
func New() Model {
	return Model{
		vp:         bubblesviewport.New(0, 0),
		focusedIdx: -1,
		autoScroll: true,
	}
}

// AppendBlock adds a block to the viewport. When autoScroll is true, focus
// moves to the new block and the viewport scrolls to the bottom.
func (m *Model) AppendBlock(b datablock.Block) {
	bm := blockui.New(b)
	if m.width > 0 {
		bm.SetWidth(m.width)
	}
	m.blocks = append(m.blocks, bm)
	m.blockIDs = append(m.blockIDs, b.ID)

	if m.autoScroll {
		m.focusedIdx = len(m.blocks) - 1
		m.vp.SetContent(m.renderBlocks())
		m.vp.GotoBottom()
	}
}

// Init satisfies tea.Model; this component has no startup commands.
func (m Model) Init() tea.Cmd { return nil }

// Update handles terminal resize, viewport scroll delegation, and j/k/arrow navigation.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height // height is pre-calculated by the root model
		for i := range m.blocks {
			m.blocks[i].SetWidth(msg.Width)
		}
		m.vp.SetContent(m.renderBlocks())

	case messages.ExecOutputMsg:
		if idx := m.findBlock(msg.BlockID); idx >= 0 {
			m.blocks[idx].AppendOutput(msg.Data)
			m.vp.SetContent(m.renderBlocks())
			if m.autoScroll {
				m.vp.GotoBottom()
			}
		}
		return m, nil

	case messages.ExecDoneMsg:
		if idx := m.findBlock(msg.BlockID); idx >= 0 {
			m.blocks[idx].SetDone(msg.ExitCode, msg.Duration)
			m.vp.SetContent(m.renderBlocks())
		}
		return m, nil

	case messages.ACPTokenMsg:
		if idx := m.findBlock(msg.BlockID); idx >= 0 {
			m.blocks[idx].AppendAIToken(msg.Token)
			m.vp.SetContent(m.renderBlocks())
			if m.autoScroll {
				m.vp.GotoBottom()
			}
		}
		return m, nil

	case messages.ACPDoneMsg:
		if idx := m.findBlock(msg.BlockID); idx >= 0 {
			m.blocks[idx].SetAIStreamDone()
			m.vp.SetContent(m.renderBlocks())
		}
		return m, nil

	case messages.DismissAIMsg:
		for i := range m.blocks {
			blk := m.blocks[i].Block()
			if blk.AICard != nil && !blk.AICard.Dismissed {
				m.blocks[i].DismissAICard()
				m.vp.SetContent(m.renderBlocks())
				break
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.focusedIdx < len(m.blocks)-1 {
				m.focusedIdx++
				if m.focusedIdx == len(m.blocks)-1 {
					m.autoScroll = true
				}
				m.vp.SetContent(m.renderBlocks())
				cmds = append(cmds, emitBlockFocused(m.blockIDs[m.focusedIdx]))
			}
		case "k", "up":
			if m.focusedIdx > 0 {
				m.focusedIdx--
				m.autoScroll = false
				m.vp.SetContent(m.renderBlocks())
				cmds = append(cmds, emitBlockFocused(m.blockIDs[m.focusedIdx]))
			}
		case "y":
			if m.focusedIdx >= 0 {
				blk := m.blocks[m.focusedIdx].Block()
				text := strings.Join(blk.Output, "\n")
				if err := clipboard.WriteAll(text); err != nil {
					m.statusNotice = "clipboard unavailable"
				} else {
					m.statusNotice = ""
				}
			}
		case "r":
			if m.focusedIdx >= 0 {
				blk := m.blocks[m.focusedIdx].Block()
				if blk.State != datablock.StateRunning {
					cmd := blk.Command
					cmds = append(cmds, func() tea.Msg {
						return messages.SubmitMsg{Input: cmd, IsAIPrompt: false}
					})
				}
			}
		case "enter", "esc", "d":
			if m.focusedIdx >= 0 {
				blk := m.blocks[m.focusedIdx].Block()
				if blk.AICard != nil && !blk.AICard.Dismissed && !blk.AICard.Streaming {
					raw, cmd := m.blocks[m.focusedIdx].Update(msg)
					m.blocks[m.focusedIdx] = raw.(blockui.Model)
					return m, cmd
				}
			}
		}
	}

	var vpCmd tea.Cmd
	m.vp, vpCmd = m.vp.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

// View renders all blocks inside the bubbles viewport widget.
// When a statusNotice is set (e.g. clipboard unavailable), it is appended below.
func (m Model) View() string {
	if m.statusNotice != "" {
		notice := lipgloss.NewStyle().Foreground(lipgloss.Color("#5a6178")).Render(m.statusNotice)
		return m.vp.View() + "\n" + notice
	}
	return m.vp.View()
}

// InitAICard initialises a streaming AICard on the block identified by blockID.
func (m *Model) InitAICard(blockID string) {
	if idx := m.findBlock(blockID); idx >= 0 {
		m.blocks[idx].StartAIStreaming()
		m.vp.SetContent(m.renderBlocks())
	}
}

// SetAIError marks the AI card on the given block as errored and stops streaming.
func (m *Model) SetAIError(blockID, errMsg string) {
	if idx := m.findBlock(blockID); idx >= 0 {
		m.blocks[idx].SetAIError(errMsg)
		m.vp.SetContent(m.renderBlocks())
	}
}

// Blocks returns the block model slice (used by tests).
func (m Model) Blocks() []blockui.Model { return m.blocks }

// FocusedIdx returns the index of the focused block (-1 means none).
func (m Model) FocusedIdx() int { return m.focusedIdx }

// renderBlocks produces the full content string handed to the viewport.
// Each block is copied locally so that SetFocused does not mutate the slice.
func (m Model) renderBlocks() string {
	if len(m.blocks) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, blk := range m.blocks {
		b := blk // local copy — SetFocused has a pointer receiver; operates on the copy
		b.SetFocused(i == m.focusedIdx)
		sb.WriteString(b.View())
		if i < len(m.blocks)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func emitBlockFocused(id uint64) tea.Cmd {
	return func() tea.Msg {
		return messages.BlockFocusedMsg{BlockID: fmt.Sprintf("%d", id)}
	}
}

// findBlock returns the index of the block whose ID matches blockID, or -1.
func (m Model) findBlock(blockID string) int {
	for i, id := range m.blockIDs {
		if fmt.Sprintf("%d", id) == blockID {
			return i
		}
	}
	return -1
}
