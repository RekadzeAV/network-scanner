package configvalidation

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// C3: Тесты для Configuration Schema
// ============================================================================

// TestSchemaRegistry_New — ветка: создание реестра схем
func TestSchemaRegistry_New(t *testing.T) {
	registry := NewSchemaRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

// TestSchemaRegistry_AddSchema — ветка: добавление схемы
func TestSchemaRegistry_AddSchema(t *testing.T) {
	registry := NewSchemaRegistry()

	schema := Schema{
		"port": FieldRule{Required: true, Min: 1, Max: 65535},
	}

	registry.AddSchema("api", schema)

	foundSchema, ok := registry.GetSchema("api")
	if !ok {
		t.Fatal("expected to find schema")
	}
	if len(foundSchema) != 1 {
		t.Errorf("expected 1 field rule, got %d", len(foundSchema))
	}
}

// TestSchemaRegistry_GetSchema_NotFound — ветка: схема не найдена
func TestSchemaRegistry_GetSchema_NotFound(t *testing.T) {
	registry := NewSchemaRegistry()

	_, ok := registry.GetSchema("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent schema")
	}
}

// TestValidateConfig_RequiredField — ветка: обязательное поле
func TestValidateConfig_RequiredField(t *testing.T) {
	schema := Schema{
		"port": FieldRule{Required: true},
	}

	config := map[string]interface{}{}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}

	err := errs[0]
	if err.Code() != "REQUIRED" {
		t.Errorf("expected REQUIRED code, got %s", err.Code())
	}
}

