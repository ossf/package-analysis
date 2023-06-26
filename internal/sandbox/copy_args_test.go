package sandbox

import (
	"reflect"
	"testing"
)

type copyCmdTestCase struct {
	name          string
	hostPath      string
	containerPath string
	containerId   string
	want          []string
}

func Test_containerToHostCopyCmdArgs(t *testing.T) {
	tests := []copyCmdTestCase{
		{
			name:          "simple relative path",
			hostPath:      "path/in/host",
			containerPath: "path/in/container",
			containerId:   "12345",
			want:          []string{"cp", "12345:path/in/container", "path/in/host"},
		},
		{
			name:          "simple absolute path",
			hostPath:      "/dest/path/in/host",
			containerPath: "/src/path/in/container",
			containerId:   "abcde",
			want:          []string{"cp", "abcde:/src/path/in/container", "/dest/path/in/host"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containerToHostCopyCmd(tt.hostPath, tt.containerPath, tt.containerId).Args()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("containerToHostCopyCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hostToContainerCopyCmdArgs(t *testing.T) {
	tests := []copyCmdTestCase{
		{
			name:          "simple relative path",
			hostPath:      "/src",
			containerPath: "/dest",
			containerId:   "12345",
			want:          []string{"cp", "/src", "12345:/dest"},
		},
		{
			name:          "simple absolute path",
			hostPath:      "/src/path/in/host",
			containerPath: "/dest/path/in/container",
			containerId:   "abcde",
			want:          []string{"cp", "/src/path/in/host", "abcde:/dest/path/in/container"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hostToContainerCopyCmd(tt.hostPath, tt.containerPath, tt.containerId).Args()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("hostToContainerCopyCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
