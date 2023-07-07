package sandbox

import (
	"fmt"
	"strings"
)

// copySpec specifies the source and destination of a copy operation.
// The copy may be made from the host into the sandbox or vice versa.
// See https://docs.podman.io/en/latest/markdown/podman-cp.1.html for
// semantics of src and dest paths.
// srcInContainer and destInContainer specify whether the copy source
// and destination are respectively in the host (false) or container (true)
type copySpec struct {
	src             string
	dest            string
	srcInContainer  bool
	destInContainer bool
	containerId     string
}

func (c copySpec) Args() []string {
	copySrc := c.src
	if c.srcInContainer {
		copySrc = fmt.Sprintf("%s:%s", c.containerId, c.src)
	}

	copyDest := c.dest
	if c.destInContainer {
		copyDest = fmt.Sprintf("%s:%s", c.containerId, c.dest)
	}

	return []string{"cp", copySrc, copyDest}
}

func (c copySpec) String() string {
	return strings.Join(c.Args(), " ")
}

// hostToContainerCopyCmd generates the arguments to podman
// that copy a file from the host to the container.
func hostToContainerCopyCmd(hostPath, containerPath, containerId string) copySpec {
	return copySpec{hostPath, containerPath, false, true, containerId}
}

// hostToContainerCopyCmd generates the arguments to podman
// that copy a file from the container to host.
func containerToHostCopyCmd(hostPath, containerPath, containerId string) copySpec {
	return copySpec{containerPath, hostPath, true, false, containerId}
}
