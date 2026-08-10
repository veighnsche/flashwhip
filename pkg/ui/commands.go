package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"google.golang.org/adk/v2/agent"
	adkmodel "google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"

	fagent "flashwhip/pkg/agent"
	"flashwhip/pkg/agent/middleware"
	"flashwhip/pkg/config"
	"flashwhip/pkg/db"
	"flashwhip/pkg/provider/ollama"
	"flashwhip/pkg/tools"
)

// CommandContext holds the execution context and state references for REPL commands.
type CommandContext struct {
	Ctx         context.Context
	Config      *config.Config
	AppAgent    *agent.Agent
	ActiveModel **ollama.Model
	Runner      **runner.Runner
	SessionSvc  session.Service
	SessionID   *string
	Readline    *readline.Instance
}

// CommandResult represents the outcome of executing a command.
type CommandResult struct {
	Handled    bool
	ShouldExit bool
	Error      error
}

// CommandHandler represents a function executing a specific REPL command.
type CommandHandler func(ctx *CommandContext, args []string) (CommandResult, error)

// Command defines a single REPL command, its aliases, description, and execution logic.
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Handler     CommandHandler
}

// CommandRegistry manages and dispatches REPL commands.
type CommandRegistry struct {
	commands map[string]*Command
	list     []*Command
}

// NewCommandRegistry creates a new empty CommandRegistry.
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*Command),
		list:     make([]*Command, 0),
	}
}

// Register registers a command with its primary name and all aliases.
func (r *CommandRegistry) Register(cmd Command) {
	cmdPtr := &cmd
	r.list = append(r.list, cmdPtr)

	registerKey := func(k string) {
		k = normalizeCommandKey(k)
		if k != "" {
			r.commands[k] = cmdPtr
		}
	}

	registerKey(cmd.Name)
	for _, alias := range cmd.Aliases {
		registerKey(alias)
	}
}

// normalizeCommandKey lowercases and strips leading slashes for key matching.
func normalizeCommandKey(k string) string {
	k = strings.TrimSpace(strings.ToLower(k))
	k = strings.TrimPrefix(k, "/")
	return k
}

// Dispatch parses the user input, looks up the command, and executes it if matched.
func (r *CommandRegistry) Dispatch(input string, ctx *CommandContext) CommandResult {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return CommandResult{Handled: false}
	}

	parts := strings.Fields(trimmed)
	firstWord := parts[0]
	isSlash := strings.HasPrefix(firstWord, "/")

	key := normalizeCommandKey(firstWord)
	cmd, exists := r.commands[key]

	// If it doesn't start with / and is not a registered command alias (like exit/quit), pass to LLM
	if !exists {
		if isSlash {
			fmt.Printf("\033[31m[Error]: Unknown command %q. Type '/help' for available commands.\033[0m\n\n", firstWord)
			return CommandResult{Handled: true}
		}
		return CommandResult{Handled: false}
	}

	res, err := cmd.Handler(ctx, parts[1:])
	res.Handled = true
	res.Error = err
	return res
}

// HelpMenu returns a formatted help menu string of all registered commands.
func (r *CommandRegistry) HelpMenu() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(ThinkingBadge.Render("⚡ FLASHWHIP REPL COMMANDS:"))
	b.WriteString("\n")

	for _, cmd := range r.list {
		label := cmd.Name
		if len(cmd.Aliases) > 0 {
			label += ", " + strings.Join(cmd.Aliases, ", ")
		}
		b.WriteString(fmt.Sprintf("  %-24s %s\n", InfoValue.Render(label), cmd.Description))
	}
	b.WriteString("\n")
	return b.String()
}

