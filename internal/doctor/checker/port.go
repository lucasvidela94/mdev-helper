package checker

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/sombi/mobile-dev-helper/internal/doctor"
)

const (
	// MetroPort is the default port for Metro bundler.
	MetroPort = 8081
	// ADBServerPort is the default port for ADB server.
	ADBServerPort = 5037
	// ADBEmulatorStartPort is the start of the ADB emulator port range.
	ADBEmulatorStartPort = 5554
	// ADBEmulatorEndPort is the end of the ADB emulator port range.
	ADBEmulatorEndPort = 5585
)

// PortChecker checks for port availability.
type PortChecker struct {
	ports []int
}

// NewPortChecker creates a new port checker for the specified ports.
func NewPortChecker(ports ...int) *PortChecker {
	if len(ports) == 0 {
		// Default ports to check
		ports = []int{MetroPort, ADBServerPort}
	}
	return &PortChecker{ports: ports}
}

// Name returns the checker name.
func (p *PortChecker) Name() string {
	return "Port Availability"
}

// Category returns the checker category.
func (p *PortChecker) Category() doctor.CheckCategory {
	return doctor.CategoryEnvironment
}

// Check performs the port availability check.
func (p *PortChecker) Check() doctor.CheckResult {
	results := make(map[int]PortStatus)
	var unavailablePorts []int
	var warningPorts []int

	for _, port := range p.ports {
		status := checkPort(port)
		results[port] = status

		switch status.State {
		case PortInUse:
			// Some ports are expected to be in use (like ADB server)
			if isExpectedPortInUse(port, status) {
				// This is expected, not an error
			} else {
				warningPorts = append(warningPorts, port)
			}
		case PortUnavailable:
			unavailablePorts = append(unavailablePorts, port)
		}
	}

	result := doctor.CheckResult{
		Name: p.Name(),
		Details: map[string]interface{}{
			"portsChecked": len(p.ports),
			"portResults":  results,
		},
	}

	if len(unavailablePorts) > 0 {
		result.Status = doctor.StatusError
		result.Message = fmt.Sprintf("Ports unavailable: %v", unavailablePorts)
	} else if len(warningPorts) > 0 {
		result.Status = doctor.StatusWarning
		result.Message = fmt.Sprintf("Ports in use (may conflict): %v", warningPorts)
	} else {
		result.Status = doctor.StatusPassed
		result.Message = "All required ports are available"
	}

	return result
}

// PortState represents the state of a port.
type PortState string

const (
	// PortAvailable means the port is free.
	PortAvailable PortState = "available"
	// PortInUse means the port is in use by a known process.
	PortInUse PortState = "in_use"
	// PortUnavailable means the port cannot be used.
	PortUnavailable PortState = "unavailable"
)

// PortStatus contains information about a port's status.
type PortStatus struct {
	Port     int       `json:"port"`
	State    PortState `json:"state"`
	Process  string    `json:"process,omitempty"`
	Protocol string    `json:"protocol,omitempty"`
}

// checkPort checks the status of a single port.
func checkPort(port int) PortStatus {
	// Try to listen on the port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Port is in use, try to identify the process
		process := identifyProcess(port)
		return PortStatus{
			Port:     port,
			State:    PortInUse,
			Process:  process,
			Protocol: "tcp",
		}
	}
	defer listener.Close()

	return PortStatus{
		Port:     port,
		State:    PortAvailable,
		Protocol: "tcp",
	}
}

// identifyProcess attempts to identify what process is using a port.
// This is a simplified version - in production you'd parse netstat/ss output.
func identifyProcess(port int) string {
	// Known ports and their expected processes
	knownPorts := map[int]string{
		MetroPort:     "Metro bundler",
		ADBServerPort: "ADB server",
	}

	if process, ok := knownPorts[port]; ok {
		return process
	}

	// Check if it's in the emulator range
	if port >= ADBEmulatorStartPort && port <= ADBEmulatorEndPort {
		return "Android emulator"
	}

	return "unknown"
}

// isExpectedPortInUse returns true if the port being in use is expected.
func isExpectedPortInUse(port int, status PortStatus) bool {
	// ADB server is expected to be running
	if port == ADBServerPort {
		return status.Process == "ADB server"
	}

	// Metro might be running intentionally
	if port == MetroPort {
		return status.Process == "Metro bundler"
	}

	// Emulator ports are expected to be in use when emulators are running
	if port >= ADBEmulatorStartPort && port <= ADBEmulatorEndPort {
		return status.Process == "Android emulator"
	}

	return false
}

// GetDefaultPorts returns the list of default ports to check.
func GetDefaultPorts() []int {
	return []int{MetroPort, ADBServerPort}
}

// GetEmulatorPorts returns the range of emulator ports.
func GetEmulatorPorts() []int {
	ports := make([]int, 0, (ADBEmulatorEndPort-ADBEmulatorStartPort)/2+1)
	for port := ADBEmulatorStartPort; port <= ADBEmulatorEndPort; port += 2 {
		ports = append(ports, port)
	}
	return ports
}

// ParsePortRange parses a port range string like "5554-5585".
func ParsePortRange(rangeStr string) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid port range format: %s", rangeStr)
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start port: %v", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end port: %v", err)
	}

	if start > end {
		return nil, fmt.Errorf("start port must be less than or equal to end port")
	}

	ports := make([]int, 0, end-start+1)
	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}

	return ports, nil
}
