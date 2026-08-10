package errors

// Code represents a stable numeric error code.
type Code int

const (
	// Unknown / Unclassified (0)
	ErrCodeUnknown Code = 0

	// System & Config Errors (1000-1999)
	ErrCodeConfigInvalid          Code = 1001
	ErrCodeDirChangeFailed        Code = 1002
	ErrCodeContextBuildFailed     Code = 1003
	ErrCodeConfigHomeDirFailed    Code = 1004
	ErrCodeConfigInvalidModel     Code = 1005
	ErrCodeConfigInvalidURL       Code = 1006
	ErrCodeConfigEnvironmentError Code = 1007

	// Database & Persistence Errors (2000-2999)
	ErrCodeDBOpenFailed      Code = 2001
	ErrCodeDBMigrationFailed Code = 2002
	ErrCodeDBSaveFailed      Code = 2003
	ErrCodeDBSessionNotFound Code = 2004
	ErrCodeDBQueryFailed     Code = 2005
	ErrCodeDBCloseFailed     Code = 2006
	ErrCodeDBExecFailed      Code = 2007
	ErrCodeDBCorrupt         Code = 2008
	ErrCodeDBMarshalFailed   Code = 2009
	ErrCodeDBUnmarshalFailed Code = 2010

	// Provider, Network & LLM Errors (3000-3999)
	ErrCodeProviderInitFailed           Code = 3001
	ErrCodeModelFetchFailed             Code = 3002
	ErrCodeStreamReadFailed             Code = 3003
	ErrCodeModelContextFailed           Code = 3004
	ErrCodeProviderRequestFailed        Code = 3005
	ErrCodeNetHTTPClientFailed          Code = 3006
	ErrCodeNetTimeout                   Code = 3007
	ErrCodeNetDNSFailed                 Code = 3008
	ErrCodeNetConnectionRefused         Code = 3009
	ErrCodeProviderEmptyModel           Code = 3101
	ErrCodeProviderNoModelsFound        Code = 3102
	ErrCodeProviderMessageBuildFailed   Code = 3103
	ErrCodeProviderMarshalFailed        Code = 3104
	ErrCodeProviderHTTPStatus           Code = 3105
	ErrCodeProviderResponseDecodeFailed Code = 3106
	ErrCodeProviderEmptyResponse        Code = 3107
	ErrCodeProviderContextExceeded      Code = 3108

	// Agent & Core Runner Errors (4000-4999)
	ErrCodeAgentBuildFailed      Code = 4001
	ErrCodeMaxTurnsReached       Code = 4002
	ErrCodeRunnerExecutionFailed Code = 4003
	ErrCodeAgentRunnerInitFailed Code = 4004
	ErrCodeAgentStreamCancelled  Code = 4005
	ErrCodeAgentMiddlewareFailed Code = 4006
	ErrCodeLoopDetected          Code = 4007

	// Tool Execution Errors (5000-5999)
	ErrCodeToolExecFailed       Code = 5001
	ErrCodeToolFileNotFound     Code = 5002
	ErrCodeToolAmbiguousMatch    Code = 5003
	ErrCodeToolPermissionDenied  Code = 5004
	ErrCodeToolNetworkError     Code = 5005
	ErrCodeToolInvalidArgs      Code = 5006
	ErrCodeToolInitFailed       Code = 5007
	ErrCodeToolTimeout          Code = 5008
	ErrCodeToolHTTPStatusError  Code = 5009
	ErrCodeToolTargetNotFound   Code = 5010
	ErrCodeToolCommandEmpty     Code = 5011
	ErrCodeToolPathEmpty        Code = 5012

	// UI & REPL Errors (6000-6999)
	ErrCodeSessionLoadFailed   Code = 6001
	ErrCodeHistorySaveFailed   Code = 6002
	ErrCodeREPLExecutionFailed Code = 6003
	ErrCodeUICommandUnknown    Code = 6004
	ErrCodeUIRendererFailed    Code = 6005
	ErrCodeUIReadlineFailed    Code = 6006
	ErrCodeUISinglePromptFailed Code = 6007
)

// ErrMaxTurnsReached is returned when the agent exceeds the configured turn limit.
var ErrMaxTurnsReached = New(ErrCodeMaxTurnsReached, "max turns reached")

// ErrLoopDetected is returned when repetitive tool call loops or stalled execution are detected.
var ErrLoopDetected = New(ErrCodeLoopDetected, "repetitive tool call loop detected")
