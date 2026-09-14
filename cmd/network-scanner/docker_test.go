package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// M6.4: Тесты для Docker-конфигурации
// ============================================================================

// getProjectRoot возвращает корень проекта
func getProjectRoot(t *testing.T) string {
	t.Helper()
	// Ищем корень проекта по go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatal("go.mod not found")
	return ""
}

// TestDockerfile_Exists — ветка: наличие Dockerfile
func TestDockerfile_Exists(t *testing.T) {
	root := getProjectRoot(t)
	path := filepath.Join(root, "Dockerfile")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Dockerfile not found in project root")
	}
}

// TestDockerfile_Content — ветка: проверка содержимого Dockerfile
func TestDockerfile_Content(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие основных директив
	requiredDirectives := []string{
		"FROM golang:",
		"FROM alpine:",
		"COPY go.mod",
		"go mod download",
		"go build",
		"ENTRYPOINT",
		"CMD",
	}

	for _, directive := range requiredDirectives {
		if !strings.Contains(contentStr, directive) {
			t.Errorf("Dockerfile missing directive: %s", directive)
		}
	}
}

// TestDockerfile_MultiStage — ветка: multi-stage сборка
func TestDockerfile_MultiStage(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие двух этапов сборки
	fromCount := strings.Count(contentStr, "FROM ")
	if fromCount < 2 {
		t.Errorf("expected at least 2 FROM statements for multi-stage build, got %d", fromCount)
	}

	// Проверяем наличие COPY --from
	if !strings.Contains(contentStr, "COPY --from=builder") {
		t.Error("Dockerfile missing COPY --from=builder directive")
	}
}

// TestDockerfile_Arguments — ветка: аргументы сборки
func TestDockerfile_Arguments(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие аргументов сборки
	arguments := []string{
		"ARG VERSION",
		"ARG BUILD_TIME",
		"ARG GIT_COMMIT",
		"-X main.Version",
		"-X main.BuildTime",
		"-X main.GitCommit",
	}

	for _, arg := range arguments {
		if !strings.Contains(contentStr, arg) {
			t.Errorf("Dockerfile missing ARG or ldflag: %s", arg)
		}
	}
}

// TestDockerfile_Security — ветка: безопасность
func TestDockerfile_Security(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем создание непривилегированного пользователя
	if !strings.Contains(contentStr, "adduser") && !strings.Contains(contentStr, "USER") {
		t.Error("Dockerfile should create non-root user or set USER directive")
	}

	// Проверяем использование USER
	if !strings.Contains(contentStr, "USER scanner") {
		t.Error("Dockerfile should switch to non-root USER")
	}
}

// TestDockerfile_Labels — ветка: метаданные образа
func TestDockerfile_Labels(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие метаданных
	labels := []string{
		"LABEL org.opencontainers.image.title",
		"LABEL org.opencontainers.image.description",
		"LABEL org.opencontainers.image.source",
	}

	for _, label := range labels {
		if !strings.Contains(contentStr, label) {
			t.Errorf("Dockerfile missing label: %s", label)
		}
	}
}

// TestDockerCompose_Exists — ветка: наличие docker-compose.yml
func TestDockerCompose_Exists(t *testing.T) {
	root := getProjectRoot(t)
	path := filepath.Join(root, "docker-compose.yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("docker-compose.yml not found in project root")
	}
}

// TestDockerCompose_Content — ветка: проверка docker-compose.yml
func TestDockerCompose_Content(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("cannot read docker-compose.yml: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие основных директив
	requiredDirectives := []string{
		"services:",
		"build:",
		"dockerfile: Dockerfile",
		"image:",
		"environment:",
		"volumes:",
		"network_mode: host",
		"cap_add:",
		"NET_ADMIN",
		"NET_RAW",
	}

	for _, directive := range requiredDirectives {
		if !strings.Contains(contentStr, directive) {
			t.Errorf("docker-compose.yml missing directive: %s", directive)
		}
	}
}

// TestBuildScript_ShellExists — ветка: наличие shell скрипта
func TestBuildScript_ShellExists(t *testing.T) {
	root := getProjectRoot(t)
	scriptPath := filepath.Join(root, "scripts", "build-docker.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatal("build-docker.sh not found in scripts/")
	}
}

// TestBuildScript_PowerShellExists — ветка: наличие PowerShell скрипта
func TestBuildScript_PowerShellExists(t *testing.T) {
	root := getProjectRoot(t)
	scriptPath := filepath.Join(root, "scripts", "build-docker.ps1")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Fatal("build-docker.ps1 not found in scripts/")
	}
}

// TestBuildScript_ShellContent — ветка: проверка shell скрипта
func TestBuildScript_ShellContent(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "scripts", "build-docker.sh"))
	if err != nil {
		t.Fatalf("cannot read build-docker.sh: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие основных элементов
	requiredElements := []string{
		"docker build",
		"--build-arg",
		"-f Dockerfile",
		"--network host",
		"--cap-add",
	}

	for _, element := range requiredElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("build-docker.sh missing element: %s", element)
		}
	}
}

// TestBuildScript_PowerShellContent — ветка: проверка PowerShell скрипта
func TestBuildScript_PowerShellContent(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "scripts", "build-docker.ps1"))
	if err != nil {
		t.Fatalf("cannot read build-docker.ps1: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие основных элементов
	requiredElements := []string{
		"docker build",
		"--build-arg",
		"-f Dockerfile",
	}

	for _, element := range requiredElements {
		if !strings.Contains(contentStr, element) {
			t.Errorf("build-docker.ps1 missing element: %s", element)
		}
	}
}

// TestDockerfile_CGOEnabled — ветка: CGO_ENABLED
func TestDockerfile_CGOEnabled(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем установку CGO_ENABLED=1 для SQLite
	if !strings.Contains(contentStr, "CGO_ENABLED=1") {
		t.Error("Dockerfile should set CGO_ENABLED=1 for SQLite support")
	}
}

// TestDockerfile_GoOSLinux — ветка: GOOS=linux
func TestDockerfile_GoOSLinux(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем сборку для Linux
	if !strings.Contains(contentStr, "GOOS=linux") {
		t.Error("Dockerfile should set GOOS=linux")
	}
}

// TestDockerfile_MinimalImage — ветка: минимальный образ
func TestDockerfile_MinimalImage(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем использование Alpine для runtime
	if !strings.Contains(contentStr, "FROM alpine:") {
		t.Error("Dockerfile should use Alpine Linux for runtime image")
	}

	// Проверяем установку минимальных зависимостей
	if !strings.Contains(contentStr, "apk add") {
		t.Error("Dockerfile should use apk for package installation")
	}
}

// TestDockerfile_EnvironmentVariables — ветка: переменные окружения
func TestDockerfile_EnvironmentVariables(t *testing.T) {
	root := getProjectRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		t.Fatalf("cannot read Dockerfile: %v", err)
	}

	contentStr := string(content)

	// Проверяем наличие переменных окружения
	envVars := []string{
		"NETWORK_SCANNER_DATA_DIR",
		"NETWORK_SCANNER_CONFIG_DIR",
		"NETWORK_SCANNER_LOG_LEVEL",
	}

	for _, envVar := range envVars {
		if !strings.Contains(contentStr, envVar) {
			t.Errorf("Dockerfile missing ENV: %s", envVar)
		}
	}
}
