package handlers

import (
	"runtime"
	"strings"
	"testing"
)

func TestBuildShellCommand(t *testing.T) {
	command := "echo hello"
	shell, args := buildShellCommand(command)

	if runtime.GOOS == "windows" {
		if shell != "powershell" {
			t.Fatalf("unexpected shell: %s", shell)
		}
		if len(args) != 2 || args[0] != "-Command" || args[1] != command {
			t.Fatalf("unexpected args: %#v", args)
		}
		return
	}

	if shell != "bash" {
		t.Fatalf("unexpected shell: %s", shell)
	}
	if len(args) != 2 || args[0] != "-c" || args[1] != command {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildShellCommandWithSudo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping sudo test on windows")
	}

	command := "sudo whoami"
	shell, args := buildShellCommand(command)

	if shell != "bash" {
		t.Fatalf("unexpected shell: %s", shell)
	}
	if len(args) != 2 || args[0] != "-c" {
		t.Fatalf("unexpected args: %#v", args)
	}
	if !strings.Contains(args[1], "sudo -S bash") {
		t.Fatalf("expected sudo wrapper in command, got: %s", args[1])
	}
}

func TestContainsSudoCommand(t *testing.T) {
	tests := []struct {
		cmd      string
		expected bool
	}{
		{"ls -l", false},
		{"sudo ls -l", true},
		{"  sudo ls", true},
		{"echo sudo", false},
		{"line1\nsudo line2", true},
	}

	for _, tt := range tests {
		if got := containsSudoCommand(tt.cmd); got != tt.expected {
			t.Errorf("containsSudoCommand(%q) = %v, want %v", tt.cmd, got, tt.expected)
		}
	}
}

func TestSummarizeText(t *testing.T) {
	text := "  hello   world  "
	expected := "hello world"
	if got := summarizeText(text, 20); got != expected {
		t.Errorf("summarizeText(%q) = %q, want %q", text, got, expected)
	}

	longText := strings.Repeat("a", 100)
	got := summarizeText(longText, 10)
	if len(got) != 13 || !strings.HasSuffix(got, "...") {
		t.Errorf("summarizeText(long) length = %d, got %q", len(got), got)
	}
}
