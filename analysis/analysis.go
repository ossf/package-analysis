package analysis

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"

	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/gcpfirestore"
)

type fileInfo struct {
	Read  bool
	Write bool
}

type analysisInfo struct {
	Files    map[string]*fileInfo
	IPs      map[string]bool
	Commands map[string]bool
}

type PkgManager struct {
	CommandFmt func(string, string) string
	GetLatest  func(string) string
	Image      string
}

type commandResult struct {
	Command     []string
	Environment []string
}

type fileResult struct {
	Path  string
	Read  bool
	Write bool
}

type Package struct {
	Ecosystem string
	Name      string
	Version   string
}

type AnalysisResult struct {
	Package  Package
	Files    []fileResult
	IPs      []string
	Commands []commandResult
}

type DocstoreData struct {
	ID       string
	Package  Package
	Files    []string
	IPs      []string
	Commands []string
}

const (
	logPath = "/tmp/runsc.log.boot"
)

var (
	SupportedPkgManagers = map[string]PkgManager{
		"npm":      NPMPackageManager,
		"pypi":     PyPIPackageManager,
		"rubygems": RubyGemsPackageManager,
	}
)

var (
	// 510 06:34:52.506847   43512 strace.go:587] [   2] python3 E openat(AT_FDCWD /app, 0x7f13f2254c50 /root/.ssh, O_RDONLY|O_CLOEXEC|O_DIRECTORY|O_NONBLOCK, 0o0)
	stracePattern = regexp.MustCompile(`.*strace.go:\d+\] \[.*?\] ([^\s]+) (E|X) ([^\s]+)\((.*)\)`)
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
	connectPattern = regexp.MustCompile(`.*AF_INET.*Addr: ([^,]+),`)
)

func recordFileAccess(info *analysisInfo, file string, read, write bool) {
	if _, exists := info.Files[file]; !exists {
		info.Files[file] = &fileInfo{}
	}
	info.Files[file].Read = info.Files[file].Read || read
	info.Files[file].Write = info.Files[file].Write || write
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

func extractCmdAndEnv(cmdAndEnv string) ([]string, []string) {
	decoder := json.NewDecoder(strings.NewReader(cmdAndEnv))
	var cmd []string
	// Decode up to end of first valid JSON (which is the command).
	err := decoder.Decode(&cmd)
	if err != nil {
		log.Panicf("failed to parse %s: %v", cmdAndEnv, err)
	}

	// Find the start of the next JSON (which is the environment).
	nextIdx := decoder.InputOffset() + int64(strings.Index(cmdAndEnv[decoder.InputOffset():], "["))
	decoder = json.NewDecoder(strings.NewReader(cmdAndEnv[nextIdx:]))
	var env []string
	err = decoder.Decode(&env)
	if err != nil {
		log.Panicf("failed to parse %s: %v", cmdAndEnv[nextIdx:], err)
	}

	return cmd, env
}

func joinPaths(dir, file string) string {
	if filepath.IsAbs(file) {
		return file
	}

	return filepath.Join(dir, file)
}

func analyzeSyscall(syscall, args string, info *analysisInfo) {
	switch syscall {
	case "creat":
		match := creatPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse creat args: %s", args)
			return
		}

		log.Printf("creat %s", match[1])
		recordFileAccess(info, match[1], false, true)
	case "open":
		match := openPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse open args: %s", args)
			return
		}

		read, write := parseOpenFlags(match[2])
		log.Printf("open %s read=%t, write=%t", match[1], read, write)
		recordFileAccess(info, match[1], read, write)
	case "openat":
		match := openatPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse openat args: %s", args)
			return
		}

		path := joinPaths(match[1], match[2])
		read, write := parseOpenFlags(match[3])
		log.Printf("openat %s read=%t, write=%t", path, read, write)
		recordFileAccess(info, path, read, write)
	case "execve":
		match := execvePattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse execve args: %s", args)
			return
		}

		log.Printf("execve %s", match[1])
		info.Commands[match[1]] = true
	case "connect":
		match := connectPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse connect args: %s", args)
			return
		}
		log.Printf("connect %s", match[1])
		info.IPs[match[1]] = true
	case "fstat":
		fallthrough
	case "lstat":
		fallthrough
	case "stat":
		match := statPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse stat args: %s", args)
			return
		}
		log.Printf("stat %s", match[1])
		recordFileAccess(info, match[1], true, false)
	case "newfstatat":
		match := newfstatatPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse newfstatat args: %s", args)
			return
		}
		path := joinPaths(match[1], match[2])
		log.Printf("newfstatat %s", path)
		recordFileAccess(info, path, true, false)
	}
}

