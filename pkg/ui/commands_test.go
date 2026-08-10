package ui

import (
	"context"
	"strings"
	"testing"
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
