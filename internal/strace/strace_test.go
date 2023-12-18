package strace_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/ossf/package-analysis/internal/strace"
	"github.com/ossf/package-analysis/internal/utils"
)

var nopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestIgnoreEntryLogs(t *testing.T) {
	input := "I1203 05:29:21.585712     173 strace.go:625] [   2] python3 E creat(0x7f015d7865d0 /tmp/abctest, 0o600)"
	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}
	if l := len(res.Files()); l != 0 {
		t.Errorf(`len(r.Files()) = %d, want 0`, l)
	}
}

func TestParseFileReadThenCreate(t *testing.T) {
	input := "I1203 00:02:39.681902     171 strace.go:625] [   1] ruby X openat(AT_FDCWD /app, 0x55c5319654f0 /app/foobar, O_RDONLY|O_CLOEXEC|O_NONBLOCK, 0o0) = 0x0 errno=2 (no such file or directory) (11.709µs)\n" +
		"I1203 00:02:38.316076     171 strace.go:587] [   2] gem X openat(AT_FDCWD /app, 0x7f3336aaf2c8 /app/foobar, O_CLOEXEC|O_CREAT|O_TRUNC, 0o666)"
	want := strace.FileInfo{
		Path:  "/app/foobar",
		Read:  true,
		Write: true,
	}

	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}
	files := res.Files()
	if len(files) != 1 || !reflect.DeepEqual(files[0], want) {
		t.Errorf(`Files() = %v, want [%v]`, files, want)
	}
}

func TestParseFileWriteMultipleWritesToSameFile(t *testing.T) {
	input := "I0928 00:18:54.794008     365 strace.go:593] [   6:   6] uname E write(0x1 host:[5], 0x555695ceaab0 \"Linux 4.4.0\\n\", 0xc)\n" +
		"I1109 06:53:19.688807     950 strace.go:593] [   3:   3] python3 E write(0x1 host:[5], 0x560d2708e7c0 \"django.template.base\\nImporting django.template.context\\nImporting django.template.context_processors\\nImporting django.template.defaultfilters\\nImporting django.template.defaulttags\\nImporting django.template.engine\\nImporting django.template.exceptions\\nImporting django.template.library\\nImporting django.template.loader\\nImporting django.template.loader_tags\\nImporting django.template.loaders\\nImporting django.template.loaders.app_directories\\nImporting django.template.loaders.base\\nImporting django.template.loaders.cached\\nImporting django.template.loaders.filesystem\\nImporting django.template.loaders.locmem\\nImporting django.template.response\\nImporting django.template.smartif\\nImporting django.template.utils\\nImporting django.templatetags\\nImporting django.templatetags.cache\\nImporting django.templatetags.i18n\\nImporting django.templatetags.l10n\\nImporting django.templatetags.static\\nImporting django.templatetags.tz\\nImporting django.test\\nImporting django.test.client\\nImporting django.test.html\\nImporting django.test.runner\\nImport\"..., 0xe64)"
	firstWriteInfoWant := strace.WriteContentInfo{
		BytesWritten:  12,
		WriteBufferId: "181080a0b2dce592f16ab55aacb18c7a4cb849c9a7f644c5c76edf56e4870ebd",
	}
	secondWriteInfoWant := strace.WriteContentInfo{
		BytesWritten:  3684,
		WriteBufferId: "7b2c403bc9eee677758c2a575b2f8602b37f880f4fdb98ee98c30a74a5b9a52b",
	}
	writeInfoWantArray := strace.WriteInfo{firstWriteInfoWant, secondWriteInfoWant}

	want := strace.FileInfo{
		Path:      "host:[5]",
		Write:     true,
		WriteInfo: writeInfoWantArray}

	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}

	files := res.Files()
	if len(files) != 1 || !reflect.DeepEqual(files[0], want) {
		t.Errorf(`Files() = %v, want %v`, files[0], want)
	}

	utils.RemoveTempFilesDirectory()
}

