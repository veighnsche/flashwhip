package errors

import "sort"

// ErrorSpec defines diagnostic metadata for a registered stable error code.
type ErrorSpec struct {
	Code        Code   `json:"code"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Remedy      string `json:"remedy"`
}

var registry = map[Code]ErrorSpec{
	// Configuration (1000-1999)
	ErrCodeConfigInvalid: {
		Code:        ErrCodeConfigInvalid,
		Name:        "ErrCodeConfigInvalid",
		Category:    "Configuration",
		Description: "The specified configuration parameter or path flag is invalid.",
		Remedy:      "Check command line flags like --cwd, --url, or environment variables for proper values.",
	},
	ErrCodeDirChangeFailed: {
		Code:        ErrCodeDirChangeFailed,
		Name:        "ErrCodeDirChangeFailed",
		Category:    "Configuration",
		Description: "Failed to change the current working directory to the requested project path.",
		Remedy:      "Verify that the directory exists and that Flashwhip has permission to access it.",
	},
	ErrCodeContextBuildFailed: {
		Code:        ErrCodeContextBuildFailed,
		Name:        "ErrCodeContextBuildFailed",
		Category:    "Configuration",
		Description: "Failed to compile the agent context or system prompt instructions.",
		Remedy:      "Ensure system rules and files in the project workspace are readable.",
	},
	ErrCodeConfigHomeDirFailed: {
		Code:        ErrCodeConfigHomeDirFailed,
		Name:        "ErrCodeConfigHomeDirFailed",
		Category:    "Configuration",
		Description: "Failed to resolve the current user's home directory.",
		Remedy:      "Ensure the HOME environment variable is properly defined.",
	},
	ErrCodeConfigInvalidModel: {
		Code:        ErrCodeConfigInvalidModel,
		Name:        "ErrCodeConfigInvalidModel",
		Category:    "Configuration",
		Description: "The requested model identifier is empty or invalid.",
		Remedy:      "Specify a valid model using the --model flag or /model command.",
	},
	ErrCodeConfigInvalidURL: {
		Code:        ErrCodeConfigInvalidURL,
		Name:        "ErrCodeConfigInvalidURL",
		Category:    "Configuration",
		Description: "The target endpoint URL format is invalid.",
		Remedy:      "Provide a valid HTTP/HTTPS URL for the provider endpoint.",
	},
	ErrCodeConfigEnvironmentError: {
		Code:        ErrCodeConfigEnvironmentError,
		Name:        "ErrCodeConfigEnvironmentError",
		Category:    "Configuration",
		Description: "Failed to load environment variables or configuration file.",
		Remedy:      "Inspect environment settings and permissions.",
	},

	// Database (2000-2999)
	ErrCodeDBOpenFailed: {
		Code:        ErrCodeDBOpenFailed,
		Name:        "ErrCodeDBOpenFailed",
		Category:    "Database",
		Description: "Failed to initialize or open the SQLite session database.",
		Remedy:      "Check that ~/.flashwhip directory is writable and disk space is available.",
	},
	ErrCodeDBMigrationFailed: {
		Code:        ErrCodeDBMigrationFailed,
		Name:        "ErrCodeDBMigrationFailed",
		Category:    "Database",
		Description: "Failed to execute database schema migrations.",
		Remedy:      "Inspect database integrity or delete ~/.flashwhip/flashwhip.db if corrupt.",
	},
	ErrCodeDBSaveFailed: {
		Code:        ErrCodeDBSaveFailed,
		Name:        "ErrCodeDBSaveFailed",
		Category:    "Database",
		Description: "Failed to persist chat message or GenAI content tree to SQLite storage.",
		Remedy:      "Ensure SQLite database file permissions are intact.",
	},
	ErrCodeDBSessionNotFound: {
		Code:        ErrCodeDBSessionNotFound,
		Name:        "ErrCodeDBSessionNotFound",
		Category:    "Database",
		Description: "The requested session ID could not be found in the database.",
		Remedy:      "Use 'flashwhip sessions' to view all available session IDs.",
	},
	ErrCodeDBQueryFailed: {
		Code:        ErrCodeDBQueryFailed,
		Name:        "ErrCodeDBQueryFailed",
		Category:    "Database",
		Description: "Failed to execute a query against SQLite storage.",
		Remedy:      "Verify session parameters and database health.",
	},
	ErrCodeDBCloseFailed: {
		Code:        ErrCodeDBCloseFailed,
		Name:        "ErrCodeDBCloseFailed",
		Category:    "Database",
		Description: "Failed to cleanly close the database connection.",
		Remedy:      "Ensure no background operations lock the database on exit.",
	},
	ErrCodeDBExecFailed: {
		Code:        ErrCodeDBExecFailed,
		Name:        "ErrCodeDBExecFailed",
		Category:    "Database",
		Description: "Failed to execute database SQL statement.",
		Remedy:      "Check SQL query parameters and database file integrity.",
	},
	ErrCodeDBCorrupt: {
		Code:        ErrCodeDBCorrupt,
		Name:        "ErrCodeDBCorrupt",
		Category:    "Database",
		Description: "Database file or payload structure is corrupted.",
		Remedy:      "Backup and delete ~/.flashwhip/flashwhip.db to recreate a fresh database.",
	},
	ErrCodeDBMarshalFailed: {
		Code:        ErrCodeDBMarshalFailed,
		Name:        "ErrCodeDBMarshalFailed",
		Category:    "Database",
		Description: "Failed to serialize message payload into JSON for storage.",
		Remedy:      "Check message payload structures for un-marshalable fields.",
	},
	ErrCodeDBUnmarshalFailed: {
		Code:        ErrCodeDBUnmarshalFailed,
		Name:        "ErrCodeDBUnmarshalFailed",
		Category:    "Database",
		Description: "Failed to deserialize JSON message payload from storage.",
		Remedy:      "Inspect message table contents for malformed JSON strings.",
	},

	// Provider & Network (3000-3999)
	ErrCodeProviderInitFailed: {
		Code:        ErrCodeProviderInitFailed,
		Name:        "ErrCodeProviderInitFailed",
		Category:    "Provider",
		Description: "Failed to initialize Ollama / OpenAI model client.",
		Remedy:      "Check the model endpoint URL (--url) and model name (--model).",
	},
	ErrCodeModelFetchFailed: {
		Code:        ErrCodeModelFetchFailed,
		Name:        "ErrCodeModelFetchFailed",
		Category:    "Provider",
		Description: "Failed to fetch available model list from Ollama server.",
		Remedy:      "Ensure Ollama daemon is running (`ollama serve`) and reachable.",
	},
	ErrCodeStreamReadFailed: {
		Code:        ErrCodeStreamReadFailed,
		Name:        "ErrCodeStreamReadFailed",
		Category:    "Provider",
		Description: "Failed to read line or parse chunk from streaming response.",
		Remedy:      "Check network connection or model endpoint stability.",
	},
	ErrCodeModelContextFailed: {
		Code:        ErrCodeModelContextFailed,
		Name:        "ErrCodeModelContextFailed",
		Category:    "Provider",
		Description: "Failed to inspect model context length window.",
		Remedy:      "Verify model exists on the provider host via `ollama list`.",
	},
	ErrCodeProviderRequestFailed: {
		Code:        ErrCodeProviderRequestFailed,
		Name:        "ErrCodeProviderRequestFailed",
		Category:    "Provider",
		Description: "Provider returned an error response during generation.",
		Remedy:      "Inspect prompt payload and model availability.",
	},
	ErrCodeNetHTTPClientFailed: {
		Code:        ErrCodeNetHTTPClientFailed,
		Name:        "ErrCodeNetHTTPClientFailed",
		Category:    "Network",
		Description: "Failed to create HTTP client or configure TLS transport.",
		Remedy:      "Check system proxy settings and TLS configuration.",
	},
	ErrCodeNetTimeout: {
		Code:        ErrCodeNetTimeout,
		Name:        "ErrCodeNetTimeout",
		Category:    "Network",
		Description: "HTTP request timed out before receiving a response.",
		Remedy:      "Verify network connection speed or increase request timeout.",
	},
	ErrCodeNetDNSFailed: {
		Code:        ErrCodeNetDNSFailed,
		Name:        "ErrCodeNetDNSFailed",
		Category:    "Network",
		Description: "DNS resolution failed for the target host URL.",
		Remedy:      "Check domain name spelling and DNS resolution settings.",
	},
	ErrCodeNetConnectionRefused: {
		Code:        ErrCodeNetConnectionRefused,
		Name:        "ErrCodeNetConnectionRefused",
		Category:    "Network",
		Description: "Target endpoint refused connection on specified port.",
		Remedy:      "Verify target server is running (e.g. `ollama serve`).",
	},
	ErrCodeProviderEmptyModel: {
		Code:        ErrCodeProviderEmptyModel,
		Name:        "ErrCodeProviderEmptyModel",
		Category:    "Provider",
		Description: "Model name argument passed to provider was empty.",
		Remedy:      "Pass a non-empty model identifier name.",
	},
	ErrCodeProviderNoModelsFound: {
		Code:        ErrCodeProviderNoModelsFound,
		Name:        "ErrCodeProviderNoModelsFound",
		Category:    "Provider",
		Description: "Provider endpoint returned an empty model list.",
		Remedy:      "Pull a model using `ollama pull <model>` on the server host.",
	},
	ErrCodeProviderMessageBuildFailed: {
		Code:        ErrCodeProviderMessageBuildFailed,
		Name:        "ErrCodeProviderMessageBuildFailed",
		Category:    "Provider",
		Description: "Failed to convert LLMRequest messages into provider chat format.",
		Remedy:      "Verify prompt message role types and content structure.",
	},
	ErrCodeProviderMarshalFailed: {
		Code:        ErrCodeProviderMarshalFailed,
		Name:        "ErrCodeProviderMarshalFailed",
		Category:    "Provider",
		Description: "Failed to marshal LLM request body into JSON payload.",
		Remedy:      "Check request payload arguments and tool call schemas.",
	},
	ErrCodeProviderHTTPStatus: {
		Code:        ErrCodeProviderHTTPStatus,
		Name:        "ErrCodeProviderHTTPStatus",
		Category:    "Provider",
		Description: "LLM endpoint returned an HTTP error status code.",
		Remedy:      "Check server logs for error details or verify model availability.",
	},
	ErrCodeProviderResponseDecodeFailed: {
		Code:        ErrCodeProviderResponseDecodeFailed,
		Name:        "ErrCodeProviderResponseDecodeFailed",
		Category:    "Provider",
		Description: "Failed to decode response JSON returned by LLM endpoint.",
		Remedy:      "Verify provider API compatibility with OpenAI / Ollama standard.",
	},
	ErrCodeProviderEmptyResponse: {
		Code:        ErrCodeProviderEmptyResponse,
		Name:        "ErrCodeProviderEmptyResponse",
		Category:    "Provider",
		Description: "LLM endpoint returned a response with zero choices.",
		Remedy:      "Retry prompt or inspect server generation status.",
	},
	ErrCodeProviderContextExceeded: {
		Code:        ErrCodeProviderContextExceeded,
		Name:        "ErrCodeProviderContextExceeded",
		Category:    "Provider",
		Description: "Prompt token length exceeds the maximum token context window of the model.",
		Remedy:      "Run /compact or start a new session to reduce prompt token size.",
	},

	// Agent & Core Runner (4000-4999)
	ErrCodeAgentBuildFailed: {
		Code:        ErrCodeAgentBuildFailed,
		Name:        "ErrCodeAgentBuildFailed",
		Category:    "Agent",
		Description: "Failed to instantiate Google ADK agent runner instance.",
		Remedy:      "Check system config, model configuration, and tool registrations.",
	},
	ErrCodeMaxTurnsReached: {
		Code:        ErrCodeMaxTurnsReached,
		Name:        "ErrCodeMaxTurnsReached",
		Category:    "Agent",
		Description: "Agent turn count reached configured maximum limit.",
		Remedy:      "Use --max-turns (-t) to increase maximum allowed tool-call turns.",
	},
	ErrCodeRunnerExecutionFailed: {
		Code:        ErrCodeRunnerExecutionFailed,
		Name:        "ErrCodeRunnerExecutionFailed",
		Category:    "Agent",
		Description: "Error occurred during event stream loop execution.",
		Remedy:      "Check tool execution outputs and LLM endpoint stability.",
	},
	ErrCodeAgentRunnerInitFailed: {
		Code:        ErrCodeAgentRunnerInitFailed,
		Name:        "ErrCodeAgentRunnerInitFailed",
		Category:    "Agent",
		Description: "Failed to initialize runner instance for single prompt execution.",
		Remedy:      "Ensure Google ADK runner configuration is valid.",
	},
	ErrCodeAgentStreamCancelled: {
		Code:        ErrCodeAgentStreamCancelled,
		Name:        "ErrCodeAgentStreamCancelled",
		Category:    "Agent",
		Description: "Agent execution stream was interrupted by user or system signal.",
		Remedy:      "Re-run prompt or command.",
	},
	ErrCodeAgentMiddlewareFailed: {
		Code:        ErrCodeAgentMiddlewareFailed,
		Name:        "ErrCodeAgentMiddlewareFailed",
		Category:    "Agent",
		Description: "Failed executing agent middleware or content transformation.",
		Remedy:      "Inspect middleware pipeline configuration.",
	},
	ErrCodeLoopDetected: {
		Code:        ErrCodeLoopDetected,
		Name:        "ErrCodeLoopDetected",
		Category:    "Agent",
		Description: "Repetitive tool call loop or stalled execution pattern detected by Stall Guard.",
		Remedy:      "Modify prompt instructions, change parameters, or break execution loop.",
	},

	// Tools (5000-5999)
	ErrCodeToolExecFailed: {
		Code:        ErrCodeToolExecFailed,
		Name:        "ErrCodeToolExecFailed",
		Category:    "Tools",
		Description: "Command or subprocess tool execution failed.",
		Remedy:      "Check command syntax, arguments, binary presence, or execution environment.",
	},
	ErrCodeToolFileNotFound: {
		Code:        ErrCodeToolFileNotFound,
		Name:        "ErrCodeToolFileNotFound",
		Category:    "Tools",
		Description: "Target file or directory path requested by tool was not found.",
		Remedy:      "Verify file path relative to project root.",
	},
	ErrCodeToolAmbiguousMatch: {
		Code:        ErrCodeToolAmbiguousMatch,
		Name:        "ErrCodeToolAmbiguousMatch",
		Category:    "Tools",
		Description: "Search pattern matched multiple occurrences when a single target was required.",
		Remedy:      "Provide more specific search context in edit_file target_content.",
	},
	ErrCodeToolPermissionDenied: {
		Code:        ErrCodeToolPermissionDenied,
		Name:        "ErrCodeToolPermissionDenied",
		Category:    "Tools",
		Description: "Tool was denied permission to read/write target file or directory.",
		Remedy:      "Check file system permissions on target file.",
	},
	ErrCodeToolNetworkError: {
		Code:        ErrCodeToolNetworkError,
		Name:        "ErrCodeToolNetworkError",
		Category:    "Tools",
		Description: "Web fetch or search tool encountered a network error.",
		Remedy:      "Check internet connectivity and target URL availability.",
	},
	ErrCodeToolInvalidArgs: {
		Code:        ErrCodeToolInvalidArgs,
		Name:        "ErrCodeToolInvalidArgs",
		Category:    "Tools",
		Description: "Tool received invalid or missing arguments.",
		Remedy:      "Ensure required argument fields are provided in tool call.",
	},
	ErrCodeToolInitFailed: {
		Code:        ErrCodeToolInitFailed,
		Name:        "ErrCodeToolInitFailed",
		Category:    "Tools",
		Description: "Failed to construct tool definition or function schema.",
		Remedy:      "Verify tool registration code in pkg/tools.",
	},
	ErrCodeToolTimeout: {
		Code:        ErrCodeToolTimeout,
		Name:        "ErrCodeToolTimeout",
		Category:    "Tools",
		Description: "Command execution timed out before completion.",
		Remedy:      "Increase timeout_seconds argument or run shorter command.",
	},
	ErrCodeToolHTTPStatusError: {
		Code:        ErrCodeToolHTTPStatusError,
		Name:        "ErrCodeToolHTTPStatusError",
		Category:    "Tools",
		Description: "Web request returned an HTTP error status code (e.g. 403, 404, 500).",
		Remedy:      "Check target URL availability or use web search snippets.",
	},
	ErrCodeToolTargetNotFound: {
		Code:        ErrCodeToolTargetNotFound,
		Name:        "ErrCodeToolTargetNotFound",
		Category:    "Tools",
		Description: "Search or edit target string was not found in target file.",
		Remedy:      "Verify exact whitespace and target string content.",
	},
	ErrCodeToolCommandEmpty: {
		Code:        ErrCodeToolCommandEmpty,
		Name:        "ErrCodeToolCommandEmpty",
		Category:    "Tools",
		Description: "Command string provided to shell tool was empty.",
		Remedy:      "Pass a non-empty command string to exec_command.",
	},
	ErrCodeToolPathEmpty: {
		Code:        ErrCodeToolPathEmpty,
		Name:        "ErrCodeToolPathEmpty",
		Category:    "Tools",
		Description: "File or directory path argument was empty.",
		Remedy:      "Provide a valid file or directory path argument.",
	},

	// UI & REPL (6000-6999)
	ErrCodeSessionLoadFailed: {
		Code:        ErrCodeSessionLoadFailed,
		Name:        "ErrCodeSessionLoadFailed",
		Category:    "UI",
		Description: "Failed to load session history for attachment.",
		Remedy:      "Verify session ID exists in database.",
	},
	ErrCodeHistorySaveFailed: {
		Code:        ErrCodeHistorySaveFailed,
		Name:        "ErrCodeHistorySaveFailed",
		Category:    "UI",
		Description: "Failed to read or write terminal REPL prompt history file.",
		Remedy:      "Ensure ~/.flashwhip_history is writable.",
	},
	ErrCodeREPLExecutionFailed: {
		Code:        ErrCodeREPLExecutionFailed,
		Name:        "ErrCodeREPLExecutionFailed",
		Category:    "UI",
		Description: "Failed to process prompt or run loop in interactive REPL.",
		Remedy:      "Check error message details.",
	},
	ErrCodeUICommandUnknown: {
		Code:        ErrCodeUICommandUnknown,
		Name:        "ErrCodeUICommandUnknown",
		Category:    "UI",
		Description: "Entered REPL slash command is not registered.",
		Remedy:      "Type '/help' to list all supported REPL slash commands.",
	},
	ErrCodeUIRendererFailed: {
		Code:        ErrCodeUIRendererFailed,
		Name:        "ErrCodeUIRendererFailed",
		Category:    "UI",
		Description: "Failed to render formatted Markdown or terminal UI component.",
		Remedy:      "Ensure terminal supports ANSI color output.",
	},
	ErrCodeUIReadlineFailed: {
		Code:        ErrCodeUIReadlineFailed,
		Name:        "ErrCodeUIReadlineFailed",
		Category:    "UI",
		Description: "Readline terminal input handler initialization failed.",
		Remedy:      "Check terminal TTY device and permissions.",
	},
	ErrCodeUISinglePromptFailed: {
		Code:        ErrCodeUISinglePromptFailed,
		Name:        "ErrCodeUISinglePromptFailed",
		Category:    "UI",
		Description: "Single-shot prompt execution failed.",
		Remedy:      "Check model endpoint availability and prompt input.",
	},
}

// Lookup returns the ErrorSpec definition for a given Code if registered.
func Lookup(code Code) (ErrorSpec, bool) {
	spec, ok := registry[code]
	return spec, ok
}

// All returns a slice of all registered ErrorSpec entries sorted by error code.
func All() []ErrorSpec {
	specs := make([]ErrorSpec, 0, len(registry))
	for _, spec := range registry {
		specs = append(specs, spec)
	}
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Code < specs[j].Code
	})
	return specs
}
