package main_test

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockTeaProgram mocks bubbletea.Program for testing UI components
type mockTeaProgram struct {
	msgs    []tea.Msg
	result  string
	hasErr  bool
}

func (m *mockTeaProgram) Run() (tea.Model, error) {
	if m.hasErr {
		return nil, fmt.Errorf("mock program error")
	}
	return m, nil
}

// captureOutput captures stdout for testing
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	io.Copy(&buf, r)
	return buf.String()
}

// Mock git command executor for testing
type mockGitCommand struct {
	expectedCmds []string
	outputs      []string
	errors       []error
	cmdIndex     int
}

func (m *mockGitCommand) execute(args ...string) error {
	if m.cmdIndex >= len(m.expectedCmds) {
		return fmt.Errorf("unexpected command: git %v", args)
	}

	cmd := "git " + strings.Join(args, " ")
	if cmd != m.expectedCmds[m.cmdIndex] {
		return fmt.Errorf("expected command '%s', got '%s'", m.expectedCmds[m.cmdIndex], cmd)
	}

	if m.errors[m.cmdIndex] != nil {
		return m.errors[m.cmdIndex]
	}

	if len(m.outputs) > m.cmdIndex && m.outputs[m.cmdIndex] != "" {
		fmt.Print(m.outputs[m.cmdIndex])
	}

	m.cmdIndex++
	return nil
}

// setupTestEnv creates a temporary directory and changes to it
func setupTestEnv(t *testing.T) (cleanup func()) {
	t.Helper()

	// Save current directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create and change to temporary directory
	tmpDir, err := os.MkdirTemp("", "goresetit-test-*")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatal(err)
	}

	// Return cleanup function
	return func() {
		os.Chdir(currentDir)
		os.RemoveAll(tmpDir)
	}
}

// mockGitHubClient mocks the GitHub API client
type mockGitHubClient struct {
	releases []*github.RepositoryRelease
	err      error
}

// mockGitLabClient mocks the GitLab API client
type mockGitLabClient struct {
	releases []*gitlab.Release
	err      error
}

// setupMockGit replaces the git command executor with a mock
func setupMockGit(t *testing.T, mock mockGitCommand) func() {
	t.Helper()
	oldRunGitCommand := runGitCommand
	runGitCommand = mock.execute
	return func() {
		runGitCommand = oldRunGitCommand
	}
}