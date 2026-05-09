// Package messages defines all Bubble Tea message types shared across
// Forge UI components. This package has no internal dependencies to
// avoid import cycles.
package messages

import "time"

// ExecOutputMsg is emitted by the exec goroutine each time the pipe
// yields a non-empty read from a running command's stdout/stderr.
type ExecOutputMsg struct {
	BlockID string // identifies the target block
	Data    []byte // raw bytes; may be partial lines
}

// ExecDoneMsg is emitted by the exec goroutine when the command exits
// and the output pipe is fully drained.
type ExecDoneMsg struct {
	BlockID  string
	ExitCode int           // 0 = success, non-zero = failure
	Duration time.Duration // wall-clock elapsed time
}

// ACPTokenMsg is emitted by the ACP streaming goroutine for each
// token received from the agent. Consumers must append, not replace.
type ACPTokenMsg struct {
	BlockID string // identifies which AI card receives this token
	Token   string // a single text fragment; may be a partial word
}

// ACPDoneMsg is emitted by the ACP streaming goroutine when the
// agent signals end-of-stream or when an error terminates the call.
type ACPDoneMsg struct {
	BlockID string
	Err     error // nil on clean completion; non-nil on transport or agent error
}

// BlockFocusedMsg is emitted by the viewport when the user navigates
// focus with arrow keys or j/k. An empty BlockID means no block is focused.
type BlockFocusedMsg struct {
	BlockID string
}

// CwdChangedMsg is emitted by session.Run() when a cd command
// completes and the working directory has changed.
type CwdChangedMsg struct {
	Cwd string // absolute path of the new working directory
}

// OpenPaletteMsg is emitted by the input bar when the user presses
// ctrl+k. The root model switches rendering to the palette overlay.
type OpenPaletteMsg struct{}

// ViewportClearMsg is emitted when the user runs the `clear` built-in.
// The viewport drops all rendered blocks; the session ring buffer is also reset.
type ViewportClearMsg struct{}

// SubmitMsg is emitted by the input bar when the user presses Enter.
// IsAIPrompt is true when Input starts with "/" or "@".
type SubmitMsg struct {
	Input      string
	IsAIPrompt bool
}

// AcceptAIMsg is emitted by the AI card when the user presses Enter
// to accept the suggested command.
type AcceptAIMsg struct {
	Command string
}

// DismissAIMsg is emitted by the AI card when the user presses Esc or "d".
type DismissAIMsg struct{}

// ClosePaletteMsg is emitted by the palette when the user presses Esc or
// selects an item, signalling the root model to close the palette overlay.
type ClosePaletteMsg struct{}

// CopyLastOutputMsg is emitted by the command palette when the user selects
// "Copy last output".
type CopyLastOutputMsg struct{}

// FixWithAIMsg is emitted by the command palette when the user selects
// "Fix with AI".
type FixWithAIMsg struct{}
