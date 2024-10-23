package main_test

import (
	"flag"
	"os"
	"testing"

	main "github.com/Moukrea/goresetit"
)

func TestParseFlags(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected main.CommandLineFlags
		wantErr  bool
	}{
		{
			name: "Valid GitHub flags",
			args: []string{"-r", "owner/repo", "-t", "token"},
			expected: main.CommandLineFlags{
				RepoPath: "owner/repo",
				Token:    "token",
				Provider: "github",
			},
			wantErr: false,
		},
		{
			name: "Valid GitLab flags with URL",
			args: []string{
				"-r", "group/repo",
				"-t", "token",
				"-p", "gitlab",
				"-g", "https://gitlab.company.com",
			},
			expected: main.CommandLineFlags{
				RepoPath:  "group/repo",
				Token:     "token",
				Provider:  "gitlab",
				GitLabURL: "https://gitlab.company.com",
			},
			wantErr: false,
		},
		{
			name: "Valid flags with short versions",
			args: []string{
				"-r", "owner/repo",
				"-t", "token",
				"-d",
				"-n",
				"-m", "test commit",
			},
			expected: main.CommandLineFlags{
				RepoPath:      "owner/repo",
				Token:         "token",
				Provider:      "github",
				DryRun:        true,
				NoInteractive: true,
				CommitMsg:     "test commit",
			},
			wantErr: false,
		},
		{
			name:    "Missing required flags",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Missing token",
			args:    []string{"-r", "owner/repo"},
			wantErr: true,
		},
		{
			name:    "Missing repo",
			args:    []string{"-t", "token"},
			wantErr: true,
		},
		{
			name: "Invalid provider",
			args: []string{
				"-r", "owner/repo",
				"-t", "token",
				"-p", "invalid",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			
			// Set test args
			os.Args = append([]string{"cmd"}, tc.args...)
			
			flags := main.ParseFlags()
			
			if tc.wantErr {
				if flags.RepoPath != "" && flags.Token != "" {
					t.Error("Expected empty flags, got values")
				}
				return
			}

			if flags.RepoPath != tc.expected.RepoPath {
				t.Errorf("Expected repoPath %s, got %s", tc.expected.RepoPath, flags.RepoPath)
			}
			if flags.Token != tc.expected.Token {
				t.Errorf("Expected token %s, got %s", tc.expected.Token, flags.Token)
			}
			if flags.Provider != tc.expected.Provider {
				t.Errorf("Expected provider %s, got %s", tc.expected.Provider, flags.Provider)
			}
			if flags.GitLabURL != tc.expected.GitLabURL {
				t.Errorf("Expected GitLabURL %s, got %s", tc.expected.GitLabURL, flags.GitLabURL)
			}
			if flags.DryRun != tc.expected.DryRun {
				t.Errorf("Expected dryRun %v, got %v", tc.expected.DryRun, flags.DryRun)
			}
			if flags.NoInteractive != tc.expected.NoInteractive {
				t.Errorf("Expected noInteractive %v, got %v", tc.expected.NoInteractive, flags.NoInteractive)
			}
			if flags.CommitMsg != tc.expected.CommitMsg {
				t.Errorf("Expected commitMsg %s, got %s", tc.expected.CommitMsg, flags.CommitMsg)
			}
		})
	}
}

func TestMainIntegration(t *testing.T) {
	// This test verifies the main flow works correctly
	// We'll mock the git commands and API calls
	cleanup := setupTestEnv(t)
	defer cleanup()

	mockGit := mockGitCommand{
		expectedCmds: []string{
			"git clone https://github.com/owner/repo.git",
			"git checkout --orphan temp_branch",
			"git add -A",
			"git commit -m test commit",
			"git branch -D main",
			"git branch -m main",
			"git push -f origin main",
		},
		outputs: make([]string, 7),
		errors:  make([]error, 7),
	}

	cleanupGit := setupMockGit(t, mockGit)
	defer cleanupGit()

	// Set up test arguments
	os.Args = []string{
		"cmd",
		"-r", "owner/repo",
		"-t", "token",
		"-n",
		"-m", "test commit",
	}

	// Capture output
	output := captureOutput(func() {
		main.Main()
	})

	// Verify output contains expected messages
	expectedMessages := []string{
		"GoresetIT",
		"Using provided commit message",
		"test commit",
		"Repository owner/repo has been reset",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected output to contain '%s'", msg)
		}
	}
}