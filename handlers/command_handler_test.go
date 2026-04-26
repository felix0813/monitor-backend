package handlers

import (
	"runtime"
	"testing"
)

func TestShouldWrapWithSudo(t *testing.T) {
	req := executeCommandRequest{Command: "whoami"}

	if runtime.GOOS == "windows" {
		if shouldWrapWithSudo(executeCommandRequest{UseSudo: true}) {
			t.Fatal("expected sudo wrapper to be disabled on windows")
		}
		return
	}

	if shouldWrapWithSudo(req) {
		t.Fatal("expected sudo wrapper to be disabled by default")
	}
	if !shouldWrapWithSudo(executeCommandRequest{UseSudo: true}) {
		t.Fatal("expected use_sudo to enable sudo wrapper")
	}
	if !shouldWrapWithSudo(executeCommandRequest{SudoPassword: "secret"}) {
		t.Fatal("expected sudo password to enable sudo wrapper")
	}
}

func TestBuildShellCommand(t *testing.T) {
	shell, args := buildShellCommand("echo hello")
	if runtime.GOOS == "windows" {
		if shell != "powershell" {
			t.Fatalf("unexpected shell: %s", shell)
		}
		if len(args) != 2 || args[0] != "-Command" {
			t.Fatalf("unexpected args: %#v", args)
		}
		return
	}

	if shell != "sh" {
		t.Fatalf("unexpected shell: %s", shell)
	}
	if len(args) != 2 || args[0] != "-lc" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildSudoShellCommand(t *testing.T) {
	shell, args := buildSudoShellCommand("echo hello")
	if runtime.GOOS == "windows" {
		if shell != "powershell" {
			t.Fatalf("unexpected shell: %s", shell)
		}
		return
	}

	if shell != "sudo" {
		t.Fatalf("unexpected shell: %s", shell)
	}
	if len(args) != 5 || args[0] != "-S" || args[2] != "sh" || args[3] != "-lc" {
		t.Fatalf("unexpected args: %#v", args)
	}
}
