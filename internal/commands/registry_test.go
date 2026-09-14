package commands

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRegistryRegister(t *testing.T) {
	reg := NewRegistry()

	if err := reg.Register(
		Command{Name: "scan", Aliases: []string{"s"}, Risk: RiskRead, Handler: nopHandler},
		Command{Name: "report", Risk: RiskWrite, Handler: nopHandler},
	); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if reg.Len() != 2 {
		t.Fatalf("Len = %d, want 2", reg.Len())
	}
	if !reg.Has("scan") || !reg.Has("S") || !reg.Has("report") {
		t.Error("Has should find commands by name and alias")
	}
}

func TestRegistryRegisterErrors(t *testing.T) {
	tests := []struct {
		name    string
		first   []Command
		second  []Command
		wantErr error
	}{
		{
			name:    "invalid command",
			second:  []Command{{Name: "", Handler: nopHandler}},
			wantErr: ErrInvalidCommand,
		},
		{
			name:    "duplicate name",
			first:   []Command{{Name: "scan", Risk: RiskRead, Handler: nopHandler}},
			second:  []Command{{Name: "SCAN", Risk: RiskRead, Handler: nopHandler}},
			wantErr: ErrDuplicateCommand,
		},
		{
			name:    "duplicate alias",
			first:   []Command{{Name: "scan", Aliases: []string{"s"}, Risk: RiskRead, Handler: nopHandler}},
			second:  []Command{{Name: "search", Aliases: []string{"S"}, Risk: RiskRead, Handler: nopHandler}},
			wantErr: ErrDuplicateCommand,
		},
		{
			name: "duplicate within batch",
			second: []Command{
				{Name: "a", Risk: RiskRead, Handler: nopHandler},
				{Name: "A", Risk: RiskRead, Handler: nopHandler},
			},
			wantErr: ErrDuplicateCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry()
			if len(tt.first) > 0 {
				if err := reg.Register(tt.first...); err != nil {
					t.Fatalf("seed Register: %v", err)
				}
			}

			err := reg.Register(tt.second...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Register error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryRegisterIsAtomic(t *testing.T) {
	reg := NewRegistry()

	err := reg.Register(
		Command{Name: "good", Risk: RiskRead, Handler: nopHandler},
		Command{Name: "bad", Risk: Risk("nope"), Handler: nopHandler},
	)
	if err == nil {
		t.Fatal("expected error for invalid command in batch")
	}
	if reg.Len() != 0 {
		t.Errorf("failed batch must not be partially applied, Len = %d", reg.Len())
	}
	if reg.Has("good") {
		t.Error("failed batch must not register valid commands")
	}
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(Command{Name: "Scan", Aliases: []string{"s"}, Risk: RiskRead, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	for _, name := range []string{"scan", "SCAN", " s "} {
		cmd, ok := reg.Get(name)
		if !ok {
			t.Fatalf("Get(%q) not found", name)
		}
		if cmd.Name != "Scan" {
			t.Errorf("Get(%q).Name = %q, want Scan", name, cmd.Name)
		}
	}

	if _, ok := reg.Get("nope"); ok {
		t.Error("Get(nope) should not be found")
	}
}

func TestRegistryListAndNames(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(
		Command{Name: "zeta", Risk: RiskRead, Handler: nopHandler},
		Command{Name: "alpha", Risk: RiskRead, Handler: nopHandler},
		Command{Name: "mid", Risk: RiskRead, Handler: nopHandler},
	); err != nil {
		t.Fatalf("Register: %v", err)
	}

	names := reg.Names()
	want := []string{"alpha", "mid", "zeta"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("Names = %v, want %v", names, want)
	}

	list := reg.List()
	if len(list) != 3 || list[0].Name != "alpha" || list[2].Name != "zeta" {
		t.Errorf("List not sorted: %v", list)
	}
}

func TestRegistryDispatch(t *testing.T) {
	reg := NewRegistry()
	err := reg.Register(Command{
		Name: "scan",
		Risk: RiskRead,
		Handler: func(ctx context.Context, req Request) (Response, error) {
			return NewOK("scanned "+req.Arg("target", "?"), req.Source), nil
		},
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	req := NewRequest("scan", SourceCLI)
	req.Args["target"] = "10.0.0.0/24"

	resp, err := reg.Dispatch(context.Background(), req)
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !resp.OK || resp.Message != "scanned 10.0.0.0/24" {
		t.Errorf("resp = %+v", resp)
	}
}

func TestRegistryDispatchNamed(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(Command{Name: "ping", Risk: RiskRead, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	resp, err := reg.DispatchNamed(context.Background(), "ping", SourceGUI)
	if err != nil || !resp.OK {
		t.Fatalf("DispatchNamed = %+v, %v", resp, err)
	}
}

func TestRegistryDispatchNotFound(t *testing.T) {
	reg := NewRegistry()
	_, err := reg.Dispatch(context.Background(), NewRequest("ghost", SourceAPI))
	if !errors.Is(err, ErrCommandNotFound) {
		t.Errorf("error = %v, want ErrCommandNotFound", err)
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error should mention command name: %v", err)
	}
}

func TestRegistryDispatchRequiresConfirmation(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(Command{Name: "factory-reset", Risk: RiskDestructive, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := reg.Dispatch(context.Background(), NewRequest("factory-reset", SourceAPI)); !errors.Is(err, ErrConfirmationRequired) {
		t.Errorf("unconfirmed dispatch error = %v, want ErrConfirmationRequired", err)
	}

	req := NewRequest("factory-reset", SourceCLI)
	req.Confirmed = true
	if resp, err := reg.Dispatch(context.Background(), req); err != nil || !resp.OK {
		t.Errorf("confirmed dispatch = %+v, %v", resp, err)
	}
}

func TestRegistryDispatchNilContext(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(Command{Name: "x", Risk: RiskRead, Handler: nopHandler}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	//lint:ignore SA1012 проверяем защиту от nil context в публичном API
	resp, err := reg.Dispatch(nil, NewRequest("x", SourceTest)) //nolint:staticcheck
	if err != nil || !resp.OK {
		t.Errorf("nil context dispatch = %+v, %v", resp, err)
	}
}

func TestRegistryDispatchClonesRequest(t *testing.T) {
	reg := NewRegistry()
	if err := reg.Register(Command{
		Name: "mutate",
		Risk: RiskRead,
		Handler: func(ctx context.Context, req Request) (Response, error) {
			req.Args["target"] = "changed"
			return NewOK("", nil), nil
		},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	req := NewRequest("mutate", SourceCLI)
	req.Args["target"] = "original"
	if _, err := reg.Dispatch(context.Background(), req); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if req.Args["target"] != "original" {
		t.Errorf("handler mutated caller request: %q", req.Args["target"])
	}
}

func TestRegistryMiddlewareOrder(t *testing.T) {
	var order []string
	mw := func(tag string) Middleware {
		return func(next Handler) Handler {
			return func(ctx context.Context, req Request) (Response, error) {
				order = append(order, "enter:"+tag)
				resp, err := next(ctx, req)
				order = append(order, "exit:"+tag)
				return resp, err
			}
		}
	}

	reg := NewRegistry(mw("outer"))
	reg.Use(mw("inner"))

	if err := reg.Register(Command{Name: "x", Risk: RiskRead, Handler: func(ctx context.Context, req Request) (Response, error) {
		order = append(order, "handler")
		return NewOK("", nil), nil
	}}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if _, err := reg.Dispatch(context.Background(), NewRequest("x", SourceTest)); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	want := "enter:outer,enter:inner,handler,exit:inner,exit:outer"
	if got := strings.Join(order, ","); got != want {
		t.Errorf("order = %s, want %s", got, want)
	}
}

func TestRegistryConcurrentAccess(t *testing.T) {
	reg := NewRegistry()
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			_ = reg.Register(Command{Name: strings.ToLower(string(rune('a'+i%26))) + string(rune('0'+i%10)), Risk: RiskRead, Handler: nopHandler})
			reg.List()
			reg.Names()
		}
	}()

	for i := 0; i < 50; i++ {
		_, _ = reg.Dispatch(context.Background(), NewRequest("unknown", SourceTest))
		reg.Has("a0")
	}
	<-done
}