func TestParseFileWritesToDifferentFiles(t *testing.T) {
	input := "I0928 00:18:54.794008     365 strace.go:593] [   6:   6] uname E write(0x1 host:[5], 0x555695ceaab0 \"Linux 4.4.0\\n\", 0xc)\n" +
		"I1109 06:53:19.688807     950 strace.go:593] [   3:   3] python3 E write(0x1 pipe:[5], 0x560d2708e7c0 \"django.template.base\\nImporting django.template.context\\nImporting django.template.context_processors\\nImporting django.template.defaultfilters\\nImporting django.template.defaulttags\\nImporting django.template.engine\\nImporting django.template.exceptions\\nImporting django.template.library\\nImporting django.template.loader\\nImporting django.template.loader_tags\\nImporting django.template.loaders\\nImporting django.template.loaders.app_directories\\nImporting django.template.loaders.base\\nImporting django.template.loaders.cached\\nImporting django.template.loaders.filesystem\\nImporting django.template.loaders.locmem\\nImporting django.template.response\\nImporting django.template.smartif\\nImporting django.template.utils\\nImporting django.templatetags\\nImporting django.templatetags.cache\\nImporting django.templatetags.i18n\\nImporting django.templatetags.l10n\\nImporting django.templatetags.static\\nImporting django.templatetags.tz\\nImporting django.test\\nImporting django.test.client\\nImporting django.test.html\\nImporting django.test.runner\\nImport\"..., 0xe64)"
	firstFileWant := strace.FileInfo{
		Path:  "host:[5]",
		Write: true,
		WriteInfo: strace.WriteInfo{
			{
				BytesWritten:  12,
				WriteBufferId: "181080a0b2dce592f16ab55aacb18c7a4cb849c9a7f644c5c76edf56e4870ebd",
			},
		},
	}

	secondFileWant := strace.FileInfo{
		Path:  "pipe:[5]",
		Write: true,
		WriteInfo: strace.WriteInfo{
			{
				BytesWritten:  3684,
				WriteBufferId: "7b2c403bc9eee677758c2a575b2f8602b37f880f4fdb98ee98c30a74a5b9a52b",
			},
		},
	}

	want := []strace.FileInfo{firstFileWant, secondFileWant}

	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}
	files := res.Files()
	if !reflect.DeepEqual(files, want) {
		t.Errorf(`Files() = %v, want  = %v`, files, want)
	}
	utils.RemoveTempFilesDirectory()
}

func TestParseFileWriteWithZeroBytesWritten(t *testing.T) {
	// Write calls where zero bytes are written are formatted as below where the write buffer argument does not include quotes.
	input := "I1202 06:13:07.127115     312 strace.go:593] [  10:  10] npm init E write(0x2 host:[6], , 0x0)"

	want := strace.FileInfo{
		Path:  "host:[6]",
		Write: true,
		WriteInfo: strace.WriteInfo{
			{
				BytesWritten: 0,
				// sha256 of empty string.
				WriteBufferId: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
		},
	}

	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}
	files := res.Files()
	if len(files) != 1 || !reflect.DeepEqual(files[0], want) {
		t.Errorf(`Files() = %v, want [%v]`, files, want)
	}
	utils.RemoveTempFilesDirectory()
}

func TestUnlinkOneEntry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  strace.FileInfo
	}{
		{
			// unlink with path
			name:  "unlink",
			input: "I0902 01:19:17.729518     303 strace.go:625] [   4:   4] python3 X unlink(0x7ff5f78e4980 /tmp/lbosrzlp) = 0 (0x0) (58.552Âµs)",
			want: strace.FileInfo{
				Path:   "/tmp/lbosrzlp",
				Delete: true,
			},
		},
		{
			// unlink with no path and error
			name:  "unlink2",
			input: "I1116 06:22:52.164421    1158 strace.go:625] [  23:  23] cmake X unlink(0x7f234e5bd500 ) = 0 (0x0) errno=2 (no such file or directory) (667ns)",
			want: strace.FileInfo{
				Path:   "",
				Delete: true,
			},
		},
		{
			name:  "unlinkat",
			input: "I0902 01:19:18.991729     303 strace.go:631] [   4:   4] python3 X unlinkat(0x3 /tmp/pip-unpack-7xfj8327, 0x7ff5f790c410 temps-0.3.0.tar.gz, 0x0) = 0 (0x0) (39.914Âµs)",
			want: strace.FileInfo{
				Path:   "/tmp/pip-unpack-7xfj8327/temps-0.3.0.tar.gz",
				Delete: true,
			},
		},
		{
			name:  "unlinkat_2",
			input: "I0907 23:56:32.113900     302 strace.go:631] [  48:  48] rm X unlinkat(AT_FDCWD /app, 0x5569a7e83380 /app/vendor/composer/e06632ca, 0x200) = 0 (0x0) (69.951µs)",
			want: strace.FileInfo{
				Path:   "/app/vendor/composer/e06632ca",
				Delete: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.input)
			res, err := strace.Parse(context.Background(), r, nopLogger)
			if err != nil || res == nil {
				t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
			}
			files := res.Files()
			if len(files) != 1 || !reflect.DeepEqual(files[0], test.want) {
				t.Errorf(`Files() = %v, want [%v]`, files, test.want)
			}
		})
	}

}

