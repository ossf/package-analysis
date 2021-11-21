package strace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// 510 06:34:52.506847   43512 strace.go:587] [   2] python3 E openat(AT_FDCWD /app, 0x7f13f2254c50 /root/.ssh, O_RDONLY|O_CLOEXEC|O_DIRECTORY|O_NONBLOCK, 0o0)
	stracePattern = regexp.MustCompile(`.*strace.go:\d+\] \[.*?\] (.+) (E|X) ([^\s]+)\((.*)\)`)
	// 0x7f1c3a0a2620 /usr/bin/uname, 0x7f1c39e12930 ["uname", "-rs"], 0x55bbefc2d070 ["HOSTNAME=63d5c9dbacb6", "PYTHON_PIP_VERSION=21.0.1", "HOME=/root"]
	execvePattern = regexp.MustCompile(`.*?(\[.*\])`)
	//0x7f13f201a0a3 /path, 0x0
	creatPattern = regexp.MustCompile(`[^\s]+ ([^,]+)`)
	//0x7f13f201a0a3 /proc/self/fd, O_RDONLY|O_CLOEXEC,
	openPattern = regexp.MustCompile(`[^\s]+ ([^,]+), ([^,]+)`)
	// AT_FDCWD /app, 0x7f13f201a0a3 /proc/self/fd, O_RDONLY|O_CLOEXEC, 0o0
	openatPattern = regexp.MustCompile(`[^\s]+ ([^,]+), [^\s]+ ([^,]+), ([^,]+)`)
	// 0x561c42f5be30 /usr/local/bin/Modules/Setup.local, 0x7fdfb323c180
	statPattern = regexp.MustCompile(`[^\s]+ ([^,]+),`)
	// 0x3 /tmp/pip-install-398qx_i7/build/bdist.linux-x86_64/wheel, 0x7ff1e4a30620 mal, 0x7fae4d8741f0, 0x100
	newfstatatPattern = regexp.MustCompile(`[^\s]+ ([^,]+), [^\s]+ ([^,]+)`)
	// 0x3 socket:[2], 0x7f1bc9e7b914 {Family: AF_INET, Addr: 8.8.8.8, Port: 53}, 0x10
	// 0x3 socket:[1], 0x7f27cbd0ac50 {Family: AF_INET, Addr: , Port: 0}, 0x10
	// 0x3 socket:[4], 0x55ed873bb510 {Family: AF_INET6, Addr: 2001:67c:1360:8001::24, Port: 80}, 0x1c
	// 0x3 socket:[16], 0x5568c5caf2d0 {Family: AF_INET, Addr: , Port: 5000}, 0x10
	socketPattern = regexp.MustCompile(`{Family: AF_INET6?, Addr: ([^,]*), Port: ([0-9]+)}`)
)

type FileInfo struct {
	Path  string
	Read  bool
	Write bool
}

type SocketInfo struct {
	Address string
	Port    int
}

type CommandInfo struct {
	Command []string
	Env     []string
}

type Result struct {
	files    map[string]*FileInfo
	sockets  map[string]*SocketInfo
	commands map[string]*CommandInfo
}

func parseOpenFlags(openFlags string) (read, write bool) {
	if strings.Contains(openFlags, "O_RDWR") {
		read = true
		write = true
	}

	if strings.Contains(openFlags, "O_CREAT") {
		write = true
	}

	if strings.Contains(openFlags, "O_WRONLY") {
		write = true
	}

	if strings.Contains(openFlags, "O_RDONLY") {
		read = true
	}
	return
}

func parsePort(portString string) (int, error) {
	return strconv.Atoi(portString)
}

func joinPaths(dir, file string) string {
	if filepath.IsAbs(file) {
		return file
	}

	return filepath.Join(dir, file)
}

func parseCmdAndEnv(cmdAndEnv string) ([]string, []string, error) {
	decoder := json.NewDecoder(strings.NewReader(cmdAndEnv))
	var cmd []string
	// Decode up to end of first valid JSON (which is the command).
	err := decoder.Decode(&cmd)
	if err != nil {
		return nil, nil, err
	}

	// Find the start of the next JSON (which is the environment).
	nextIdx := decoder.InputOffset() + int64(strings.Index(cmdAndEnv[decoder.InputOffset():], "["))
	decoder = json.NewDecoder(strings.NewReader(cmdAndEnv[nextIdx:]))
	var env []string
	err = decoder.Decode(&env)
	if err != nil {
		return nil, nil, err
	}

	return cmd, env, nil
}

