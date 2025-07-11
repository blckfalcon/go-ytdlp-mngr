package url

import (
	"io"
	"os"
	"os/exec"
)

// CommandExecutor defines the interface for executing external commands
type CommandExecutor interface {
	CreateCommand(name string, args ...string) Command
}

// ProcessState defines the interface for process state information
type ProcessState interface {
	Exited() bool
}

// Command defines the interface for a command that can be executed
type Command interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	GetProcess() *os.Process
	GetProcessState() ProcessState
}

// RealCommandExecutor implements CommandExecutor using actual exec.Command
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) CreateCommand(name string, args ...string) Command {
	return &RealCommand{cmd: exec.Command(name, args...)}
}

// RealCommand wraps exec.Cmd to implement the Command interface
type RealCommand struct {
	cmd *exec.Cmd
}

func (r *RealCommand) Start() error {
	return r.cmd.Start()
}

func (r *RealCommand) Wait() error {
	return r.cmd.Wait()
}

func (r *RealCommand) StdoutPipe() (io.ReadCloser, error) {
	return r.cmd.StdoutPipe()
}

func (r *RealCommand) StderrPipe() (io.ReadCloser, error) {
	return r.cmd.StderrPipe()
}

func (r *RealCommand) GetProcess() *os.Process {
	return r.cmd.Process
}

func (r *RealCommand) GetProcessState() ProcessState {
	if r.cmd.ProcessState == nil {
		return nil
	}
	return &RealProcessState{state: r.cmd.ProcessState}
}

// RealProcessState wraps os.ProcessState to implement ProcessState interface
type RealProcessState struct {
	state *os.ProcessState
}

func (r *RealProcessState) Exited() bool {
	return r.state.Exited()
}