// TestValidateConfig_IntMin — ветка: проверка минимального int
func TestValidateConfig_IntMin(t *testing.T) {
	schema := Schema{
		"port": FieldRule{Min: 1, Max: 65535},
	}

	config := map[string]interface{}{
		"port": 0,
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Code() != "MIN_VALUE" {
		t.Errorf("expected MIN_VALUE code, got %s", errs[0].Code())
	}
}

// TestValidateConfig_IntMax — ветка: проверка максимального int
func TestValidateConfig_IntMax(t *testing.T) {
	schema := Schema{
		"port": FieldRule{Min: 1, Max: 65535},
	}

	config := map[string]interface{}{
		"port": 70000,
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Code() != "MAX_VALUE" {
		t.Errorf("expected MAX_VALUE code, got %s", errs[0].Code())
	}
}

// TestValidateConfig_StringPattern — ветка: проверка pattern
func TestValidateConfig_StringPattern(t *testing.T) {
	schema := Schema{
		"host": FieldRule{Pattern: `^[0-9.]+$`},
	}

	config := map[string]interface{}{
		"host": "invalid-host!",
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Code() != "PATTERN_MISMATCH" {
		t.Errorf("expected PATTERN_MISMATCH code, got %s", errs[0].Code())
	}
}

// TestValidateConfig_StringAllowed — ветка: проверка allowed values
func TestValidateConfig_StringAllowed(t *testing.T) {
	schema := Schema{
		"logLevel": FieldRule{
			Allowed: []string{"debug", "info", "warn", "error"},
		},
	}

	config := map[string]interface{}{
		"logLevel": "verbose",
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Code() != "ALLOWED_VALUES" {
		t.Errorf("expected ALLOWED_VALUES code, got %s", errs[0].Code())
	}
}

// TestValidateConfig_Custom — ветка: custom validation
func TestValidateConfig_Custom(t *testing.T) {
	schema := Schema{
		"port": FieldRule{
			Custom: func(value interface{}) error {
				if v, ok := value.(int); ok && v > 10000 {
					return fmt.Errorf("port must be <= 10000")
				}
				return nil
			},
		},
	}

	config := map[string]interface{}{
		"port": 20000,
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if errs[0].Code() != "CUSTOM" {
		t.Errorf("expected CUSTOM code, got %s", errs[0].Code())
	}
}

// TestValidateConfig_Valid — ветка: валидная config
func TestValidateConfig_Valid(t *testing.T) {
	schema := Schema{
		"port": FieldRule{Required: true, Min: 1, Max: 65535},
		"host": FieldRule{Required: true},
	}

	config := map[string]interface{}{
		"port": 8080,
		"host": "localhost",
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

// TestValidationError_Error — ветка: формат ошибки
func TestValidationError_Error(t *testing.T) {
	err := NewValidationError("port", "must be between 1 and 65535", "INVALID_PORT")

	errStr := err.Error()
	if errStr == "" {
		t.Error("expected non-empty error string")
	}
	if err.Field() != "port" {
		t.Errorf("expected field port, got %s", err.Field())
	}
	if err.Message() != "must be between 1 and 65535" {
		t.Errorf("unexpected message")
	}
	if err.Code() != "INVALID_PORT" {
		t.Errorf("expected code INVALID_PORT, got %s", err.Code())
	}
}

// TestValidationErrors_IsValid — ветка: проверка IsValid
func TestValidationErrors_IsValid(t *testing.T) {
	var errs ValidationErrors

	if !errs.IsValid() {
		t.Error("expected empty errors to be valid")
	}

	errs = append(errs, NewValidationError("port", "invalid", "INVALID"))
	if errs.IsValid() {
		t.Error("expected errors to be invalid")
	}
}

// TestValidatePort — ветка: валидация порта
func TestValidatePort(t *testing.T) {
	// Валидные порты
	if err := ValidatePort(80); err != nil {
		t.Error("expected port 80 to be valid")
	}
	if err := ValidatePort(8080); err != nil {
		t.Error("expected port 8080 to be valid")
	}
	if err := ValidatePort(65535); err != nil {
		t.Error("expected port 65535 to be valid")
	}

	// Невалидные порты
	if err := ValidatePort(0); err == nil {
		t.Error("expected port 0 to be invalid")
	}
	if err := ValidatePort(-1); err == nil {
		t.Error("expected port -1 to be invalid")
	}
	if err := ValidatePort(70000); err == nil {
		t.Error("expected port 70000 to be invalid")
	}
}

// TestValidateHost — ветка: валидация host
func TestValidateHost(t *testing.T) {
	// Валидные host
	if err := ValidateHost("localhost"); err != nil {
		t.Error("expected localhost to be valid")
	}
	if err := ValidateHost("192.168.1.1"); err != nil {
		t.Error("expected 192.168.1.1 to be valid")
	}
	if err := ValidateHost("example.com"); err != nil {
		t.Error("expected example.com to be valid")
	}

	// Невалидные host
	if err := ValidateHost(""); err == nil {
		t.Error("expected empty host to be invalid")
	}
	if err := ValidateHost("invalid host!"); err == nil {
		t.Error("expected invalid host to be invalid")
	}
}

// TestValidateDuration — ветка: валидация duration
func TestValidateDuration(t *testing.T) {
	// Валидные durations
	if err := ValidateDuration(1 * time.Second); err != nil {
		t.Error("expected 1s duration to be valid")
	}
	if err := ValidateDuration(10 * time.Second); err != nil {
		t.Error("expected 10s duration to be valid")
	}

	// Невалидные durations
	if err := ValidateDuration(0); err == nil {
		t.Error("expected 0 duration to be invalid")
	}
	if err := ValidateDuration(-1 * time.Second); err == nil {
		t.Error("expected negative duration to be invalid")
	}
	if err := ValidateDuration(301 * time.Second); err == nil {
		t.Error("expected 301s duration to be invalid (> 300s)")
	}
}

// TestValidateCIDR — ветка: валидация CIDR
func TestValidateCIDR(t *testing.T) {
	// Валидные CIDR
	if err := ValidateCIDR("192.168.1.0/24"); err != nil {
		t.Error("expected 192.168.1.0/24 to be valid")
	}
	if err := ValidateCIDR("10.0.0.0/8"); err != nil {
		t.Error("expected 10.0.0.0/8 to be valid")
	}

	// Невалидные CIDR
	if err := ValidateCIDR(""); err == nil {
		t.Error("expected empty CIDR to be invalid")
	}
	if err := ValidateCIDR("invalid"); err == nil {
		t.Error("expected invalid CIDR to be invalid")
	}
	if err := ValidateCIDR("192.168.1.1"); err == nil {
		t.Error("expected IP without mask to be invalid")
	}
}

// TestValidatePortRange — ветка: валидация диапазона портов
func TestValidatePortRange(t *testing.T) {
	// Валидные диапазоны
	if err := ValidatePortRange("1-1024"); err != nil {
		t.Errorf("expected 1-1024 to be valid: %v", err)
	}
	if err := ValidatePortRange("80,443,8080"); err != nil {
		t.Errorf("expected 80,443,8080 to be valid: %v", err)
	}
	if err := ValidatePortRange("*"); err != nil {
		t.Error("expected * to be valid")
	}
	if err := ValidatePortRange(""); err != nil {
		t.Error("expected empty to be valid")
	}

	// Невалидные диапазоны
	if err := ValidatePortRange("1024-1"); err == nil {
		t.Error("expected 1024-1 (reversed) to be invalid")
	}
	if err := ValidatePortRange("0-1024"); err == nil {
		t.Error("expected 0-1024 (port 0) to be invalid")
	}
	if err := ValidatePortRange("1-70000"); err == nil {
		t.Error("expected 1-70000 (port > 65535) to be invalid")
	}
	if err := ValidatePortRange("abc"); err == nil {
		t.Error("expected abc to be invalid")
	}
}

// TestValidateConfig_StringLength — ветка: проверка длины string
func TestValidateConfig_StringLength(t *testing.T) {
	schema := Schema{
		"hostname": FieldRule{Min: 3, Max: 50},
	}

	// Слишком короткий
	config := map[string]interface{}{
		"hostname": "ab",
	}
	errs := ValidateConfig(config, schema)
	if len(errs) != 1 || errs[0].Code() != "MIN_LENGTH" {
		t.Errorf("expected MIN_LENGTH error, got %v", errs)
	}

	// Слишком длинный
	config = map[string]interface{}{
		"hostname": strings.Repeat("a", 51),
	}
	errs = ValidateConfig(config, schema)
	if len(errs) != 1 || errs[0].Code() != "MAX_LENGTH" {
		t.Errorf("expected MAX_LENGTH error, got %v", errs)
	}

	// Валидный
	config = map[string]interface{}{
		"hostname": "server",
	}
	errs = ValidateConfig(config, schema)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

// TestValidateConfig_NonRequiredMissing — ветка: необязательное поле
func TestValidateConfig_NonRequiredMissing(t *testing.T) {
	schema := Schema{
		"port": FieldRule{Required: true},
		"host": FieldRule{Required: false},
	}

	config := map[string]interface{}{
		"port": 8080,
	}

	errs := ValidateConfig(config, schema)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors (host is not required), got %d", len(errs))
	}
}

// TestValidateConfig_Description — ветка: description в ошибке
func TestValidateConfig_Description(t *testing.T) {
	schema := Schema{
		"port": FieldRule{
			Required:    true,
			Description: "API port number",
		},
	}

	config := map[string]interface{}{}

	errs := ValidateConfig(config, schema)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if !strings.Contains(errs[0].Message(), "API port number") {
		t.Errorf("expected description in error message, got: %s", errs[0].Message())
	}
}

// TestValidatePortRange_EdgeCases — ветка: крайние случаи диапазона портов
func TestValidatePortRange_EdgeCases(t *testing.T) {
	// Крайние валидные значения
	if err := ValidatePortRange("1-65535"); err != nil {
		t.Errorf("expected 1-65535 to be valid: %v", err)
	}
	if err := ValidatePortRange("1"); err != nil {
		t.Error("expected single port 1 to be valid")
	}
	if err := ValidatePortRange("65535"); err != nil {
		t.Error("expected single port 65535 to be valid")
	}

	// Крайние невалидные значения
	if err := ValidatePortRange("0"); err == nil {
		t.Error("expected port 0 to be invalid")
	}
	if err := ValidatePortRange("65536"); err == nil {
		t.Error("expected port 65536 to be invalid")
	}
}
