// Package configvalidation предоставляет систему валидации конфигурации
// через JSON schema и custom rules.
//
// # Основные компоненты
//
// Schema — структура для определения правил валидации
//
//	FieldRules — map name -> rules для каждого поля
//	Validate() — валидация config структур
//
// Validator — валидатор для config объектов
//
//	NewValidator() — создает новый валидатор
//	AddSchema() — добавление схемы валидации
//	Validate() — валидация config
//
// ValidationError — ошибка валидации
//
//	Field() — имя поля с ошибкой
//	Message() — описание ошибки
//	Code() — код ошибки
//
// # Пример использования
//
//	schema := Schema{
//	    "port": {
//	        Required: true,
//	        Min:      1,
//	        Max:      65535,
//	    },
//	    "host": {
//	        Required: true,
//	        Pattern:  `^[0-9.]+$`,
//	    },
//	}
//
//	validator := configvalidation.NewValidator()
//	validator.AddSchema("api", schema)
//
//	err := validator.Validate(apiConfig)
//	if err != nil {
//	    // Обработка ошибок валидации
//	}

package configvalidation

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"
)

// FieldRule определяет правила валидации для одного поля
type FieldRule struct {
	// Required поле обязательно
	Required bool
	// Min минимальное значение (для int, string length)
	Min interface{}
	// Max максимальное значение (для int, string length)
	Max interface{}
	// Pattern regex pattern (для string)
	Pattern string
	// Allowed allowed values (для enum)
	Allowed []string
	// Custom custom validation function
	Custom func(value interface{}) error
	// Description описание поля для сообщений об ошибках
	Description string
}

// Schema определяет схему валидации для config структуры
type Schema map[string]FieldRule

// SchemaRegistry хранит схемы валидации для разных типов config
type SchemaRegistry struct {
	schemas map[string]Schema
}

// NewSchemaRegistry создает новый реестр схем
func NewSchemaRegistry() *SchemaRegistry {
	return &SchemaRegistry{
		schemas: make(map[string]Schema),
	}
}

// AddSchema добавляет схему валидации
func (sr *SchemaRegistry) AddSchema(name string, schema Schema) {
	sr.schemas[name] = schema
}

// GetSchema возвращает схему по имени
func (sr *SchemaRegistry) GetSchema(name string) (Schema, bool) {
	schema, ok := sr.schemas[name]
	return schema, ok
}

// ValidationError — ошибка валидации
type ValidationError struct {
	field   string
	message string
	code    string
}

// NewValidationError создает новую ошибку валидации
func NewValidationError(field, message, code string) *ValidationError {
	return &ValidationError{
		field:   field,
		message: message,
		code:    code,
	}
}

func (e *ValidationError) Field() string {
	return e.field
}

func (e *ValidationError) Message() string {
	return e.message
}

func (e *ValidationError) Code() string {
	return e.code
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.code, e.field, e.message)
}

// ValidationErrors — список ошибок валидации
type ValidationErrors []*ValidationError

func (errs ValidationErrors) Error() string {
	messages := make([]string, len(errs))
	for i, err := range errs {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// IsValid возвращает true если ошибок нет
func (errs ValidationErrors) IsValid() bool {
	return len(errs) == 0
}

// ValidateConfig валидирует config map по схеме
func ValidateConfig(config map[string]interface{}, schema Schema) ValidationErrors {
	var errs ValidationErrors

	for fieldName, rule := range schema {
		value, exists := config[fieldName]

		// Проверка Required
		if rule.Required && !exists {
			desc := rule.Description
			if desc == "" {
				desc = fieldName
			}
			errs = append(errs, NewValidationError(
				fieldName,
				fmt.Sprintf("required field %s is missing", desc),
				"REQUIRED",
			))
			continue
		}

		// Если поле не существует и не required — пропускаем
		if !exists {
			continue
		}

		// Валидация значения
		fieldErrs := validateFieldValue(fieldName, value, rule)
		errs = append(errs, fieldErrs...)
	}

	return errs
}

// validateFieldValue валидирует значение поля по правилам
func validateFieldValue(fieldName string, value interface{}, rule FieldRule) ValidationErrors {
	var errs ValidationErrors

	// Валидация int
	if intValue, ok := value.(int); ok {
		if rule.Min != nil {
			if min, ok := rule.Min.(int); ok && intValue < min {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("value %d must be >= %d", intValue, min),
					"MIN_VALUE",
				))
			}
		}
		if rule.Max != nil {
			if max, ok := rule.Max.(int); ok && intValue > max {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("value %d must be <= %d", intValue, max),
					"MAX_VALUE",
				))
			}
		}
	}

	// Валидация string
	if strValue, ok := value.(string); ok {
		// Length validation
		if rule.Min != nil {
			if min, ok := rule.Min.(int); ok && len(strValue) < min {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("length %d must be >= %d", len(strValue), min),
					"MIN_LENGTH",
				))
			}
		}
		if rule.Max != nil {
			if max, ok := rule.Max.(int); ok && len(strValue) > max {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("length %d must be <= %d", len(strValue), max),
					"MAX_LENGTH",
				))
			}
		}

		// Pattern validation
		if rule.Pattern != "" {
			matched, err := regexp.MatchString(rule.Pattern, strValue)
			if err != nil {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("invalid pattern: %s", err),
					"PATTERN_ERROR",
				))
			} else if !matched {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("value %q does not match pattern %s", strValue, rule.Pattern),
					"PATTERN_MISMATCH",
				))
			}
		}

		// Allowed values validation
		if len(rule.Allowed) > 0 {
			found := false
			for _, allowed := range rule.Allowed {
				if strValue == allowed {
					found = true
					break
				}
			}
			if !found {
				errs = append(errs, NewValidationError(
					fieldName,
					fmt.Sprintf("value %q is not in allowed values: %v", strValue, rule.Allowed),
					"ALLOWED_VALUES",
				))
			}
		}
	}

	// Custom validation
	if rule.Custom != nil {
		if err := rule.Custom(value); err != nil {
			errs = append(errs, NewValidationError(
				fieldName,
				err.Error(),
				"CUSTOM",
			))
		}
	}

	return errs
}

