package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

type commandRecord struct {
	Command     string
	Environment string
}

type analysisInfo struct {
	Files    map[string]bool
	IPs      map[string]bool
	Commands map[commandRecord]bool
}

const (
	logPath = "/tmp/runsc.log.boot"
)

var (
	// 510 06:34:52.506847   43512 strace.go:587] [   2] python3 E openat(AT_FDCWD /app, 0x7f13f2254c50 /root/.ssh, O_RDONLY|O_CLOEXEC|O_DIRECTORY|O_NONBLOCK, 0o0)
	stracePattern = regexp.MustCompile(`.*strace.go:\d+\] \[.*?\] ([^\s]+) (E|X) ([^\s]+)\((.*)\)`)
	// 0x7f1c3a0a2620 /usr/bin/uname, 0x7f1c39e12930 ["uname", "-rs"], 0x55bbefc2d070 ["HOSTNAME=63d5c9dbacb6", "PYTHON_PIP_VERSION=21.0.1", "HOME=/root"]
	execvePattern = regexp.MustCompile(`.*?(\[.*?\])[^[]+(\[.*\])?`)
	// 0x7f1bc9c84e50 /usr/local/lib/python3.9/encodings/__pycache__/aliases.cpython-39.pyc,
	openPattern = regexp.MustCompile(`[^\s]+ ([^\s]+),`)
	// AT_FDCWD /app, 0x7f13f201a0a3 /proc/self/fd, O_RDONLY|O_CLOEXEC, 0o0
	openatPattern = regexp.MustCompile(`[^\s]+ ([^\s]+), [^\s]+ ([^\s]+),`)
	// 0x561c42f5be30 /usr/local/bin/Modules/Setup.local, 0x7fdfb323c180
	statPattern = regexp.MustCompile(`[^\s]+ ([^\s]+),`)
	// 0x3 socket:[2], 0x7f1bc9e7b914 {Family: AF_INET, Addr: 8.8.8.8, Port: 53}, 0x10
	connectPattern = regexp.MustCompile(`.*AF_INET.*Addr: ([^\s]+),`)
)

var (
	image   = flag.String("image", "", "image to use for analysis")
	command = flag.String("command", "", "command to use for analysis")
	bucket  = flag.String("bucket", "", "bucket for results")
	upload  = flag.String("upload", "", "path within bucket for results")
)

func main() {
	flag.Parse()
	if *image == "" || *command == "" {
		flag.Usage()
		return
	}

	info := runAnalysis(*image, *command)

	if *bucket != "" && *upload != "" {
		uploadResults(info)
	}
}

func analyzeSyscall(syscall, args string, info *analysisInfo) {
	switch syscall {
	case "open":
		match := openPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse open args: %s", args)
			return
		}

		log.Printf("open %s", match[1])
		info.Files[match[1]] = true
	case "openat":
		match := openatPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse openat args: %s", args)
			return
		}

		var path string
		if filepath.IsAbs(match[2]) {
			path = match[2]
		} else {
			path = filepath.Join(match[1], match[2])
		}
		log.Printf("openat %s", path)
		info.Files[path] = true
	case "execve":
		// TODO(ochang): Other exec syscalls?
		match := execvePattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse execve args: %s", args)
			return
		}

		cmd := commandRecord{
			Command:     match[1],
			Environment: match[2],
		}

		log.Printf("execve %s %s", match[1], match[2])
		info.Commands[cmd] = true
	case "connect":
		match := connectPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse connect args: %s", args)
			return
		}
		log.Printf("connect %s", match[1])
		info.IPs[match[1]] = true
	case "stat":
		fallthrough
	case "lstat":
		fallthrough
	case "fstat":
		match := statPattern.FindStringSubmatch(args)
		if match == nil {
			log.Printf("failed to parse stat args: %s", args)
			return
		}
		log.Printf("stat %s", match[1])
		info.Files[match[1]] = true
	}
}

func runAnalysis(image, command string) *analysisInfo {
	cmd := exec.Command("podman", "run", "--runtime=/usr/local/bin/runsc", "--cgroup-manager=cgroupfs", "--rm", image, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Panic(err)
	}

	file, err := os.Open(logPath)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	info := &analysisInfo{
		Files:    make(map[string]bool),
		IPs:      make(map[string]bool),
		Commands: make(map[commandRecord]bool),
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

	return info
}

type commandResult struct {
	Command     []string
	Environment []string
}

type data struct {
	Files    []string
	IPs      []string
	Commands []commandResult
}

func uploadResults(info *analysisInfo) {
	d := data{}
	for f, _ := range info.Files {
		d.Files = append(d.Files, f)
	}
	for ip, _ := range info.IPs {
		d.IPs = append(d.IPs, ip)
	}
	for command, _ := range info.Commands {
		result := commandResult{}
		if command.Command != "" {
			err := json.Unmarshal([]byte(command.Command), &result.Command)
			if err != nil {
				log.Panicf("Failed to parse %s: %v", command.Command, err)
			}
		}

		if command.Environment != "" {
			err := json.Unmarshal([]byte(command.Environment), &result.Environment)
			if err != nil {
				log.Panicf("Failed to parse %s: %v", command.Environment, err)
			}
		}

		d.Commands = append(d.Commands, result)
	}

	b, err := json.Marshal(d)
	if err != nil {
		log.Print(err)
		return
	}

	ctx := context.Background()
	bkt, err := blob.OpenBucket(ctx, *bucket)
	if err != nil {
		log.Panic(err)
	}
	defer bkt.Close()

	w, err := bkt.NewWriter(ctx, *upload, nil)
	if err != nil {
		log.Panic(err)
	}
	if _, err := w.Write(b); err != nil {
		log.Panic(err)
	}
	if err := w.Close(); err != nil {
		log.Panic(err)
	}
}
