package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultCommandTimeout = 60 * time.Second

type CommandHandler struct{}

type executeCommandRequest struct {
	Command          string `json:"command"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	WorkingDirectory string `json:"working_directory"`
	UseSudo          bool   `json:"use_sudo"`
	SudoPassword     string `json:"sudo_password"`
}

type executeCommandResponse struct {
	Success      bool   `json:"success"`
	Command      string `json:"command"`
	ExitCode     int    `json:"exit_code"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	DurationMS   int64  `json:"duration_ms"`
	TimedOut     bool   `json:"timed_out"`
	UsedSudo     bool   `json:"used_sudo"`
	Shell        string `json:"shell"`
	WorkingDir   string `json:"working_directory,omitempty"`
	ErrorMessage string `json:"error,omitempty"`
}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

func (h *CommandHandler) ExecuteCommand(c *gin.Context) {
	var req executeCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	commandText := strings.TrimSpace(req.Command)
	if commandText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "command is required"})
		return
	}

	timeout := defaultCommandTimeout
	if req.TimeoutSeconds > 0 {
		timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	startedAt := time.Now()
	result := runCommand(ctx, req)
	result.Command = commandText
	result.DurationMS = time.Since(startedAt).Milliseconds()

	statusCode := http.StatusOK
	if result.TimedOut {
		statusCode = http.StatusRequestTimeout
	} else if result.ErrorMessage != "" {
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, result)
}

func runCommand(ctx context.Context, req executeCommandRequest) executeCommandResponse {
	commandText := strings.TrimSpace(req.Command)
	shell, args := buildShellCommand(commandText)
	usedSudo := shouldWrapWithSudo(req)

	var cmd *exec.Cmd
	if usedSudo {
		shell, args = buildSudoShellCommand(commandText)
		cmd = exec.CommandContext(ctx, shell, args...)
	} else {
		cmd = exec.CommandContext(ctx, shell, args...)
	}

	if req.WorkingDirectory != "" {
		cmd.Dir = req.WorkingDirectory
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if usedSudo && req.SudoPassword != "" {
		cmd.Stdin = strings.NewReader(req.SudoPassword + "\n")
	}

	err := cmd.Run()
	result := executeCommandResponse{
		Success:    err == nil,
		ExitCode:   extractExitCode(err),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		TimedOut:   errors.Is(ctx.Err(), context.DeadlineExceeded),
		UsedSudo:   usedSudo,
		Shell:      shell,
		WorkingDir: req.WorkingDirectory,
	}

	if err != nil && !isExitError(err) && !result.TimedOut {
		result.ErrorMessage = err.Error()
	}

	if result.TimedOut && result.ErrorMessage == "" {
		result.ErrorMessage = "command timed out"
	}

	return result
}

func buildShellCommand(commandText string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{"-Command", commandText}
	}
	return "sh", []string{"-lc", commandText}
}

func buildSudoShellCommand(commandText string) (string, []string) {
	if runtime.GOOS == "windows" {
		return buildShellCommand(commandText)
	}
	return "sudo", []string{"-S", "--", "sh", "-lc", commandText}
}

func shouldWrapWithSudo(req executeCommandRequest) bool {
	if runtime.GOOS == "windows" {
		return false
	}
	return req.UseSudo || strings.TrimSpace(req.SudoPassword) != ""
}

func extractExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return -1
	}

	return -1
}

func isExitError(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}
