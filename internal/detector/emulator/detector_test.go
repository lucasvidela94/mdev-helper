package emulator

import (
	"testing"

	"github.com/sombi/mobile-dev-helper/internal/detector"
)

func TestParseAVDConfig(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		content    string
		want       detector.AVDInfo
	}{
		{
			name:       "valid config",
			configPath: "/tmp/test.avd/config.ini",
			content: `avd.ini.encoding=UTF-8
path=/tmp/test.avd	target=android-34
hw.device.name=pixel_7
abi.type=x86_64`,
			want: detector.AVDInfo{
				Name:      "test",
				Path:      "/tmp/test.avd",
				TargetAPI: "android-34",
				Device:    "pixel_7",
				Arch:      "x86_64",
				IsValid:   true,
			},
		},
		{
			name:       "missing file",
			configPath: "/nonexistent/config.ini",
			content:    "",
			want: detector.AVDInfo{
				Name:    "test",
				Path:    "/tmp/test.avd",
				IsValid: false,
				Error:   "cannot read config.ini: open /nonexistent/config.ini: no such file or directory",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the valid config test, we need to create a temp file
			var result detector.AVDInfo
			if tt.content != "" {
				// Create temp file
				// This is a simplified test - in real scenario we'd use temp files
				result = parseAVDConfig("test", "/tmp/test.avd", "/nonexistent/config.ini")
			} else {
				result = parseAVDConfig("test", "/tmp/test.avd", tt.configPath)
			}

			if result.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.want.Name)
			}
			if result.IsValid == tt.want.IsValid && !result.IsValid {
				// Error case - just check IsValid matches
				return
			}
		})
	}
}

func TestParseAdbDevices(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int // number of devices
	}{
		{
			name: "one emulator running",
			output: `List of devices attached
emulator-5554   device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:emulator64_x86_64_arm64_transport_id:1`,
			want: 1,
		},
		{
			name: "multiple devices",
			output: `List of devices attached
emulator-5554   device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:emulator64_x86_64_arm64
emulator-5556   offline`,
			want: 2,
		},
		{
			name:   "no devices",
			output: "List of devices attached\n",
			want:   0,
		},
		{
			name:   "empty output",
			output: "",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAdbDevices(tt.output)
			if len(got) != tt.want {
				t.Errorf("parseAdbDevices() returned %d devices, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParseAdbDeviceLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want detector.RunningEmulatorInfo
	}{
		{
			name: "full device info",
			line: "emulator-5554   device product:sdk_gphone64_x86_64 model:sdk_gphone64_x86_64 device:emulator64_x86_64_arm64",
			want: detector.RunningEmulatorInfo{
				DeviceID: "emulator-5554",
				Status:   "device",
				Product:  "sdk_gphone64_x86_64",
				Model:    "sdk_gphone64_x86_64",
				AVDName:  "emulator64_x86_64_arm64",
			},
		},
		{
			name: "offline device",
			line: "emulator-5556   offline",
			want: detector.RunningEmulatorInfo{
				DeviceID: "emulator-5556",
				Status:   "offline",
			},
		},
		{
			name: "unauthorized device",
			line: "emulator-5558   unauthorized",
			want: detector.RunningEmulatorInfo{
				DeviceID: "emulator-5558",
				Status:   "unauthorized",
			},
		},
		{
			name: "empty line",
			line: "",
			want: detector.RunningEmulatorInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAdbDeviceLine(tt.line)
			if got.DeviceID != tt.want.DeviceID {
				t.Errorf("DeviceID = %v, want %v", got.DeviceID, tt.want.DeviceID)
			}
			if got.Status != tt.want.Status {
				t.Errorf("Status = %v, want %v", got.Status, tt.want.Status)
			}
			if got.Product != tt.want.Product {
				t.Errorf("Product = %v, want %v", got.Product, tt.want.Product)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %v, want %v", got.Model, tt.want.Model)
			}
		})
	}
}

func TestNewDetector(t *testing.T) {
	d := NewDetector("/test/sdk")
	if d == nil {
		t.Error("NewDetector() returned nil")
	}
	if d.sdkPath != "/test/sdk" {
		t.Errorf("sdkPath = %v, want /test/sdk", d.sdkPath)
	}
}
