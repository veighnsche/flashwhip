package middleware

import (
	"testing"
)

func TestCanonicalizeArgs_KeySorting(t *testing.T) {
	map1 := map[string]any{"b": 2, "a": 1}
	map2 := map[string]any{"a": 1, "b": 2}

	sig1 := CanonicalSignature("read_file", map1)
	sig2 := CanonicalSignature("read_file", map2)

	if sig1 != sig2 {
		t.Errorf("CanonicalSignature expected identical signatures for map key order variants, got %q vs %q", sig1, sig2)
	}
}

func TestStallDetector_ConsecutiveLoop(t *testing.T) {
	sd := NewStallDetector(3)

	args := map[string]any{"file_path": "main.go"}

	// First call
	isLoop, _ := sd.RecordCall("read_file", args)
	if isLoop {
		t.Errorf("expected isLoop=false on call 1")
	}

	// Second call
	isLoop, _ = sd.RecordCall("read_file", args)
	if isLoop {
		t.Errorf("expected isLoop=false on call 2")
	}

	// Third call -> consecutive loop detected!
	isLoop, info := sd.RecordCall("read_file", args)
	if !isLoop {
		t.Errorf("expected isLoop=true on call 3")
	}
	if info.Pattern != PatternConsecutive {
		t.Errorf("Pattern = %s, want %s", info.Pattern, PatternConsecutive)
	}
	if info.ToolName != "read_file" {
		t.Errorf("ToolName = %s, want 'read_file'", info.ToolName)
	}
}

func TestStallDetector_OscillatingLoop(t *testing.T) {
	sd := NewStallDetector(3)

	argsA := map[string]any{"cmd": "git diff"}
	argsB := map[string]any{"file_path": "main.go"}

	// Sequence: A, B, A, B, A, B
	sd.RecordCall("exec_command", argsA) // 1
	sd.RecordCall("read_file", argsB)    // 2
	sd.RecordCall("exec_command", argsA) // 3
	sd.RecordCall("read_file", argsB)    // 4
	sd.RecordCall("exec_command", argsA) // 5

	isLoop, info := sd.RecordCall("read_file", argsB) // 6
	if !isLoop {
		t.Errorf("expected isLoop=true on 6th alternating call")
	}
	if info.Pattern != PatternOscillating {
		t.Errorf("Pattern = %s, want %s", info.Pattern, PatternOscillating)
	}
}

func TestStallDetector_NonLooping(t *testing.T) {
	sd := NewStallDetector(3)

	sd.RecordCall("read_file", map[string]any{"file_path": "a.go"})
	sd.RecordCall("read_file", map[string]any{"file_path": "b.go"})
	isLoop, _ := sd.RecordCall("read_file", map[string]any{"file_path": "c.go"})

	if isLoop {
		t.Errorf("expected isLoop=false for distinct file calls")
	}
}
