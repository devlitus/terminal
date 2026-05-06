package viewport

import (
	"fmt"
	"strings"

	bubblesviewport "github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	datablock "github.com/forge-tui/forge/internal/block"
	"github.com/forge-tui/forge/internal/messages"
	blockui "github.com/forge-tui/forge/internal/ui/block"
)

// Model is the scrollable container that stacks command block views vertically.
// It owns focus state and maps keyboard navigation to BlockFocusedMsg events.
type Model struct {
	blocks     []blockui.Model
	blockIDs   []uint64 // parallel to blocks; tracks each block's ID for focus events
	vp         bubblesviewport.Model
	focusedIdx int
	autoScroll bool
	width      int
	height     int
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
		m.vp.Height = msg.Height - 2 // reserve one row for header, one for input bar
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
		}
	}

	var vpCmd tea.Cmd
	m.vp, vpCmd = m.vp.Update(msg)
	cmds = append(cmds, vpCmd)

	return m, tea.Batch(cmds...)
}

// View renders all blocks inside the bubbles viewport widget.
func (m Model) View() string {
	return m.vp.View()
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
