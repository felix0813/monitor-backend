package handlers

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultCommandTimeout = 20 * time.Second

type CommandHandler struct{}

type executeCommandRequest struct {
	Command          string `json:"command"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	WorkingDirectory string `json:"working_directory"`
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
		log.Printf("command_handler: invalid request body from ip=%s: %v", c.ClientIP(), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	commandText := strings.TrimSpace(req.Command)
	if commandText == "" {
		log.Printf("command_handler: empty command rejected from ip=%s", c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{"error": "command is required"})
		return
	}

	timeout := defaultCommandTimeout
	if req.TimeoutSeconds > 0 {
		timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}

	log.Printf(
		"command_handler: executing command=%q timeout=%s working_dir=%q ip=%s",
		summarizeCommand(commandText),
		timeout,
		req.WorkingDirectory,
		c.ClientIP(),
	)

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

	log.Printf(
		"command_handler: finished command=%q success=%t exit_code=%d timed_out=%t duration_ms=%d status=%d used_sudo=%t",
		summarizeCommand(result.Command),
		result.Success,
		result.ExitCode,
		result.TimedOut,
		result.DurationMS,
		statusCode,
		result.UsedSudo,
	)
	if result.ErrorMessage != "" {
		log.Printf(
			"command_handler: command error command=%q error=%q stderr=%q",
			summarizeCommand(result.Command),
			result.ErrorMessage,
			summarizeOutput(result.Stderr),
		)
	}

	c.JSON(statusCode, result)
}

func runCommand(ctx context.Context, req executeCommandRequest) executeCommandResponse {
	commandText := strings.TrimSpace(req.Command)

	shell, args := buildShellCommand(commandText)
	cmd := exec.CommandContext(ctx, shell, args...)

	if req.WorkingDirectory != "" {
		cmd.Dir = req.WorkingDirectory
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	sudoPassword := os.Getenv("SUDO_PASSWORD")
	needsSudo := containsSudoCommand(commandText)

	if needsSudo && sudoPassword != "" {
		log.Printf("command_handler: sudo password detected for command=%q", summarizeCommand(commandText))
		cmd.Env = append(os.Environ(), "SUDO_PASSWORD="+sudoPassword)
		cmd.Stdin = strings.NewReader(sudoPassword + "\n")
	} else if needsSudo {
		log.Printf("command_handler: sudo command detected without configured password for command=%q", summarizeCommand(commandText))
	}

	err := cmd.Run()
	result := executeCommandResponse{
		Success:    err == nil,
		ExitCode:   extractExitCode(err),
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		TimedOut:   errors.Is(ctx.Err(), context.DeadlineExceeded),
		UsedSudo:   needsSudo,
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

func summarizeCommand(commandText string) string {
	return summarizeText(commandText, 200)
}

func summarizeOutput(output string) string {
	return summarizeText(output, 300)
}

func summarizeText(text string, limit int) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if normalized == "" {
		return ""
	}
	if len(normalized) <= limit {
		return normalized
	}
	return normalized[:limit] + "..."
}

func buildShellCommand(commandText string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "powershell", []string{"-Command", commandText}
	}

	lines := strings.Split(commandText, "\n")
	var processedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "sudo ") {
			actualCmd := strings.TrimPrefix(line, "sudo ")
			processedLines = append(processedLines, "echo \"$SUDO_PASSWORD\" | sudo -S -- sh -c '"+actualCmd+"'")
		} else {
			processedLines = append(processedLines, line)
		}
	}

	combinedCommand := strings.Join(processedLines, "; ")
	return "sh", []string{"-lc", combinedCommand}
}
func containsSudoCommand(commandText string) bool {
	lines := strings.Split(commandText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "sudo ") {
			return true
		}
	}
	return false
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
