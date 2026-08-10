package ui

import (
	"context"
	"strings"
	"testing"

	"flashwhip/pkg/config"
)

func TestCommandRegistry_ExitAliases(t *testing.T) {
	reg := DefaultRegistry()
	cmdCtx := &CommandContext{
		Ctx: context.Background(),
	}

	exitInputs := []string{"exit", "quit", "/exit", "/quit", "/EXIT", "QUIT"}
	for _, input := range exitInputs {
		res := reg.Dispatch(input, cmdCtx)
		if !res.Handled {
			t.Errorf("Expected input %q to be handled by command registry", input)
		}
		if !res.ShouldExit {
			t.Errorf("Expected input %q to return ShouldExit = true", input)
		}
	}
}

func TestCommandRegistry_NonCommandInput(t *testing.T) {
	reg := DefaultRegistry()
	cmdCtx := &CommandContext{
		Ctx: context.Background(),
	}

	prompts := []string{"hello world", "how do I write a Go server?", "refactor this function"}
	for _, prompt := range prompts {
		res := reg.Dispatch(prompt, cmdCtx)
		if res.Handled {
			t.Errorf("Expected user prompt %q NOT to be handled by command registry", prompt)
		}
	}
}

func TestCommandRegistry_UnknownSlashCommand(t *testing.T) {
	reg := DefaultRegistry()
	cmdCtx := &CommandContext{
		Ctx: context.Background(),
	}

	res := reg.Dispatch("/unknowncommand", cmdCtx)
	if !res.Handled {
		t.Errorf("Expected unknown slash command to be handled (intercepted) by registry")
	}
	if res.ShouldExit {
		t.Errorf("Expected unknown command NOT to exit")
	}
}

func TestCommandRegistry_CustomRegistration(t *testing.T) {
	reg := NewCommandRegistry()
	executed := false

	reg.Register(Command{
		Name:        "/custom",
		Aliases:     []string{"c"},
		Description: "Test custom command",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			executed = true
			if len(args) != 2 || args[0] != "foo" || args[1] != "bar" {
				t.Errorf("Unexpected args: %v", args)
			}
			return CommandResult{}, nil
		},
	})

	res := reg.Dispatch("/custom foo bar", &CommandContext{})
	if !res.Handled || !executed {
		t.Errorf("Failed to dispatch custom command")
	}

	executed = false
	resAlias := reg.Dispatch("c foo bar", &CommandContext{})
	if !resAlias.Handled || !executed {
		t.Errorf("Failed to dispatch custom command via alias")
	}
}

func TestCommandRegistry_HelpMenu(t *testing.T) {
	reg := DefaultRegistry()
	help := reg.HelpMenu()

	if !strings.Contains(help, "FLASHWHIP REPL COMMANDS") {
		t.Errorf("Help menu missing title header")
	}
	if !strings.Contains(help, "/exit, /quit, exit, quit") {
		t.Errorf("Help menu missing exit aliases")
	}
}

func TestCommandRegistry_TokensCommand(t *testing.T) {
	reg := DefaultRegistry()
	cmdCtx := &CommandContext{
		Ctx:    context.Background(),
		Config: &config.Config{ModelName: "qwen2.5-coder:7b"},
	}

	for _, cmdStr := range []string{"/tokens", "/usage", "/context"} {
		res := reg.Dispatch(cmdStr, cmdCtx)
		if !res.Handled {
			t.Errorf("Expected %q to be handled", cmdStr)
		}
		if res.Error != nil {
			t.Errorf("Unexpected error dispatching %q: %v", cmdStr, res.Error)
		}
	}
}

func TestCommandRegistry_NewCommand(t *testing.T) {
	reg := DefaultRegistry()
	sessionID := "chat-1234567890"
	cmdCtx := &CommandContext{
		Ctx:       context.Background(),
		SessionID: &sessionID,
	}

	res := reg.Dispatch("/new", cmdCtx)
	if !res.Handled {
		t.Errorf("Expected /new to be handled")
	}
	if res.Error != nil {
		t.Errorf("Unexpected error dispatching /new: %v", res.Error)
	}
	if *cmdCtx.SessionID == "chat-1234567890" {
		t.Errorf("Expected session ID to be updated after /new")
	}
	if !strings.HasPrefix(*cmdCtx.SessionID, "chat-") {
		t.Errorf("Expected new session ID to start with 'chat-'")
	}
}

func TestCommandRegistry_ResetCommand(t *testing.T) {
	reg := DefaultRegistry()
	sessionID := "chat-1234567890"
	cmdCtx := &CommandContext{
		Ctx:       context.Background(),
		SessionID: &sessionID,
	}

	res := reg.Dispatch("/reset", cmdCtx)
	if !res.Handled {
		t.Errorf("Expected /reset to be handled")
	}
	if res.Error != nil {
		t.Errorf("Unexpected error dispatching /reset: %v", res.Error)
	}
	if *cmdCtx.SessionID == "chat-1234567890" {
		t.Errorf("Expected session ID to be updated after /reset")
	}
	if !strings.HasPrefix(*cmdCtx.SessionID, "chat-") {
		t.Errorf("Expected new session ID to start with 'chat-'")
	}
}
