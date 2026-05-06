package block

import "time"

// BlockState represents the execution state of a command block.
type BlockState string

const (
	StateRunning BlockState = "running"
	StateSuccess BlockState = "success"
	StateFailed  BlockState = "failed"
)

// AICard holds the AI suggestion associated with a block.
type AICard struct {
	Prompt    string
	Tokens    []string
	Streaming bool
	Accepted  bool
	Dismissed bool
}

// Block represents a single executed command and its output.
type Block struct {
	ID        uint64
	Command   string
	Dir       string
	StartedAt time.Time
	Duration  time.Duration
	ExitCode  int
	State     BlockState
	Output    []string
	AICard    *AICard
}
