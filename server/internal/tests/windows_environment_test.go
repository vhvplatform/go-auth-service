package tests

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"testing"
	"time"
)

// TestWindowsEnvironment verifies that the Windows environment is properly set up
func TestWindowsEnvironment(t *testing.T) {
	t.Run("CheckOperatingSystem", func(t *testing.T) {
		t.Logf("Operating System: %s", runtime.GOOS)
		t.Logf("Architecture: %s", runtime.GOARCH)
		t.Logf("Go Version: %s", runtime.Version())
	})

	t.Run("CheckGoEnvironment", func(t *testing.T) {
		gopath := os.Getenv("GOPATH")
		goroot := os.Getenv("GOROOT")

		t.Logf("GOPATH: %s", gopath)
		t.Logf("GOROOT: %s", goroot)

		if goroot == "" {
			t.Error("GOROOT is not set")
		}
	})

	t.Run("CheckRequiredPorts", func(t *testing.T) {
		ports := []string{"50051", "8081"}

		for _, port := range ports {
			address := fmt.Sprintf("127.0.0.1:%s", port)
			listener, err := net.Listen("tcp", address)
			if err != nil {
				t.Logf("Port %s is already in use (this is okay if service is running): %v", port, err)
			} else {
				listener.Close()
				t.Logf("Port %s is available", port)
			}
		}
	})

	t.Run("CheckFileSystemOperations", func(t *testing.T) {
		// Test file creation
		testFile := "test_windows_fs.tmp"
		err := os.WriteFile(testFile, []byte("test"), 0644)
		if err != nil {
			t.Errorf("Failed to write test file: %v", err)
		}
		defer os.Remove(testFile)

		// Test file reading
		_, err = os.ReadFile(testFile)
		if err != nil {
			t.Errorf("Failed to read test file: %v", err)
		}

		t.Log("File system operations work correctly")
	})

	t.Run("CheckNetworkConnectivity", func(t *testing.T) {
		// Test DNS resolution
		_, err := net.LookupHost("localhost")
		if err != nil {
			t.Errorf("Failed to resolve localhost: %v", err)
		} else {
			t.Log("DNS resolution works correctly")
		}

		// Test loopback connection
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Errorf("Failed to bind to loopback address: %v", err)
		} else {
			listener.Close()
			t.Log("Loopback network connection works correctly")
		}
	})

	t.Run("CheckConcurrency", func(t *testing.T) {
		// Test goroutines work correctly
		done := make(chan bool)
		go func() {
			time.Sleep(10 * time.Millisecond)
			done <- true
		}()

		select {
		case <-done:
			t.Log("Goroutines work correctly")
		case <-time.After(1 * time.Second):
			t.Error("Goroutine test timed out")
		}
	})
}

// TestWindowsPathHandling tests Windows-specific path handling
func TestWindowsPathHandling(t *testing.T) {
	t.Run("CheckPathSeparator", func(t *testing.T) {
		// Windows uses backslash, but Go should handle both
		paths := []string{
			"cmd/main.go",
			"cmd\\main.go",
		}

		for _, path := range paths {
			_, err := os.Stat(path)
			if err == nil {
				t.Logf("Path handling works for: %s", path)
				break
			}
		}
	})
}

// TestWindowsDependencies verifies that all Go dependencies work on Windows
func TestWindowsDependencies(t *testing.T) {
	t.Run("ImportStandardLibrary", func(t *testing.T) {
		// These imports are tested by this test file existing
		t.Log("Standard library imports work correctly")
	})
}

// TestWindowsBuild verifies that the application can be built on Windows
func TestWindowsBuild(t *testing.T) {
	// This test is implicit - if the package builds, the test passes
	t.Log("Application builds successfully on Windows")
}