// DefaultRegistry constructs and registers all standard REPL commands for flashwhip.
func DefaultRegistry() *CommandRegistry {
	reg := NewCommandRegistry()

	reg.Register(Command{
		Name:        "/help",
		Aliases:     []string{"/h"},
		Description: "Show this help menu",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			fmt.Print(reg.HelpMenu())
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/exit",
		Aliases:     []string{"/quit", "exit", "quit"},
		Description: "Exit interactive session",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			fmt.Println("Goodbye!")
			return CommandResult{ShouldExit: true}, nil
		},
	})

	reg.Register(Command{
		Name:        "/sessions",
		Aliases:     []string{"/list", "/history"},
		Description: "List stored conversation sessions",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			database, err := db.DefaultDB()
			if err != nil {
				fmt.Printf("\033[31m[Error]: Failed to open database: %v\033[0m\n", err)
				return CommandResult{}, err
			}
			sessions, err := database.ListSessions()
			if err != nil {
				fmt.Printf("\033[31m[Error]: Failed to list sessions: %v\033[0m\n", err)
				return CommandResult{}, err
			}
			fmt.Println()
			fmt.Println(RenderSessionList(sessions))
			fmt.Println()
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/load",
		Aliases:     []string{"/reload"},
		Description: "Load and resume a past conversation session",
		Usage:       "/load <session_id>",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			if len(args) < 1 {
				fmt.Println(ThinkingBadge.Render("Usage: /load <session_id>"))
				return CommandResult{}, nil
			}
			targetID := args[0]
			database, err := db.DefaultDB()
			if err != nil {
				fmt.Printf("\033[31m[Error]: Failed to open database: %v\033[0m\n", err)
				return CommandResult{}, err
			}
			sessionInfo, msgs, sErr := database.GetSession(targetID)
			if sErr != nil || sessionInfo == nil {
				fmt.Printf("\033[31m[Error]: Session %q not found in database\033[0m\n", targetID)
				return CommandResult{}, sErr
			}

			contents, cErr := database.GetSessionGenAIContents(targetID)
			if cErr == nil && len(contents) > 0 && ctx.SessionSvc != nil {
				sessResp, crErr := ctx.SessionSvc.Create(ctx.Ctx, &session.CreateRequest{
					AppName:   "flashwhip",
					UserID:    "user",
					SessionID: targetID,
				})
				if crErr == nil && sessResp != nil {
					for _, gc := range contents {
						evt := &session.Event{
							Author: gc.Role,
							LLMResponse: adkmodel.LLMResponse{
								Content: gc,
							},
						}
						_ = ctx.SessionSvc.AppendEvent(ctx.Ctx, sessResp.Session, evt)
					}
				}
			}

			if ctx.SessionID != nil {
				*ctx.SessionID = targetID
			}
			fmt.Printf("\n%s Attached to session %s (%d turns, title: %q)\n\n", ToolResultBadge.Render("✔ [Session Loaded]:"), InfoValue.Render(targetID), sessionInfo.TurnCount, sessionInfo.Title)
			for _, m := range msgs {
				if m.Role == "user" {
					fmt.Printf("%s %s\n", AssistantBadge.Render("[User]:"), m.Content)
				} else {
					fmt.Printf("%s %s\n\n", AssistantBadge.Render("[Assistant]:"), m.Content)
				}
			}
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/model",
		Aliases:     []string{"/models"},
		Description: "View or switch active model",
		Usage:       "/model [model_name]",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			if len(args) >= 1 {
				targetModel := args[0]
				ctx.Config.ModelName = targetModel
				newAgent, newModel, aErr := fagent.BuildAgentWithModel(ctx.Ctx, ctx.Config)
				if aErr != nil {
					fmt.Printf("\033[31m[Error]: Failed to switch model to %q: %v\033[0m\n\n", targetModel, aErr)
					return CommandResult{}, aErr
				}
				*ctx.AppAgent = newAgent
				if ctx.ActiveModel != nil {
					*ctx.ActiveModel = newModel
				}
				if ctx.Runner != nil && *ctx.Runner != nil {
					newRunner, _ := runner.New(runner.Config{
						AppName:           "flashwhip",
						Agent:             *ctx.AppAgent,
						SessionService:    ctx.SessionSvc,
						AutoCreateSession: true,
					})
					*ctx.Runner = newRunner
				}
				fmt.Printf("\n%s Switched active model to %s\n\n", ToolResultBadge.Render("✔ [Model Switched]:"), InfoValue.Render(ctx.Config.ModelName))
				return CommandResult{}, nil
			}

			fmt.Println()
			fmt.Printf("%s Querying endpoint %s...\n", AssistantBadge.Render("[Flashwhip]"), InfoValue.Render(ctx.Config.BaseURL))
			modelsList, mErr := ollama.FetchAvailableModels(ctx.Config.BaseURL, ctx.Config.APIKey)
			if mErr != nil {
				fmt.Printf("\033[31m[Error]: Failed to fetch models: %v\033[0m\n\n", mErr)
			} else {
				fmt.Println(ThinkingBadge.Render("Available Endpoint Models:"))
				for _, mName := range modelsList {
					mCtxLen, _ := ollama.FetchModelContextLength(ctx.Config.BaseURL, ctx.Config.APIKey, mName)
					ctxStr := fmt.Sprintf("(max context: %d tokens)", mCtxLen)
					if mName == ctx.Config.ModelName {
						fmt.Printf("  • %-25s %s %s\n", InfoValue.Render(mName), ToolResultBadge.Render("(active)"), ctxStr)
					} else {
						fmt.Printf("  • %-25s %s\n", mName, ctxStr)
					}
				}
				fmt.Println("\n  Tip: Switch active model using '/model <model_name>'")
				fmt.Println()
			}
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/tokens",
		Aliases:     []string{"/usage", "/context"},
		Description: "View active session token usage and model max context window",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			sID := ""
			if ctx.SessionID != nil {
				sID = *ctx.SessionID
			}

			ctxLen, _ := ollama.FetchModelContextLength(ctx.Config.BaseURL, ctx.Config.APIKey, ctx.Config.ModelName)
			if ctxLen <= 0 {
				ctxLen = 32768
			}

			tokens := 0
			if ctx.ActiveModel != nil && *ctx.ActiveModel != nil {
				_, _, lastTotal := (*ctx.ActiveModel).Usage().LastContextTokens()
				if lastTotal > 0 {
					tokens = lastTotal
				}
			}

			turnCount := 0
			totalChars := 0
			database, err := db.DefaultDB()
			if err == nil && sID != "" {
				contents, cErr := database.GetSessionGenAIContents(sID)
				if cErr == nil {
					turnCount = len(contents)
					for _, c := range contents {
						for _, p := range c.Parts {
							totalChars += len(p.Text)
						}
					}
				}
			}

			if tokens == 0 {
				tokens = 1800 + (totalChars / 4)
			}

			pct := float64(tokens) / float64(ctxLen) * 100
			if pct > 100 {
				pct = 100
			}

			barWidth := 20
			filled := int((pct / 100.0) * float64(barWidth))
			if filled > barWidth {
				filled = barWidth
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

			var statusNote string
			if pct >= 90.0 {
				statusNote = " 🚨 [Context Almost Full! Run /compact]"
			} else if pct >= 75.0 {
				statusNote = " ⚠️ [Context Filling Up]"
			}

			fmt.Println()
			fmt.Println(ThinkingBadge.Render("⚡ FLASHWHIP TOKEN & CONTEXT USAGE:"))
			fmt.Printf("  • Active Model:       %s\n", InfoValue.Render(ctx.Config.ModelName))
			fmt.Printf("  • Max Context Window: %s tokens\n", InfoValue.Render(fmt.Sprintf("%d", ctxLen)))
			fmt.Printf("  • Conversation Turns: %s\n", InfoValue.Render(fmt.Sprintf("%d", turnCount)))
			fmt.Printf("  • Context Tokens:     %s / %s tokens\n", InfoValue.Render(fmt.Sprintf("%d", tokens)), InfoValue.Render(fmt.Sprintf("%d", ctxLen)))
			fmt.Printf("  • Context Usage:      [%s] %.2f%%%s\n\n", bar, pct, statusNote)

			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/tools",
		Description: "List all active coding harness tools",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			fmt.Println()
			toolInfos, tErr := tools.GetToolDescriptions()
			if tErr != nil {
				fmt.Printf("\033[31m[Error]: Failed to inspect tools: %v\033[0m\n\n", tErr)
			} else {
				fmt.Println(ThinkingBadge.Render("Active Code Harness Tools:"))
				for _, ti := range toolInfos {
					fmt.Printf("  • %-20s %s\n", ToolCallBadge.Render(ti.Name), ti.Description)
				}
				fmt.Println()
			}
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/compact",
		Description: "Force compact conversation context history",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			database, err := db.DefaultDB()
			if err != nil {
				fmt.Printf("\033[31m[Error]: Failed to open database: %v\033[0m\n", err)
				return CommandResult{}, err
			}
			sID := ""
			if ctx.SessionID != nil {
				sID = *ctx.SessionID
			}
			contents, cErr := database.GetSessionGenAIContents(sID)
			if cErr != nil || len(contents) == 0 {
				fmt.Println(ThinkingBadge.Render("No active session history to compact."))
				return CommandResult{}, nil
			}

			pruned := middleware.PruneContents(contents, 1)
			_ = database.ReplaceSessionGenAIContents(sID, pruned)

			if ctx.SessionSvc != nil {
				sessResp, crErr := ctx.SessionSvc.Create(ctx.Ctx, &session.CreateRequest{
					AppName:   "flashwhip",
					UserID:    "user",
					SessionID: sID,
				})
				if crErr == nil {
					for _, c := range pruned {
						if c == nil {
							continue
						}
						evt := &session.Event{
							Author: c.Role,
							LLMResponse: adkmodel.LLMResponse{
								Content: c,
							},
						}
						_ = ctx.SessionSvc.AppendEvent(ctx.Ctx, sessResp.Session, evt)
					}
				}
			}

			totalOrigChars := 0
			for _, c := range contents {
				for _, p := range c.Parts {
					totalOrigChars += len(p.Text)
				}
			}
			totalPrunedChars := 0
			for _, c := range pruned {
				for _, p := range c.Parts {
					totalPrunedChars += len(p.Text)
				}
			}

			fmt.Println()
			fmt.Printf("%s Context Compacting Completed:\n", ToolResultBadge.Render("✔ [Context Compacted]:"))
			fmt.Printf("  • Original history parts: %d (%d chars)\n", len(contents), totalOrigChars)
			fmt.Printf("  • Compacted history parts: %d (%d chars)\n\n", len(pruned), totalPrunedChars)
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/clear",
		Description: "Clear terminal screen",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			fmt.Print("\033[H\033[2J")
			fmt.Println(RenderBanner(ctx.Config.ModelName, ctx.Config.BaseURL, ctx.Config.ProjectRoot))
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/new",
		Description: "Create a new session (fresh chat)",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			newSessionID := fmt.Sprintf("chat-%d", time.Now().Unix())
			if ctx.SessionID != nil {
				*ctx.SessionID = newSessionID
			}
			fmt.Printf("✅ [New Session]: %s\n", newSessionID)
			return CommandResult{}, nil
		},
	})

	reg.Register(Command{
		Name:        "/reset",
		Description: "Reset to a fresh session (same as /new)",
		Handler: func(ctx *CommandContext, args []string) (CommandResult, error) {
			newSessionID := fmt.Sprintf("chat-%d", time.Now().Unix())
			if ctx.SessionID != nil {
				*ctx.SessionID = newSessionID
			}
			fmt.Printf("✅ [Reset Session]: %s\n", newSessionID)
			return CommandResult{}, nil
		},
	})

	return reg
}
