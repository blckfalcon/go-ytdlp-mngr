package url

import (
	"errors"
	"slices"
	"testing"
	"time"
)

// TestUrlItem_Start_Basic tests basic functionality
func TestUrlItem_Start_Basic(t *testing.T) {
	t.Run("successful_start", func(t *testing.T) {
		mockExecutor := NewMockCommandExecutor()
		urlItem := NewUrlItemEx("https://example.com/video", mockExecutor)

		urlItem.Start()

		cmd := mockExecutor.Command
		if cmd.Name != "yt-dlp" {
			t.Errorf("Expected command name 'yt-dlp', got '%s'", cmd.Name)
		}

		if !slices.Contains(cmd.Args, "https://example.com/video") {
			t.Error("URL not found in command arguments")
		}

		if urlItem.Recording != StageDownloading {
			t.Errorf("Expected stage StageDownloading, got %v", urlItem.Recording)
		}

		if urlItem.StartedAt.IsZero() {
			t.Error("StartedAt should be set")
		}
	})

	t.Run("constructor_tests", func(t *testing.T) {
		mockExecutor := NewMockCommandExecutor()
		urlItem := NewUrlItemEx("https://example.com/video", mockExecutor)

		if urlItem.Url != "https://example.com/video" {
			t.Errorf("Expected URL 'https://example.com/video', got '%s'", urlItem.Url)
		}

		if urlItem.executor != mockExecutor {
			t.Error("Expected executor to be set to provided mock")
		}

		urlItem2 := NewUrlItem("https://example.com/video3")
		if _, ok := urlItem2.executor.(*RealCommandExecutor); !ok {
			t.Error("Expected RealCommandExecutor for defaults")
		}
	})
}

// TestUrlItem_Start_PreConfiguredMocks tests with pre-configured mock behavior
func TestUrlItem_Start_PreConfiguredMocks(t *testing.T) {
	t.Run("start_error", func(t *testing.T) {
		mockExecutor := NewMockCommandExecutor()

		// Pre-configure a command that will fail to start
		failingCmd := &MockCommand{
			Name:     "yt-dlp",
			Args:     []string{"-f", "best[height<=1080]", "--fixup", "warn", "-4", "https://example.com/video"},
			StartErr: errors.New("command not found"),
		}

		originalCreateCommand := mockExecutor.CreateCommandFunc
		mockExecutor.CreateCommandFunc = func(name string, args ...string) Command {
			return failingCmd
		}
		defer func() { mockExecutor.CreateCommandFunc = originalCreateCommand }()

		urlItem := NewUrlItemEx("https://example.com/video", mockExecutor)
		urlItem.Start()

		if urlItem.Recording != StageNotStarted {
			t.Errorf("Expected stage StageNotStarted on start failure, got %v", urlItem.Recording)
		}
	})

	t.Run("wait_error", func(t *testing.T) {
		mockExecutor := NewMockCommandExecutor()

		// Pre-configure a command that starts but fails during wait
		failingCmd := &MockCommand{
			Name:         "yt-dlp",
			Args:         []string{"-f", "best[height<=1080]", "--fixup", "warn", "-4", "https://example.com/video"},
			StartErr:     nil,
			WaitErr:      errors.New("process failed"),
			waitDuration: 50 * time.Millisecond,
		}

		originalCreateCommand := mockExecutor.CreateCommandFunc
		mockExecutor.CreateCommandFunc = func(name string, args ...string) Command {
			return failingCmd
		}
		defer func() { mockExecutor.CreateCommandFunc = originalCreateCommand }()

		urlItem := NewUrlItemEx("https://example.com/video", mockExecutor)
		urlItem.Start()

		time.Sleep(100 * time.Millisecond)

		if urlItem.Recording != StageError {
			t.Errorf("Expected stage StageError on wait failure, got %v", urlItem.Recording)
		}
	})

	t.Run("successful_completion", func(t *testing.T) {
		mockExecutor := NewMockCommandExecutor()

		// Pre-configure a command that completes successfully
		successCmd := &MockCommand{
			Name:         "yt-dlp",
			Args:         []string{"-f", "best[height<=1080]", "--fixup", "warn", "-4", "https://example.com/video"},
			StartErr:     nil,
			WaitErr:      nil,
			waitDuration: 50 * time.Millisecond,
		}

		originalCreateCommand := mockExecutor.CreateCommandFunc
		mockExecutor.CreateCommandFunc = func(name string, args ...string) Command {
			return successCmd
		}
		defer func() { mockExecutor.CreateCommandFunc = originalCreateCommand }()

		urlItem := NewUrlItemEx("https://example.com/video", mockExecutor)
		urlItem.Start()

		time.Sleep(100 * time.Millisecond)

		if urlItem.Recording != StageCompleted {
			t.Errorf("Expected stage StageCompleted on success, got %v", urlItem.Recording)
		}

		if urlItem.StoppedAt.IsZero() {
			t.Error("StoppedAt should be set after completion")
		}
	})
}

// TestUrlItem_Start_ArgumentValidation tests command argument construction
func TestUrlItem_Start_ArgumentValidation(t *testing.T) {
	mockExecutor := NewMockCommandExecutor()
	urlItem := NewUrlItemEx("https://example.com/test-video", mockExecutor)

	urlItem.Start()

	cmd := mockExecutor.Command
	expectedArgs := []string{"-f", "best[height<=1080]", "--fixup", "warn", "-4", "https://example.com/test-video"}

	if len(cmd.Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(cmd.Args))
	}

	for i, expected := range expectedArgs {
		if i >= len(cmd.Args) || cmd.Args[i] != expected {
			t.Errorf("Arg[%d]: expected '%s', got '%s'", i, expected, cmd.Args[i])
		}
	}
}

func TestUrlItem_StateTransition(t *testing.T) {

	t.Run("stage_progression", func(t *testing.T) {
		mockExecutor := NewMockCommandExecutor()
		urlItem := NewUrlItemEx("https://example.com/video", mockExecutor)

		stages := []DownloadStage{}
		stages = append(stages, urlItem.Recording)

		urlItem.Start()
		stages = append(stages, urlItem.Recording)

		// wait for the command to complete
		time.Sleep(500 * time.Millisecond)

		urlItem.Stop()
		stages = append(stages, urlItem.Recording)

		expectedStages := []DownloadStage{
			StageNotStarted,
			StageDownloading,
			StageCompleted,
		}
		for i, expected := range expectedStages {
			if i >= len(stages) || stages[i] != expected {
				t.Errorf("Stage[%d]: expected '%s', got '%s'", i, expected, stages[i])
			}
		}
	})
}
