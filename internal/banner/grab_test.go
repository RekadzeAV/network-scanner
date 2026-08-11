package banner

import "testing"

func TestExtractVersionHint(t *testing.T) {
	tests := []struct {
		name   string
		port   int
		banner string
		want   string
	}{
		{
			name:   "ssh banner",
			port:   22,
			banner: "SSH-2.0-OpenSSH_9.3",
			want:   "SSH-2.0-OpenSSH_9.3",
		},
		{
			name:   "http status and server",
			port:   80,
			banner: "HTTP/1.1 200 OK | Server=nginx/1.25.0",
			want:   "HTTP/1.1 200 OK (nginx/1.25.0)",
		},
		{
			name:   "ftp trims response code",
			port:   21,
			banner: "FTP 220 FileZilla Server 1.8.0",
			want:   "FileZilla Server 1.8.0",
		},
		{
			name:   "smtp trims code",
			port:   25,
			banner: "SMTP 220 smtp.example.com ESMTP Postfix",
			want:   "smtp.example.com ESMTP Postfix",
		},
		{
			name:   "pop3 trims ok",
			port:   110,
			banner: "POP3 +OK Dovecot ready.",
			want:   "Dovecot ready.",
		},
		{
			name:   "unknown banner fallback",
			port:   9999,
			banner: "custom-app v1.2.3",
			want:   "custom-app v1.2.3",
		},
		{
			name:   "no response marker returns empty",
			port:   80,
			banner: "нет ответа",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractVersionHint(tt.port, tt.banner)
			if got != tt.want {
				t.Fatalf("ExtractVersionHint(%d, %q) = %q, want %q", tt.port, tt.banner, got, tt.want)
			}
		})
	}
}

func TestExtractVersionHintEmptyBanner(t *testing.T) {
	got := ExtractVersionHint(80, "   ")
	if got != "" {
		t.Fatalf("expected empty hint for empty banner, got %q", got)
	}
}

func TestExtractVersionHint_HTTPS(t *testing.T) {
	got := ExtractVersionHint(443, "HTTP/1.1 200 OK | Server=Apache")
	if got != "HTTP/1.1 200 OK (Apache)" {
		t.Fatalf("expected HTTPS hint, got %q", got)
	}
}

func TestExtractVersionHint_8080(t *testing.T) {
	got := ExtractVersionHint(8080, "HTTP/1.1 200 OK | Server=tomcat")
	if got != "HTTP/1.1 200 OK (tomcat)" {
		t.Fatalf("expected 8080 hint, got %q", got)
	}
}

func TestExtractVersionHint_8443(t *testing.T) {
	got := ExtractVersionHint(8443, "HTTP/1.1 200 OK | Server=nginx")
	if got != "HTTP/1.1 200 OK (nginx)" {
		t.Fatalf("expected 8443 hint, got %q", got)
	}
}

func TestExtractVersionHint_OnlyStatus(t *testing.T) {
	got := ExtractVersionHint(80, "HTTP/1.1 200 OK")
	if got != "HTTP/1.1 200 OK" {
		t.Fatalf("expected status only, got %q", got)
	}
}

func TestExtractVersionHint_OnlyServer(t *testing.T) {
	got := ExtractVersionHint(80, "Server=nginx")
	if got != "nginx" {
		t.Fatalf("expected server only, got %q", got)
	}
}
