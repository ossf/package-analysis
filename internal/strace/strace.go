package strace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/ossf/package-analysis/internal/log"
)

var (
	// 510 06:34:52.506847   43512 strace.go:587] [   2] python3 E openat(AT_FDCWD /app, 0x7f13f2254c50 /root/.ssh, O_RDONLY|O_CLOEXEC|O_DIRECTORY|O_NONBLOCK, 0o0)
	stracePattern = regexp.MustCompile(`.*strace.go:\d+\] \[.*?\] (.+) (E|X) (\S+)\((.*)\)`)
	// 0x7f1c3a0a2620 /usr/bin/uname, 0x7f1c39e12930 ["uname", "-rs"], 0x55bbefc2d070 ["HOSTNAME=63d5c9dbacb6", "PYTHON_PIP_VERSION=21.0.1", "HOME=/root"]
	execvePattern = regexp.MustCompile(`.*?(\[.*\])`)
	//0x7f13f201a0a3 /path, 0x0
	creatPattern = regexp.MustCompile(`\S+ ([^,]+)`)
	//0x7f13f201a0a3 /proc/self/fd, O_RDONLY|O_CLOEXEC,
	openPattern = regexp.MustCompile(`\S+ ([^,]+), ([^,]+)`)
	// AT_FDCWD /app, 0x7f13f201a0a3 /proc/self/fd, O_RDONLY|O_CLOEXEC, 0o0
	openatPattern = regexp.MustCompile(`\S+ ([^,]+), \S+ ([^,]+), ([^,]+)`)
	// 0x561c42f5be30 /usr/local/bin/Modules/Setup.local, 0x7fdfb323c180
	statPattern = regexp.MustCompile(`\S+ ([^,]+),`)
	// 0x3 /tmp/pip-install-398qx_i7/build/bdist.linux-x86_64/wheel, 0x7ff1e4a30620 mal, 0x7fae4d8741f0, 0x100
	newfstatatPattern = regexp.MustCompile(`\S+ ([^,]+), \S+ ([^,]+)`)
	// 0x3 socket:[2], 0x7f1bc9e7b914 {Family: AF_INET, Addr: 8.8.8.8, Port: 53}, 0x10
	// 0x3 socket:[1], 0x7f27cbd0ac50 {Family: AF_INET, Addr: , Port: 0}, 0x10
	// 0x3 socket:[4], 0x55ed873bb510 {Family: AF_INET6, Addr: 2001:67c:1360:8001::24, Port: 80}, 0x1c
	// 0x3 socket:[16], 0x5568c5caf2d0 {Family: AF_INET, Addr: , Port: 5000}, 0x10
	socketPattern = regexp.MustCompile(`{Family: ([^,]+), (Addr: ([^,]*), Port: ([0-9]+)|[^}]+)}`)

	// 0x7fe003272980 /tmp/jpu6po61
	unlinkPatten = regexp.MustCompile(`0x[a-f\d]+ ([^)]+)`)

	// unlinkat(0x4 /tmp/pip-pip-egg-info-ng4_5gp_/temps.egg-info, 0x7fe0031c9a10 top_level.txt, 0x0)
	// unlinkat(AT_FDCWD /app, 0x5569a7e83380 /app/vendor/composer/e06632ca, 0x200)
	unlinkatPattern = regexp.MustCompile(`\S+ ([^,]+), 0x[a-f\d]+ ([^,]+), 0x[a-f\d]+`)
)

