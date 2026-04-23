package wol

import "testing"

func TestParseMAC(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{name: "colon format", in: "aa:bb:cc:dd:ee:ff", wantErr: false},
		{name: "dash format", in: "aa-bb-cc-dd-ee-ff", wantErr: false},
		{name: "invalid", in: "bad-mac", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMAC(tt.in)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.in)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.in, err)
			}
		})
	}
}

func TestResolveBroadcastAddr(t *testing.T) {
	tests := []struct {
		name    string
		bcast   string
		iface   string
		want    string
		wantErr bool
	}{
		{
			name:    "explicit without port",
			bcast:   "192.168.1.255",
			iface:   "",
			want:    "192.168.1.255:9",
			wantErr: false,
		},
		{
			name:    "explicit with port",
			bcast:   "192.168.1.255:7",
			iface:   "",
			want:    "192.168.1.255:7",
			wantErr: false,
		},
		{
			name:    "default broadcast",
			bcast:   "",
			iface:   "",
			want:    "255.255.255.255:9",
			wantErr: false,
		},
		{
			name:    "unknown iface",
			bcast:   "",
			iface:   "__wol_test_no_such_iface__",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveBroadcastAddr(tt.bcast, tt.iface)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveBroadcastAddr(%q, %q) = %q, want %q", tt.bcast, tt.iface, got, tt.want)
			}
		})
	}
}

func TestSendMagicPacketWithInterface_InvalidMAC(t *testing.T) {
	_, err := SendMagicPacketWithInterface("bad-mac", "", "")
	if err == nil {
		t.Fatalf("expected error for invalid MAC")
	}
}

func TestResolveBroadcastAddrHostWithoutPort(t *testing.T) {
	got, err := resolveBroadcastAddr("not-an-ip", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "not-an-ip:9" {
		t.Fatalf("unexpected target: got %q, want %q", got, "not-an-ip:9")
	}
}