func TestParseFilesOneEntry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  strace.FileInfo
	}{
		{
			name:  "creat",
			input: "I1203 05:29:21.585712     173 strace.go:625] [   2] python3 X creat(0x7f015d7865d0 /tmp/abctest, 0o600) = 0x6 (598.693µs)",
			want: strace.FileInfo{
				Path:  "/tmp/abctest",
				Read:  false,
				Write: true,
			},
		},
		{
			name:  "open_rdwr",
			input: "I1203 05:29:21.585712     173 strace.go:625] [   2] python3 X open(0x7f015d7865d0 /root/.cache/pip/selfcheck/fe300af6f7d708c14827daac3afc81fbb8306b73de8dd6e3f1f8ea3bb56zgzfi.tmp, O_RDWR|O_CLOEXEC|O_EXCL|O_NOFOLLOW, 0o600) = 0x6 (598.693µs)",
			want: strace.FileInfo{
				Path:  "/root/.cache/pip/selfcheck/fe300af6f7d708c14827daac3afc81fbb8306b73de8dd6e3f1f8ea3bb56zgzfi.tmp",
				Read:  true,
				Write: true,
			},
		},
		{
			name:  "open_rdonly",
			input: "I1203 00:02:39.681902     171 strace.go:625] [   1] ruby X open(0x55c5319654f0 /usr/local/lib/ruby/vendor_ruby/3.0.0/digest.so, O_RDONLY|O_CLOEXEC|O_NONBLOCK, 0o0) = 0x0 errno=2 (no such file or directory) (11.709µs)",
			want: strace.FileInfo{
				Path:  "/usr/local/lib/ruby/vendor_ruby/3.0.0/digest.so",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "open_wronly",
			input: "I1203 00:02:38.316076     171 strace.go:587] [   2] gem X open(0x7f3336aaf2c8 /dev/null, O_CLOEXEC|O_WRONLY|O_TRUNC, 0o666)",
			want: strace.FileInfo{
				Path:  "/dev/null",
				Read:  false,
				Write: true,
			},
		},
		{
			name:  "open_create",
			input: "I1203 00:02:38.316076     171 strace.go:587] [   2] gem X open(0x7f3336aaf2c8 /dev/null, O_CLOEXEC|O_CREAT|O_TRUNC, 0o666)",
			want: strace.FileInfo{
				Path:  "/dev/null",
				Read:  false,
				Write: true,
			},
		},
		{
			name:  "openat_rdwr",
			input: "I1203 05:29:21.585712     173 strace.go:625] [   2] python3 X openat(AT_FDCWD /app, 0x7f015d7865d0 /root/.cache/pip/selfcheck/fe300af6f7d708c14827daac3afc81fbb8306b73de8dd6e3f1f8ea3bb56zgzfi.tmp, O_RDWR|O_CLOEXEC|O_EXCL|O_NOFOLLOW, 0o600) = 0x6 (598.693µs)",
			want: strace.FileInfo{
				Path:  "/root/.cache/pip/selfcheck/fe300af6f7d708c14827daac3afc81fbb8306b73de8dd6e3f1f8ea3bb56zgzfi.tmp",
				Read:  true,
				Write: true,
			},
		},
		{
			name:  "openat_rdonly",
			input: "I1203 00:02:39.681902     171 strace.go:625] [   1] ruby X openat(AT_FDCWD /app, 0x55c5319654f0 /usr/local/lib/ruby/vendor_ruby/3.0.0/digest.so, O_RDONLY|O_CLOEXEC|O_NONBLOCK, 0o0) = 0x0 errno=2 (no such file or directory) (11.709µs)",
			want: strace.FileInfo{
				Path:  "/usr/local/lib/ruby/vendor_ruby/3.0.0/digest.so",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "openat_wronly",
			input: "I1203 00:02:38.316076     171 strace.go:587] [   2] gem X openat(AT_FDCWD /app, 0x7f3336aaf2c8 /dev/null, O_CLOEXEC|O_WRONLY|O_TRUNC, 0o666)",
			want: strace.FileInfo{
				Path:  "/dev/null",
				Read:  false,
				Write: true,
			},
		},
		{
			name:  "openat_creat",
			input: "I1203 00:02:38.316076     171 strace.go:587] [   2] gem X openat(AT_FDCWD /app, 0x7f3336aaf2c8 /dev/null, O_CLOEXEC|O_CREAT|O_TRUNC, 0o666)",
			want: strace.FileInfo{
				Path:  "/dev/null",
				Read:  false,
				Write: true,
			},
		},
		{
			name:  "openat_relative_path",
			input: "I1205 23:19:13.505292     172 strace.go:625] [  18] npm X openat(AT_FDCWD /app, 0x4b626d0 .git/config, O_RDONLY|O_CLOEXEC, 0o0) = 0x0 errno=2 (no such file or directory) (104.863µs)",
			want: strace.FileInfo{
				Path:  "/app/.git/config",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "fstat",
			input: "I1203 05:30:11.960582     173 strace.go:619] [   1] python3 X fstat(0x3 /usr/local/lib/python3.9/codeop.py, 0x7fa2ba4ba780 {dev=11, ino=66, mode=S_IFREG|0o644, nlink=1, uid=0, gid=0, rdev=0, size=6326, blksize=4096, blocks=12, atime=2021-05-04 18:26:00 +0000 UTC, mtime=2021-05-04 18:26:00 +0000 UTC, ctime=2021-12-02 02:59:30.078976068 +0000 UTC}) = 0x0 (5.233µs)",
			want: strace.FileInfo{
				Path:  "/usr/local/lib/python3.9/codeop.py",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "lstat",
			input: "I1203 05:28:25.561795     173 strace.go:619] [   1] python3 X lstat(0x7fa2ba4adb50 /usr, 0x7fa2ba4ada60 {dev=11, ino=18, mode=S_IFDIR|0o755, nlink=10, uid=0, gid=0, rdev=0, size=4096, blksize=4096, blocks=8, atime=2021-12-02 02:59:30.654969556 +0000 UTC, mtime=2021-04-08 00:00:00 +0000 UTC, ctime=2021-12-02 02:59:30.634969784 +0000 UTC}) = 0x0 (5.924µs)",
			want: strace.FileInfo{
				Path:  "/usr",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "stat",
			input: "I1203 05:28:25.273429     173 strace.go:619] [   1] python3 X stat(0x55714f3be5c0 /usr/local/sbin/python3, 0x7fa2ba4be460) = 0x0 errno=2 (no such file or directory) (18.061µs)",
			want: strace.FileInfo{
				Path:  "/usr/local/sbin/python3",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "newfstatat",
			input: "I0722 17:06:36.466808     616 strace.go:632] [   6] isolate X newfstatat(AT_FDCWD /, 0xc0000ac180 /envs/test, 0xc00015a928 {dev=11, ino=37, mode=S_IFDIR|0o550, nlink=3, uid=0, gid=1001, rdev=0, size=20, blksize=4096, blocks=0, atime=2021-07-20 17:35:20.259535202 +0000 UTC, mtime=2021-07-20 17:20:52.529806118 +0000 UTC, ctime=2021-07-20 17:35:04.831007038 +0000 UTC}, 0x0) = 0x0 (596.593µs)",
			want: strace.FileInfo{
				Path:  "/envs/test",
				Read:  true,
				Write: false,
			},
		},
		{
			name:  "newfstatat_relative",
			input: "I0722 17:06:36.466808     616 strace.go:632] [   6] isolate X newfstatat(AT_FDCWD /envs, 0xc0000ac180 test, 0xc00015a928 {dev=11, ino=37, mode=S_IFDIR|0o550, nlink=3, uid=0, gid=1001, rdev=0, size=20, blksize=4096, blocks=0, atime=2021-07-20 17:35:20.259535202 +0000 UTC, mtime=2021-07-20 17:20:52.529806118 +0000 UTC, ctime=2021-07-20 17:35:04.831007038 +0000 UTC}, 0x0) = 0x0 (596.593µs)",
			want: strace.FileInfo{
				Path:  "/envs/test",
				Read:  true,
				Write: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.input)
			res, err := strace.Parse(context.Background(), r, nopLogger)
			if err != nil || res == nil {
				t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
			}
			files := res.Files()
			if len(files) != 1 || !reflect.DeepEqual(files[0], test.want) {
				t.Errorf(`Files() = %v, want [%v]`, files, test.want)
			}
		})
	}
}

func TestParseIgnoredSockets(t *testing.T) {
	input := "I1206 02:02:36.966250     205 strace.go:622] [   2] gem X connect(0x5 socket:[2], 0x7f414ed92ba0 {Family: AF_UNIX, Addr: \"/var/run/nscd/socket\"}, 0x6e) = 0x0 errno=2 (no such file or directory) (364.345µs)\n" +
		"I1206 02:02:36.989375     205 strace.go:622] [   2] gem X bind(0x5 socket:[1], 0x7f414ed92cf8 {Family: AF_NETLINK, PortID: 0, Groups: 0}, 0xc) = 0x0 (16.276µs)\n" +
		"I1206 02:02:36.990646     205 strace.go:622] [   2] gem X connect(0x5 socket:[2], 0x7f414ed93080 {Family: AF_UNSPEC, family addr format unknown}, 0x10) = 0x0 (8.598µs)\n"
	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}
	sockets := res.Sockets()
	if got := len(sockets); got != 0 {
		t.Errorf(`len(Sockets()) = %d, want 0`, got)
	}
}

func TestParseSocketsOneEntry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  strace.SocketInfo
	}{
		{
			name:  "bind_ipv4_web",
			input: "I1206 00:04:38.644850     175 strace.go:622] [  15] nc X bind(0x12 socket:[1], 0x7faa3cc00dcc {Family: AF_INET, Addr: 127.0.0.1, Port: 8080}, 0x10) = 0x0 (94.161µs)",
			want: strace.SocketInfo{
				Address: "127.0.0.1",
				Port:    8080,
			},
		},
		{
			name:  "bind_ipv6_web",
			input: "I1206 01:06:29.430943     203 strace.go:622] [   2] nc X bind(0x4 socket:[8], 0x560348812700 {Family: AF_INET6, Addr: ::1, Port: 8888}, 0x1c) = 0x0 errno=113 (no route to host) (4.817µs)",
			want: strace.SocketInfo{
				Address: "::1",
				Port:    8888,
			},
		},
		{
			name:  "bind_noaddr_ipv4",
			input: "I1206 01:51:11.594502     204 strace.go:584] [ 278] nc X bind(0x3 socket:[17], 0x55b3821492d0 {Family: AF_INET, Addr: , Port: 5555}, 0x10)",
			want: strace.SocketInfo{
				Address: "",
				Port:    5555,
			},
		},
		{
			name:  "bind_noaddr_ipv6",
			input: "I1206 01:53:22.858785     204 strace.go:622] [ 279] nc X bind(0x3 socket:[18], 0x55d6ca0682d0 {Family: AF_INET6, Addr: , Port: 8080}, 0x1c) = 0x0 (15.285µs)",
			want: strace.SocketInfo{
				Address: "",
				Port:    8080,
			},
		},
		{
			name:  "connect_ipv4_https",
			input: "I1206 00:04:41.714862     175 strace.go:622] [  19] npm install @go X connect(0x1d socket:[57], 0x7f34c41402d0 {Family: AF_INET, Addr: 104.16.19.35, Port: 443}, 0x10) = 0x0 errno=115 (operation now in progress) (130.736µs)",
			want: strace.SocketInfo{
				Address: "104.16.19.35",
				Port:    443,
			},
		},
		{
			name:  "connect_ipv4_dns",
			input: "I1206 00:04:38.644850     175 strace.go:622] [  15] npm X connect(0x12 socket:[1], 0x7faa3cc00dcc {Family: AF_INET, Addr: 8.8.8.8, Port: 53}, 0x10) = 0x0 (94.161µs)",
			want: strace.SocketInfo{
				Address: "8.8.8.8",
				Port:    53,
			},
		},
		{
			name:  "connect_ipv6_https",
			input: "I1206 01:06:29.430943     203 strace.go:622] [   2] python3 X connect(0x4 socket:[8], 0x560348812700 {Family: AF_INET6, Addr: 2a04:4e42:400::319, Port: 443}, 0x1c) = 0x0 errno=113 (no route to host) (4.817µs)",
			want: strace.SocketInfo{
				Address: "2a04:4e42:400::319",
				Port:    443,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.input)
			res, err := strace.Parse(context.Background(), r, nopLogger)
			if err != nil || res == nil {
				t.Errorf(`Parse(r) = %v, %v, want _, nil`, res, err)
			}
			sockets := res.Sockets()
			if len(sockets) != 1 || sockets[0] != test.want {
				t.Errorf(`Sockets() = %v, want [%v]`, sockets, test.want)
			}
		})
	}
}

func TestReallyLongLogLine(t *testing.T) {
	part := "{base=0x4a2ab20, len=1378, \"" + strings.Repeat("\x00", 1378) + "\"...}, "
	inputTmpl := "I0303 03:31:30.374817     206 strace.go:591] [  60:  79] node E writev(0x13 /tmp/archive.tar.gz, 0x4c45c70 %s0x6a)"
	input := fmt.Sprintf(inputTmpl, strings.Repeat(part, 1000))

	r := strings.NewReader(input)
	res, err := strace.Parse(context.Background(), r, nopLogger)
	if err != nil || res == nil {
		t.Fatalf(`Parse(r) = %v, %v, want _, nil`, res, err)
	}
	files := res.Files()
	if len(files) != 0 {
		t.Fatalf(`Files() = %v, want []`, files)
	}
}
