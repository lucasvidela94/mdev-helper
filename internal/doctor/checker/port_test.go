package checker

import (
	"fmt"
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

func TestPortCheckerName(t *testing.T) {
	checker := NewPortChecker()
	if checker.Name() != "Port Availability" {
		t.Errorf("Expected name 'Port Availability', got %s", checker.Name())
	}
}

func TestPortCheckerCategory(t *testing.T) {
	checker := NewPortChecker()
	if checker.Category() != doctor.CategoryEnvironment {
		t.Errorf("Expected category 'environment', got %s", checker.Category())
	}
}

func TestPortCheckerWithDefaultPorts(t *testing.T) {
	checker := NewPortChecker()

	// Should have default ports
	if len(checker.ports) != 2 {
		t.Errorf("Expected 2 default ports, got %d", len(checker.ports))
	}

	// Should include Metro and ADB ports
	hasMetro := false
	hasADB := false
	for _, port := range checker.ports {
		if port == MetroPort {
			hasMetro = true
		}
		if port == ADBServerPort {
			hasADB = true
		}
	}

	if !hasMetro {
		t.Error("Expected default ports to include Metro port (8081)")
	}
	if !hasADB {
		t.Error("Expected default ports to include ADB port (5037)")
	}
}

func TestPortCheckerWithCustomPorts(t *testing.T) {
	customPorts := []int{3000, 8080, 9000}
	checker := NewPortChecker(customPorts...)

	if len(checker.ports) != 3 {
		t.Errorf("Expected 3 custom ports, got %d", len(checker.ports))
	}

	for i, port := range customPorts {
		if checker.ports[i] != port {
			t.Errorf("Expected port %d at index %d, got %d", port, i, checker.ports[i])
		}
	}
}

func TestPortCheckerCheck(t *testing.T) {
	// Use a high port that's likely available
	checker := NewPortChecker(65534)

	result := checker.Check()

	// Should have a name
	if result.Name != "Port Availability" {
		t.Errorf("Expected result name 'Port Availability', got %s", result.Name)
	}

	// Should have valid status
	if result.Status != doctor.StatusPassed && result.Status != doctor.StatusWarning {
		t.Errorf("Expected status 'passed' or 'warning', got %s", result.Status)
	}

	// Should have a message
	if result.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Should have details
	if result.Details == nil {
		t.Fatal("Expected details to be present")
	}

	// Check required detail fields
	if _, ok := result.Details["portsChecked"]; !ok {
		t.Error("Expected 'portsChecked' in details")
	}
	if _, ok := result.Details["portResults"]; !ok {
		t.Error("Expected 'portResults' in details")
	}
}

func TestCheckPort(t *testing.T) {
	// Test with a high port that's likely available
	status := checkPort(65533)

	if status.Port != 65533 {
		t.Errorf("Expected port 65533, got %d", status.Port)
	}

	if status.State != PortAvailable {
		t.Errorf("Expected port to be available, got %s", status.State)
	}

	if status.Protocol != "tcp" {
		t.Errorf("Expected protocol 'tcp', got %s", status.Protocol)
	}
}

func TestIdentifyProcess(t *testing.T) {
	tests := []struct {
		port     int
		expected string
	}{
		{MetroPort, "Metro bundler"},
		{ADBServerPort, "ADB server"},
		{5554, "Android emulator"},
		{5556, "Android emulator"},
		{5585, "Android emulator"},
		{9999, "unknown"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Port%d", tt.port), func(t *testing.T) {
			process := identifyProcess(tt.port)
			if process != tt.expected {
				t.Errorf("Expected process '%s' for port %d, got '%s'", tt.expected, tt.port, process)
			}
		})
	}
}

func TestIsExpectedPortInUse(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		status   PortStatus
		expected bool
	}{
		{
			name:     "ADB server expected",
			port:     ADBServerPort,
			status:   PortStatus{Process: "ADB server"},
			expected: true,
		},
		{
			name:     "Metro expected",
			port:     MetroPort,
			status:   PortStatus{Process: "Metro bundler"},
			expected: true,
		},
		{
			name:     "Emulator port expected",
			port:     5554,
			status:   PortStatus{Process: "Android emulator"},
			expected: true,
		},
		{
			name:     "Random port not expected",
			port:     3000,
			status:   PortStatus{Process: "node"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExpectedPortInUse(tt.port, tt.status)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetDefaultPorts(t *testing.T) {
	ports := GetDefaultPorts()

	if len(ports) != 2 {
		t.Errorf("Expected 2 default ports, got %d", len(ports))
	}

	if ports[0] != MetroPort {
		t.Errorf("Expected first port to be %d, got %d", MetroPort, ports[0])
	}

	if ports[1] != ADBServerPort {
		t.Errorf("Expected second port to be %d, got %d", ADBServerPort, ports[1])
	}
}

func TestGetEmulatorPorts(t *testing.T) {
	ports := GetEmulatorPorts()

	// Should have ports from 5554 to 5585, stepping by 2
	expectedCount := (ADBEmulatorEndPort-ADBEmulatorStartPort)/2 + 1
	if len(ports) != expectedCount {
		t.Errorf("Expected %d emulator ports, got %d", expectedCount, len(ports))
	}

	// First port should be 5554
	if ports[0] != ADBEmulatorStartPort {
		t.Errorf("Expected first port to be %d, got %d", ADBEmulatorStartPort, ports[0])
	}

	// Last port should be 5584 (5585 is odd, not included when stepping by 2 from 5554)
	expectedLastPort := ADBEmulatorEndPort - 1
	if ports[len(ports)-1] != expectedLastPort {
		t.Errorf("Expected last port to be %d, got %d", expectedLastPort, ports[len(ports)-1])
	}
}

func TestParsePortRange(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedLen int
	}{
		{"Valid range", "5554-5558", false, 5},
		{"Single port range", "8080-8080", false, 1},
		{"Invalid format", "8080", true, 0},
		{"Invalid start", "abc-8080", true, 0},
		{"Invalid end", "8080-xyz", true, 0},
		{"Reversed range", "8080-8070", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports, err := ParsePortRange(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(ports) != tt.expectedLen {
					t.Errorf("Expected %d ports, got %d", tt.expectedLen, len(ports))
				}
			}
		})
	}
}

func TestPortStateConstants(t *testing.T) {
	if string(PortAvailable) != "available" {
		t.Errorf("Expected PortAvailable to be 'available', got '%s'", PortAvailable)
	}
	if string(PortInUse) != "in_use" {
		t.Errorf("Expected PortInUse to be 'in_use', got '%s'", PortInUse)
	}
	if string(PortUnavailable) != "unavailable" {
		t.Errorf("Expected PortUnavailable to be 'unavailable', got '%s'", PortUnavailable)
	}
}

// BenchmarkCheckPort benchmarks the port checking function
func BenchmarkCheckPort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = checkPort(65532)
	}
}
