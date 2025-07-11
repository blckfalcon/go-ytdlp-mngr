package url

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	Command           *MockCommand
	CreateCommandFunc func(name string, args ...string) Command
}

func NewMockCommandExecutor() *MockCommandExecutor {
	executor := &MockCommandExecutor{
		Command: &MockCommand{},
	}
	// Set default behavior
	executor.CreateCommandFunc = executor.defaultCreateCommand
	return executor
}

func (m *MockCommandExecutor) defaultCreateCommand(name string, args ...string) Command {
	cmd := &MockCommand{
		Name:         name,
		Args:         args,
		StartErr:     nil,
		WaitErr:      nil,
		ExitCode:     0,
		StdoutData:   "",
		StderrData:   "",
		ProcessState: &os.ProcessState{},
		Process:      &os.Process{},
		started:      false,
		waited:       false,
	}
	m.Command = cmd
	return cmd
}

func (m *MockCommandExecutor) CreateCommand(name string, args ...string) Command {
	return m.CreateCommandFunc(name, args...)
}

// MockCommand implements Command interface for testing
type MockCommand struct {
	Name         string
	Args         []string
	StartErr     error
	WaitErr      error
	ExitCode     int
	StdoutData   string
	StderrData   string
	ProcessState *os.ProcessState
	Process      *os.Process

	// Internal state tracking
	started      bool
	waited       bool
	stdout       io.ReadCloser
	stderr       io.ReadCloser
	mutex        sync.Mutex
	waitDuration time.Duration
}

func (m *MockCommand) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.StartErr != nil {
		return m.StartErr
	}

	m.started = true

	// Create mock pipes
	m.stdout = io.NopCloser(strings.NewReader(m.StdoutData))
	m.stderr = io.NopCloser(strings.NewReader(m.StderrData))

	return nil
}

func (m *MockCommand) Wait() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.started {
		return nil
	}

	// Simulate wait duration if set (default to longer duration to prevent immediate completion)
	waitTime := m.waitDuration
	if waitTime == 0 {
		waitTime = 500 * time.Millisecond // Default wait time to simulate real command
	}

	// Release lock during sleep to allow other operations
	m.mutex.Unlock()
	time.Sleep(waitTime)
	m.mutex.Lock()

	m.waited = true
	return m.WaitErr
}

func (m *MockCommand) StdoutPipe() (io.ReadCloser, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.stdout == nil {
		m.stdout = io.NopCloser(strings.NewReader(m.StdoutData))
	}
	return m.stdout, nil
}

func (m *MockCommand) StderrPipe() (io.ReadCloser, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.stderr == nil {
		m.stderr = io.NopCloser(strings.NewReader(m.StderrData))
	}
	return m.stderr, nil
}

func (m *MockCommand) GetProcess() *os.Process {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.started {
		return nil
	}
	return m.Process
}

func (m *MockCommand) GetProcessState() ProcessState {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.waited {
		return nil
	}

	return &MockProcessState{exited: true}
}

// MockProcessState implements ProcessState interface for testing
type MockProcessState struct {
	exited bool
}

func (m *MockProcessState) Exited() bool {
	return m.exited
}

// Helper methods for test setup
func (m *MockCommand) SetStartError(err error) *MockCommand {
	m.StartErr = err
	return m
}

func (m *MockCommand) SetWaitError(err error) *MockCommand {
	m.WaitErr = err
	return m
}

func (m *MockCommand) SetExitCode(code int) *MockCommand {
	m.ExitCode = code
	return m
}

func (m *MockCommand) SetStdoutData(data string) *MockCommand {
	m.StdoutData = data
	return m
}

func (m *MockCommand) SetStderrData(data string) *MockCommand {
	m.StderrData = data
	return m
}

func (m *MockCommand) SetWaitDuration(duration time.Duration) *MockCommand {
	m.waitDuration = duration
	return m
}

func (m *MockCommand) IsStarted() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.started
}

func (m *MockCommand) IsWaited() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.waited
}
