package alerting_test

import (
	"fmt"

	"network-scanner/internal/alerting"
)

// ExampleRuleType демонстрирует типы правил алертинга.
func ExampleRuleType() {
	rules := []alerting.RuleType{
		alerting.RuleTypeNewHost,
		alerting.RuleTypeNewPort,
		alerting.RuleTypePortClosed,
		alerting.RuleTypeDeviceRemoved,
		alerting.RuleTypeOSChanged,
		alerting.RuleTypeHostnameChanged,
	}

	for _, rule := range rules {
		fmt.Printf("%s\n", rule)
	}

	// Output:
	// new_host
	// new_port
	// port_closed
	// device_removed
	// os_changed
	// hostname_changed
}

// ExampleSeverity демонстрирует уровни серьёзности алертов.
func ExampleSeverity() {
	severities := []alerting.Severity{
		alerting.SeverityLow,
		alerting.SeverityMedium,
		alerting.SeverityHigh,
		alerting.SeverityCritical,
	}

	for _, s := range severities {
		fmt.Println(s)
	}

	// Output:
	// LOW
	// MEDIUM
	// HIGH
	// CRITICAL
}
