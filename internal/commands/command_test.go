package commands

import (
	"context"
	"errors"
	"testing"
)

func nopHandler(ctx context.Context, req Request) (Response, error) {
	return NewOK("ok", nil), nil
}

func TestSourceValid(t *testing.T) {
	tests := []struct {
		source Source
		want   bool
	}{
		{SourceCLI, true},
		{SourceGUI, true},
		{SourceAPI, true},
		{SourceTest, true},
		{Source("telegram"), false},
		{Source(""), false},
	}
	for _, tt := range tests {
		if got := tt.source.Valid(); got != tt.want {
			t.Errorf("Source(%q).Valid() = %v, want %v", tt.source, got, tt.want)
		}
	}
}

func TestRiskNeedsConfirmation(t *testing.T) {
	if RiskRead.NeedsConfirmation() {
		t.Error("RiskRead should not require confirmation")
	}
	if RiskWrite.NeedsConfirmation() {
		t.Error("RiskWrite should not require confirmation")
	}
	if !RiskDestructive.NeedsConfirmation() {
		t.Error("RiskDestructive should require confirmation")
	}
}

func TestNewRequest(t *testing.T) {
	req := NewRequest("scan", SourceCLI)
	if req.Command != "scan" || req.Source != SourceCLI {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.Args == nil || req.Flags == nil {
		t.Fatal("maps should be initialized")
	}
}

func TestRequestArgHelpers(t *testing.T) {
	req := Request{
		Args:  map[string]string{"target": "192.168.1.0/24", "empty": "  "},
		Flags: map[string]bool{"verbose": true},
	}

	if got := req.Arg("target", "default"); got != "192.168.1.0/24" {
		t.Errorf("Arg(target) = %q", got)
	}
	if got := req.Arg("missing", "default"); got != "default" {
		t.Errorf("Arg(missing) = %q, want default", got)
	}
	if req.Flag("verbose") != true {
		t.Error("Flag(verbose) should be true")
	}
	if req.Flag("unknown") != false {
		t.Error("Flag(unknown) should be false")
	}

	value, err := req.RequireArg("target")
	if err != nil || value != "192.168.1.0/24" {
		t.Errorf("RequireArg(target) = %q, %v", value, err)
	}

	// Пустой аргумент считается отсутствующим.
	if _, err := req.RequireArg("empty"); !errors.Is(err, ErrInvalidCommand) {
		t.Errorf("RequireArg(empty) error = %v, want ErrInvalidCommand", err)
	}
	if _, err := req.RequireArg("missing"); !errors.Is(err, ErrInvalidCommand) {
		t.Errorf("RequireArg(missing) error = %v, want ErrInvalidCommand", err)
	}
}

func TestRequestCloneIsDeep(t *testing.T) {
	req := Request{
		Command: "scan",
		Args:    map[string]string{"target": "a"},
		Flags:   map[string]bool{"force": true},
	}

	clone := req.Clone()
	clone.Args["target"] = "b"
	clone.Flags["force"] = false
	clone.Command = "changed"

	if req.Args["target"] != "a" {
		t.Error("Clone shared Args map with original")
	}
	if req.Flags["force"] != true {
		t.Error("Clone shared Flags map with original")
	}
	if req.Command != "scan" {
		t.Error("Clone should not affect scalar fields of original")
	}
}

func TestRequestCloneNilMaps(t *testing.T) {
	clone := Request{Command: "x"}.Clone()
	if clone.Args == nil || clone.Flags == nil {
		t.Fatal("Clone should always return initialized maps")
	}
}

func TestResponseConstructors(t *testing.T) {
	ok := NewOK("готово", []string{"a"})
	if !ok.OK || ok.Message != "готово" {
		t.Fatalf("NewOK = %+v", ok)
	}

	fail := NewFail("ошибка")
	if fail.OK || fail.Message != "ошибка" {
		t.Fatalf("NewFail = %+v", fail)
	}
}

func TestResponseWithWarning(t *testing.T) {
	base := NewOK("m", nil)
	withWarning := base.WithWarning("осторожно")

	if len(base.Warnings) != 0 {
		t.Error("WithWarning must not mutate receiver")
	}
	if len(withWarning.Warnings) != 1 || withWarning.Warnings[0] != "осторожно" {
		t.Errorf("warnings = %v", withWarning.Warnings)
	}

	second := withWarning.WithWarning("ещё")
	if len(second.Warnings) != 2 || len(withWarning.Warnings) != 1 {
		t.Errorf("chained warnings = %v / %v", withWarning.Warnings, second.Warnings)
	}
}

func TestResponseWithMeta(t *testing.T) {
	base := NewOK("m", nil)
	withMeta := base.WithMeta("duration", "1s")

	if len(base.Meta) != 0 {
		t.Error("WithMeta must not mutate receiver")
	}
	if withMeta.Meta["duration"] != "1s" {
		t.Errorf("meta = %v", withMeta.Meta)
	}

	second := withMeta.WithMeta("source", "cli")
	if second.Meta["duration"] != "1s" || second.Meta["source"] != "cli" {
		t.Errorf("chained meta = %v", second.Meta)
	}
	if len(withMeta.Meta) != 1 {
		t.Error("second WithMeta mutated the first response")
	}
}

func TestCommandValidate(t *testing.T) {
	handler := nopHandler

	tests := []struct {
		name    string
		command Command
		wantErr bool
	}{
		{"valid", Command{Name: "scan", Risk: RiskRead, Handler: handler}, false},
		{"no name", Command{Name: "  ", Risk: RiskRead, Handler: handler}, true},
		{"no handler", Command{Name: "scan", Risk: RiskRead}, true},
		{"unknown risk", Command{Name: "scan", Risk: Risk("boom"), Handler: handler}, true},
		{"empty risk defaults invalid", Command{Name: "scan", Handler: handler}, true},
		{"empty alias", Command{Name: "scan", Aliases: []string{""}, Risk: RiskRead, Handler: handler}, true},
		{"duplicate alias", Command{Name: "scan", Aliases: []string{"s", "S"}, Risk: RiskRead, Handler: handler}, true},
		{"alias equals name", Command{Name: "scan", Aliases: []string{"SCAN"}, Risk: RiskRead, Handler: handler}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.command.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidCommand) {
				t.Errorf("error should wrap ErrInvalidCommand, got %v", err)
			}
		})
	}
}

func TestCommandMatches(t *testing.T) {
	cmd := Command{Name: "Scan", Aliases: []string{"s", "NetScan"}, Risk: RiskRead}

	for _, name := range []string{"scan", "SCAN", " Scan ", "s", "netscan"} {
		if !cmd.Matches(name) {
			t.Errorf("Matches(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"", "topology", "scan2"} {
		if cmd.Matches(name) {
			t.Errorf("Matches(%q) = true, want false", name)
		}
	}
}