func Run(ecosystem, pkgName, version, image, command string) *AnalysisResult {
	log.Printf("Running analysis using %s: %s", image, command)

	cmd := exec.Command(
		"podman", "run", "--runtime=/usr/local/bin/runsc", "--cgroup-manager=cgroupfs",
		"--events-backend=file", "--rm", image, "sh", "-c", command)
	cmd.Stdout = os.Stdout

	pipe, err := cmd.StderrPipe()
	if err != nil {
		log.Panic(err)
	}

	if err := cmd.Start(); err != nil {
		log.Panic(err)
	}
	stderr, err := io.ReadAll(pipe)
	if err != nil {
		log.Panic(err)
	}

	if err := cmd.Wait(); err != nil {
		// Not really an error
		if !strings.Contains(string(stderr), "gofer is still running") {
			log.Panic(err)
		}
	}

	file, err := os.Open(logPath)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	info := &analysisInfo{
		Files:    make(map[string]*fileInfo),
		IPs:      make(map[string]bool),
		Commands: make(map[string]bool),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := stracePattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		if match[2] == "X" {
			// Analyze exit events only.
			analyzeSyscall(match[3], match[4], info)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Panic(err)
	}

	result := AnalysisResult{}
	result.setData(ecosystem, pkgName, version, info)
	return &result
}

func (d *AnalysisResult) setData(ecosystem, pkgName, version string, info *analysisInfo) {
	d.Package.Ecosystem = ecosystem
	d.Package.Name = pkgName
	d.Package.Version = version

	for f, info := range info.Files {
		d.Files = append(d.Files, fileResult{
			Path:  f,
			Read:  info.Read,
			Write: info.Write,
		})
	}
	for ip, _ := range info.IPs {
		d.IPs = append(d.IPs, ip)
	}
	for command, _ := range info.Commands {
		cmd, env := extractCmdAndEnv(command)
		result := commandResult{
			Command:     cmd,
			Environment: env,
		}
		d.Commands = append(d.Commands, result)
	}

}

func normalizeDocstoreName(ID string) string {
	return strings.ReplaceAll(ID, "/", "\\")
}

func (d *DocstoreData) setData(result *AnalysisResult) {
	d.ID = normalizeDocstoreName(
		fmt.Sprintf("%s:%s:%s", result.Package.Ecosystem, result.Package.Name, result.Package.Version))
	d.Package = result.Package

	// Index touched file components.
	fileParts := map[string]bool{}
	for _, f := range result.Files {
		cur := f.Path
		for cur != "/" && cur != "." {
			name := filepath.Base(cur)
			fileParts[name] = true
			cur = filepath.Dir(cur)
		}
	}
	for part, _ := range fileParts {
		d.Files = append(d.Files, part)
	}

	// IPs are indexed as is.
	for _, ip := range result.IPs {
		d.IPs = append(d.IPs, ip)
	}

	// Index command components.
	cmdParts := map[string]bool{}
	for _, cmd := range result.Commands {
		for _, part := range cmd.Command {
			cmdParts[part] = true
		}
	}
	for part, _ := range cmdParts {
		d.Commands = append(d.Commands, part)
	}
}

func UploadResults(ctx context.Context, bucket, path string, result *AnalysisResult) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	bkt, err := blob.OpenBucket(ctx, bucket)
	if err != nil {
		return err
	}
	defer bkt.Close()

	uploadPath := filepath.Join(path, result.Package.Version+".json")
	log.Printf("uploading to bucket=%s, path=%s", bucket, uploadPath)

	w, err := bkt.NewWriter(ctx, uploadPath, nil)
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func WriteResultsToDocstore(ctx context.Context, collectionPath string, result *AnalysisResult) error {
	coll, err := docstore.OpenCollection(ctx, collectionPath)
	if err != nil {
		return err
	}
	defer coll.Close()

	d := DocstoreData{}
	d.setData(result)

	err = coll.Put(ctx, &d)
	if err != nil {
		return err
	}
	return nil
}
