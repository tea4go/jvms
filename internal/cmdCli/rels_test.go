package cmdCli

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tea4go/jvms/internal/entity"
	"github.com/urfave/cli/v2"
)

func TestRlsCmdShowsEveryVersionFromLocalCache(t *testing.T) {
	exePath, err := os.Executable()
	if err != nil {
		t.Fatalf("get test executable path: %v", err)
	}

	cacheFile := filepath.Join(filepath.Dir(exePath), "jdkdlindex.json")
	versions := []entity.TJDKVersion{
		{Version: "openjdk-26.0.1+8", Url: "https://example.test/26.zip"},
		{Version: "openjdk-25.0.3+9-LTS", Url: "https://example.test/25.zip"},
		{Version: "openjdk-24.0.2+12", Url: "https://example.test/24.zip"},
		{Version: "openjdk-23.0.2+7", Url: "https://example.test/23.zip"},
		{Version: "openjdk-22.0.2+9", Url: "https://example.test/22.zip"},
		{Version: "openjdk-21.0.11+10-LTS", Url: "https://example.test/21.zip"},
		{Version: "openjdk-20.0.2+9", Url: "https://example.test/20.zip"},
		{Version: "openjdk-19.0.2+7", Url: "https://example.test/19.zip"},
		{Version: "openjdk-18.0.2.1+1", Url: "https://example.test/18.zip"},
		{Version: "openjdk-17.0.19+10", Url: "https://example.test/17.zip"},
		{Version: "openjdk-16.0.2+7", Url: "https://example.test/16.zip"},
		{Version: "openjdk-11.0.31+11", Url: "https://example.test/11.zip"},
		{Version: "openjdk-1.8.0_492-b09", Url: "https://example.test/8.zip"},
	}
	data, err := json.Marshal(versions)
	if err != nil {
		t.Fatalf("marshal versions: %v", err)
	}
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		t.Fatalf("write cache file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(cacheFile)
	})

	output := captureStdout(t, func() {
		ctx := newRlsTestContext(false)
		if err := rlsCmd(ctx, &entity.TConfig{}); err != nil {
			t.Fatalf("rlsCmd returned error: %v", err)
		}
	})

	if !strings.Contains(output, " 13) openjdk-1.8.0_492-b09") {
		t.Fatalf("expected rls to print every cached version, got:\n%s", output)
	}
	if strings.Contains(output, "show all versions") {
		t.Fatalf("expected no truncation hint when local cache is used, got:\n%s", output)
	}
}

func TestShouldSaveJdkVersionsCacheSkipsShowAllResult(t *testing.T) {
	if shouldSaveJdkVersionsCache(&entity.TConfig{WebAll: true}) {
		t.Fatal("expected show-all version results to avoid overwriting the normal cache")
	}

	if !shouldSaveJdkVersionsCache(&entity.TConfig{WebAll: true, RefreshCache: true}) {
		t.Fatal("expected explicit cache refresh to save show-all version results")
	}

	if !shouldSaveJdkVersionsCache(&entity.TConfig{WebAll: false}) {
		t.Fatal("expected normal version results to be saved in the cache")
	}
}

func TestShouldRefreshJdkVersionsCacheWhenWebtypeIsExplicit(t *testing.T) {
	if !shouldRefreshJdkVersionsCache(newRlsTestContextWithArgs(t, "-a", "--webtype", "huawei")) {
		t.Fatal("expected explicit webtype to refresh the version cache")
	}

	if !shouldRefreshJdkVersionsCache(newRlsTestContextWithArgs(t, "-a", "-t", "huawei")) {
		t.Fatal("expected explicit webtype alias to refresh the version cache")
	}

	if shouldRefreshJdkVersionsCache(newRlsTestContextWithArgs(t, "-a")) {
		t.Fatal("expected show-all without explicit webtype to avoid refreshing the version cache")
	}
}

func newRlsTestContext(showAll bool) *cli.Context {
	set := flag.NewFlagSet("rls", flag.ContinueOnError)
	set.Bool("s", false, "")
	set.Bool("a", showAll, "")
	set.String("webtype", "lzu", "")
	return cli.NewContext(cli.NewApp(), set, nil)
}

func newRlsTestContextWithArgs(t *testing.T, args ...string) *cli.Context {
	t.Helper()

	set := flag.NewFlagSet("rls", flag.ContinueOnError)
	set.Bool("s", false, "")
	set.Bool("a", false, "")
	set.String("webtype", "lzu", "")
	set.String("t", "", "")
	if err := set.Parse(args); err != nil {
		t.Fatalf("parse rls flags: %v", err)
	}
	return cli.NewContext(cli.NewApp(), set, nil)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = writer

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(&buf, reader)
		done <- err
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original
	if err := <-done; err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}

	return buf.String()
}