// ============================================================================
// Специфичные валидаторы для типов config
// ============================================================================

// ValidatePort валидирует порт
func ValidatePort(port int) *ValidationError {
	if port < 1 || port > 65535 {
		return NewValidationError(
			"port",
			fmt.Sprintf("port must be between 1 and 65535, got %d", port),
			"INVALID_PORT",
		)
	}
	return nil
}

// ValidateHost валидирует host
func ValidateHost(host string) *ValidationError {
	if host == "" {
		return NewValidationError(
			"host",
			"host cannot be empty",
			"EMPTY_HOST",
		)
	}

	// Проверяем IP адрес
	if ip := net.ParseIP(host); ip != nil {
		return nil // Валидный IP
	}

	// Проверяем hostname
	if regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]+$`).MatchString(host) {
		return nil // Валидный hostname
	}

	return NewValidationError(
		"host",
		fmt.Sprintf("invalid host: %s", host),
		"INVALID_HOST",
	)
}

// ValidateDuration валидирует duration
func ValidateDuration(duration time.Duration) *ValidationError {
	if duration <= 0 {
		return NewValidationError(
			"duration",
			"duration must be positive",
			"INVALID_DURATION",
		)
	}
	if duration > 300*time.Second {
		return NewValidationError(
			"duration",
			"duration must not exceed 300s",
			"DURATION_TOO_LONG",
		)
	}
	return nil
}

// ValidateCIDR валидирует CIDR диапазон
func ValidateCIDR(cidr string) *ValidationError {
	if cidr == "" {
		return NewValidationError(
			"cidr",
			"cidr cannot be empty",
			"EMPTY_CIDR",
		)
	}

	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return NewValidationError(
			"cidr",
			fmt.Sprintf("invalid CIDR: %s", cidr),
			"INVALID_CIDR",
		)
	}
	return nil
}

// ValidatePortRange валидирует диапазон портов
func ValidatePortRange(portRange string) *ValidationError {
	if portRange == "" {
		return nil // Empty is allowed
	}

	// Проверяем формат "1-1024" или "80,443,8080" или "*"
	if portRange == "*" {
		return nil
	}

	// Проверяем список портов
	if strings.Contains(portRange, ",") {
		parts := strings.Split(portRange, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if _, err := fmt.Sscanf(part, "%d", new(int)); err != nil {
				return NewValidationError(
					"portRange",
					fmt.Sprintf("invalid port in list: %s", part),
					"INVALID_PORT",
				)
			}
		}
		return nil
	}

	// Проверяем диапазон
	if strings.Contains(portRange, "-") {
		parts := strings.Split(portRange, "-")
		if len(parts) != 2 {
			return NewValidationError(
				"portRange",
				fmt.Sprintf("invalid port range: %s", portRange),
				"INVALID_RANGE",
			)
		}

		var start, end int
		_, err := fmt.Sscanf(parts[0], "%d", &start)
		if err != nil {
			return NewValidationError(
				"portRange",
				fmt.Sprintf("invalid port range start: %s", parts[0]),
				"INVALID_PORT",
			)
		}
		_, err = fmt.Sscanf(parts[1], "%d", &end)
		if err != nil {
			return NewValidationError(
				"portRange",
				fmt.Sprintf("invalid port range end: %s", parts[1]),
				"INVALID_PORT",
			)
		}

		if start > end {
			return NewValidationError(
				"portRange",
				fmt.Sprintf("range start %d must be <= end %d", start, end),
				"INVALID_RANGE",
			)
		}
		if start < 1 || end > 65535 {
			return NewValidationError(
				"portRange",
				"ports must be between 1 and 65535",
				"PORT_OUT_OF_RANGE",
			)
		}

		return nil
	}

	// Одиночный порт
	var port int
	_, err := fmt.Sscanf(portRange, "%d", &port)
	if err != nil {
		return NewValidationError(
			"portRange",
			fmt.Sprintf("invalid port: %s", portRange),
			"INVALID_PORT",
		)
	}

	if port < 1 || port > 65535 {
		return NewValidationError(
			"portRange",
			"port must be between 1 and 65535",
			"PORT_OUT_OF_RANGE",
		)
	}

	return nil
}
