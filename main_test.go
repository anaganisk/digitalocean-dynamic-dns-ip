package main

import (
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anaganisk/digitalocean-dynamic-dns-ip/config"
	"github.com/anaganisk/digitalocean-dynamic-dns-ip/logger"
)

func TestMainSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("203.0.113.5"))
	}))
	defer server.Close()

	configJSON := `{
		"apiKey": "testkey",
		"doPageSize": 20,
		"useIPv4": true,
		"useIPv6": false,
		"ipv4CheckUrl": "` + server.URL + `",
		"allowIPv4InIPv6": false,
		"domains": []
	}`

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("unable to write config file: %v", err)
	}

	oldArgs := os.Args
	oldFlags := flag.CommandLine
	oldConfig := config.Get()
	oldExit := logger.ExitFunc
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlags
		config.Set(oldConfig)
		logger.ExitFunc = oldExit
	}()

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"test", configPath}
	logger.ExitFunc = func(code int) {
		t.Fatalf("main exited with code %d", code)
	}

	main()
}

func TestMainFailureWhenNoIPAddressesFound(t *testing.T) {
	configJSON := `{
		"apiKey": "testkey",
		"doPageSize": 20,
		"useIPv4": false,
		"useIPv6": false,
		"domains": []
	}`

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("unable to write config file: %v", err)
	}

	oldArgs := os.Args
	oldFlags := flag.CommandLine
	oldExit := logger.ExitFunc
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlags
		logger.ExitFunc = oldExit
	}()

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"test", configPath}

	exited := false
	logger.ExitFunc = func(code int) {
		exited = true
		if code != 1 {
			t.Fatalf("expected exit code 1, got %d", code)
		}
	}

	main()

	if !exited {
		t.Fatal("expected main to exit with code 1")
	}
}

func TestUsageOutput(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unable to create pipe: %v", err)
	}
	os.Stdout = w
	usage()
	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unable to read usage output: %v", err)
	}
	if !strings.Contains(string(output), "-h | -help") {
		t.Fatalf("usage output missing expected text: %s", output)
	}
}

// TestMainDebugOutputWithFlag validates that debug output is sent to stdout when -d flag is set
func TestMainDebugOutputWithFlag(t *testing.T) {
	configJSON := `{
		"apiKey": "testkey",
		"doPageSize": 20,
		"useIPv4": true,
		"useIPv6": false,
		"domains": []
	}`

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("unable to write config file: %v", err)
	}

	oldArgs := os.Args
	oldFlags := flag.CommandLine
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlags
		os.Stdout = oldStdout
		// Reset logger after test
		logger.SetDebugOutput(os.Stdout)
	}()

	// Capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unable to create pipe: %v", err)
	}
	os.Stdout = w

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"test", "-d", configPath}

	run()

	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unable to read output: %v", err)
	}

	if !strings.Contains(string(output), "Using config file:") {
		t.Errorf("debug output missing 'Using config file:' message when -d flag is set. Got: %s", output)
	}
}

// TestMainDebugOutputWithoutFlag validates that debug output is discarded when -d flag is not set
func TestMainDebugOutputWithoutFlag(t *testing.T) {
	configJSON := `{
		"apiKey": "testkey",
		"doPageSize": 20,
		"useIPv4": true,
		"useIPv6": false,
		"domains": []
	}`

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, []byte(configJSON), 0600); err != nil {
		t.Fatalf("unable to write config file: %v", err)
	}

	oldArgs := os.Args
	oldFlags := flag.CommandLine
	oldStdout := os.Stdout
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlags
		os.Stdout = oldStdout
		// Reset logger after test
		logger.SetDebugOutput(os.Stdout)
	}()

	// Capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unable to create pipe: %v", err)
	}
	os.Stdout = w

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{"test", configPath}

	run()

	w.Close()
	os.Stdout = oldStdout

	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("unable to read output: %v", err)
	}

	if strings.Contains(string(output), "Using Config file:") {
		t.Errorf("debug output should be suppressed when -d flag is not set. Got: %s", output)
	}
}
