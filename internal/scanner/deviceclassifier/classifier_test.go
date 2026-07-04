package deviceclassifier

import "testing"

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want string
	}{
		{
			name: "printer by 9100",
			in:   Input{Ports: []Port{{Port: 9100, State: "open"}}},
			want: CategoryPrinter,
		},
		{
			name: "camera by 554",
			in:   Input{Ports: []Port{{Port: 554, State: "open"}}},
			want: CategoryCamera,
		},
		{
			name: "router by ssh+http",
			in:   Input{Ports: []Port{{Port: 22, State: "open"}, {Port: 80, State: "open"}}},
			want: CategoryRouterSwitch,
		},
		{
			name: "desktop by rdp",
			in:   Input{Ports: []Port{{Port: 3389, State: "open"}}},
			want: CategoryDesktopLaptop,
		},
		{
			name: "unknown empty",
			in:   Input{},
			want: CategoryUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.in)
			if got != tt.want {
				t.Fatalf("Classify() = %q, want %q", got, tt.want)
			}
		})
	}
}