type FileInfo struct {
	Path   string
	Read   bool
	Write  bool
	Delete bool
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

func (r *Result) recordFileAccess(file string, read, write, delete bool) {
	if _, exists := r.files[file]; !exists {
		r.files[file] = &FileInfo{Path: file}
	}
	r.files[file].Read = r.files[file].Read || read
	r.files[file].Write = r.files[file].Write || write
	r.files[file].Delete = r.files[file].Delete || delete
}

func (r *Result) recordSocket(address string, port int) {
	// Use a '-' dash as the address may contain colons if IPv6
	// Pad the integer field so that keys can be sorted.
	key := fmt.Sprintf("%s-%05d", address, port)
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

		log.Debug("creat",
			"path", match[1])
		r.recordFileAccess(match[1], false, true, false)
	case "open":
		match := openPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse open args: %s", args)
		}

		read, write := parseOpenFlags(match[2])
		log.Debug("open",
			"path", match[1],
			"read", read,
			"write", write)
		r.recordFileAccess(match[1], read, write, false)
	case "openat":
		match := openatPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse openat args: %s", args)
		}

		path := joinPaths(match[1], match[2])
		read, write := parseOpenFlags(match[3])
		log.Debug("openat",
			"path", path,
			"read", read,
			"write", write)
		r.recordFileAccess(path, read, write, false)
	case "execve":
		match := execvePattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse execve args: %s", args)
		}
		log.Debug("execve",
			"cmdAndEnv", match[1])
		cmd, env, err := parseCmdAndEnv(match[1])
		if err != nil {
			return err
		}
		r.recordCommand(cmd, env)
	case "bind":
		fallthrough
	case "connect":
		match := socketPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse socket args: %s", args)
		}
		family := match[1]
		if family != "AF_INET" && family != "AF_INET6" {
			log.Debug("Ignoring socket",
				"family", family,
				"socket", match[2])
			return nil
		}
		address := match[3]
		port, err := parsePort(match[4])
		if err != nil {
			return err
		}
		log.Debug("socket",
			"address", address,
			"port", port)
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
		log.Debug("stat",
			"path", match[1])
		r.recordFileAccess(match[1], true, false, false)
	case "newfstatat":
		match := newfstatatPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse newfstatat args: %s", args)
		}
		path := joinPaths(match[1], match[2])
		log.Debug("newfstatat",
			"path", path)
		r.recordFileAccess(path, true, false, false)
	case "unlink":
		match := unlinkPatten.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse unlink args: %s", args)
		}
		path := match[1]
		log.Debug("unlink",
			"path", path)
		r.recordFileAccess(path, false, false, true)
	case "unlinkat":
		match := unlinkatPattern.FindStringSubmatch(args)
		if match == nil {
			return fmt.Errorf("Failed to parse unlinkat args: %s", args)
		}
		path := joinPaths(match[1], match[2])
		log.Debug("unlinkat",
			"path", path)
		r.recordFileAccess(path, false, false, true)
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

	// Use a buffered reader, rather than scanner, to allow for lines with
	// unlimited length.
	bufR := bufio.NewReader(r)
	for {
		line, err := bufR.ReadString('\n')
		// Trim any trailing space
		line = strings.TrimRightFunc(line, unicode.IsSpace)

		match := stracePattern.FindStringSubmatch(line)
		if match != nil && match[2] == "X" {
			// Analyze exit events only.
			err := result.parseSyscall(match[3], match[4])
			if err != nil {
				// Log errors and continue.
				log.Warn("Failed to parse syscall",
					"error", err)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Files returns all the files access from the parsed strace.
func (r *Result) Files() []FileInfo {
	// Sort the keys so the output is in a stable order
	paths := make([]string, 0, len(r.files))
	for p := range r.files {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	files := make([]FileInfo, 0, len(paths))
	for _, p := range paths {
		files = append(files, *r.files[p])
	}
	return files
}

// Sockets returns all the IPv4 and IPv6 sockets from the parsed strace.
func (r *Result) Sockets() []SocketInfo {
	// Sort the keys so the output is in a stable order
	keys := make([]string, 0, len(r.sockets))
	for k := range r.sockets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sockets := make([]SocketInfo, 0, len(keys))
	for _, k := range keys {
		sockets = append(sockets, *r.sockets[k])
	}
	return sockets
}

// Commands returns all the exec'd commands from the parsed strace.
func (r *Result) Commands() []CommandInfo {
	// Sort the keys so the output is in a stable order
	keys := make([]string, 0, len(r.commands))
	for k := range r.commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	commands := make([]CommandInfo, 0, len(keys))
	for _, k := range keys {
		commands = append(commands, *r.commands[k])
	}
	return commands
}