func (r *Result) recordFileAccess(file string, read, write bool) {
	if _, exists := r.files[file]; !exists {
		r.files[file] = &FileInfo{Path: file}
	}
	r.files[file].Read = r.files[file].Read || read
	r.files[file].Write = r.files[file].Write || write
}

func (r *Result) recordSocket(address string, port int) {
	// Use a '-' dash as the address may contain colons if IPv6
	key := fmt.Sprintf("%s-%d", address, port)
	if _, exists := r.sockets[key]; !exists {
		r.sockets[key] = &SocketInfo{
			Address: address,
			Port:    port,
		}
	}
}

func (r *Result) recordCommand(cmd, env []string) {
	key := fmt.Sprintf("%s-%s", cmd, env)
	if _, exists := r.commands[key]; !exists {
		r.commands[key] = &CommandInfo{
			Command: cmd,
			Env:     env,
		}
	}
}

func (r *Result) parseSyscall(syscall, args string) error {
	switch syscall {
	case "creat":
		match := creatPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse create args: %s", args)
		}

		log.Printf("creat %s", match[1])
		r.recordFileAccess(match[1], false, true)
	case "open":
		match := openPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse open args: %s", args)
		}

		read, write := parseOpenFlags(match[2])
		log.Printf("open %s read=%t, write=%t", match[1], read, write)
		r.recordFileAccess(match[1], read, write)
	case "openat":
		match := openatPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse openat args: %s", args)
		}

		path := joinPaths(match[1], match[2])
		read, write := parseOpenFlags(match[3])
		log.Printf("openat %s read=%t, write=%t", path, read, write)
		r.recordFileAccess(path, read, write)
	case "execve":
		match := execvePattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse execve args: %s", args)
		}

		log.Printf("execve %s", match[1])
		cmd, env, err := parseCmdAndEnv(match[1])
		if err != nil {
			return err
		}
		r.recordCommand(cmd, env)
	case "bind":
		fallthrough
	case "listen":
		fallthrough
	case "connect":
		match := socketPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse socket args: %s", args)
		}
		address := match[1]
		port, err := parsePort(match[2])
		if err != nil {
			return err
		}
		log.Printf("socket %s : %d", address, port)
		r.recordSocket(address, port)
	case "fstat":
		fallthrough
	case "lstat":
		fallthrough
	case "stat":
		match := statPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse stat args: %s", args)
		}
		log.Printf("stat %s", match[1])
		r.recordFileAccess(match[1], true, false)
	case "newfstatat":
		match := newfstatatPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse newfstatat args: %s", args)
		}
		path := joinPaths(match[1], match[2])
		log.Printf("newfstatat %s", path)
		r.recordFileAccess(path, true, false)
	}
	return nil
}

// Parse reads an strace and collects the files, sockets and commands that were
// accessed.
func Parse(r io.Reader) (*Result, error) {
	result := &Result{
		files:    make(map[string]*FileInfo),
		sockets:  make(map[string]*SocketInfo),
		commands: make(map[string]*CommandInfo),
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		match := stracePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		if match[2] == "X" {
			// Analyze exit events only.
			err := result.parseSyscall(match[3], match[4])
			if err != nil {
				// Log errors and continue.
				log.Println(err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Files returns all the files access from the parsed strace.
func (r *Result) Files() []FileInfo {
	files := make([]FileInfo, 0, len(r.files))
	for _, file := range r.files {
		files = append(files, *file)
	}
	return files
}

// Sockets returns all the IPv4 and IPv6 sockets from the parsed strace.
func (r *Result) Sockets() []SocketInfo {
	sockets := make([]SocketInfo, 0, len(r.sockets))
	for _, socket := range r.sockets {
		sockets = append(sockets, *socket)
	}
	return sockets
}

// Commands returns all the exec'd commands from the parsed strace.
func (r *Result) Commands() []CommandInfo {
	commands := make([]CommandInfo, 0, len(r.commands))
	for _, command := range r.commands {
		commands = append(commands, *command)
	}
	return commands
}
