package main_test

import (
	"fmt"
	"testing"

	"github.com/google/go-github/v38/github"
	"github.com/xanzy/go-gitlab"
	main "github.com/Moukrea/goresetit"
)

func TestResetRepo(t *testing.T) {
	testCases := []struct {
		name        string
		repoInfo    main.RepoInfo
		commitMsg   string
		mockGit     mockGitCommand
		expectError bool
	}{
		{
			name: "Successful GitHub reset",
			repoInfo: main.RepoInfo{
				Provider: main.GitHub,
				FullPath: "owner",
				RepoName: "repo",
				Token:    "token",
			},
			commitMsg: "test commit",
			mockGit: mockGitCommand{
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
			},
			expectError: false,
		},
		{
			name: "Failed clone",
			repoInfo: main.RepoInfo{
				Provider: main.GitHub,
				FullPath: "owner",
				RepoName: "repo",
				Token:    "token",
			},
			commitMsg: "test commit",
			mockGit: mockGitCommand{
				expectedCmds: []string{
					"git clone https://github.com/owner/repo.git",
				},
				outputs: []string{"error cloning"},
				errors:  []error{fmt.Errorf("clone failed")},
			},
			expectError: true,
		},
		{
			name: "Successful GitLab reset",
			repoInfo: main.RepoInfo{
				Provider:  main.GitLab,
				FullPath:  "group",
				RepoName: "repo",
				Token:    "token",
				GitLabURL: "https://gitlab.com",
			},
			commitMsg: "test commit",
			mockGit: mockGitCommand{
				expectedCmds: []string{
					"git clone https://gitlab.com/group/repo.git",
					"git checkout --orphan temp_branch",
					"git add -A",
					"git commit -m test commit",
					"git branch -D main",
					"git branch -m main",
					"git push -f origin main",
				},
				outputs: make([]string, 7),
				errors:  make([]error, 7),
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupTestEnv(t)
			defer cleanup()

			cleanupGit := setupMockGit(t, tc.mockGit)
			defer cleanupGit()

			err := main.ResetRepo(tc.repoInfo, tc.commitMsg)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetGitTags(t *testing.T) {
	testCases := []struct {
		name        string
		mockOutput  string
		mockError   error
		expectedTags []string
		expectError bool
	}{
		{
			name:         "No tags",
			mockOutput:   "",
			expectedTags: []string{},
			expectError:  false,
		},
		{
			name:         "Single tag",
			mockOutput:   "v1.0.0\n",
			expectedTags: []string{"v1.0.0"},
			expectError:  false,
		},
		{
			name:         "Multiple tags",
			mockOutput:   "v1.0.0\nv1.1.0\nv2.0.0\n",
			expectedTags: []string{"v1.0.0", "v1.1.0", "v2.0.0"},
			expectError:  false,
		},
		{
			name:        "Command error",
			mockError:   fmt.Errorf("git command failed"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupTestEnv(t)
			defer cleanup()

			mockGit := mockGitCommand{
				expectedCmds: []string{"git tag"},
				outputs:     []string{tc.mockOutput},
				errors:     []error{tc.mockError},
			}

			cleanupGit := setupMockGit(t, mockGit)
			defer cleanupGit()

			tags, err := main.GetGitTags()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(tags) != len(tc.expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(tc.expectedTags), len(tags))
			}

			for i, tag := range tags {
				if tag != tc.expectedTags[i] {
					t.Errorf("Expected tag %s, got %s", tc.expectedTags[i], tag)
				}
			}
		})
	}
}

func TestDeleteGitHubReleases(t *testing.T) {
	testCases := []struct {
		name        string
		repoInfo    main.RepoInfo
		releases    []*github.RepositoryRelease
		listErr     error
		deleteErr   error
		expectError bool
	}{
		{
			name: "No releases",
			repoInfo: main.RepoInfo{
				FullPath: "owner",
				RepoName: "repo",
				Token:    "token",
			},
			releases:    []*github.RepositoryRelease{},
			expectError: false,
		},
		{
			name: "Multiple releases",
			repoInfo: main.RepoInfo{
				FullPath: "owner",
				RepoName: "repo",
				Token:    "token",
			},
			releases: []*github.RepositoryRelease{
				{
					ID:      github.Int64(1),
					Name:    github.String("v1.0.0"),
					TagName: github.String("v1.0.0"),
				},
				{
					ID:      github.Int64(2),
					Name:    github.String("v1.1.0"),
					TagName: github.String("v1.1.0"),
				},
			},
			expectError: false,
		},
		{
			name: "List error",
			repoInfo: main.RepoInfo{
				FullPath: "owner",
				RepoName: "repo",
				Token:    "token",
			},
			listErr:     fmt.Errorf("API error"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock GitHub client
			oldNewGitHubClient := main.NewGitHubClient
			defer func() { main.NewGitHubClient = oldNewGitHubClient }()

			main.NewGitHubClient = func(token string) *github.Client {
				return &github.Client{
					Repositories: &mockGitHubClient{
						releases: tc.releases,
						err:      tc.listErr,
					},
				}
			}

			err := main.DeleteGitHubReleases(tc.repoInfo)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDeleteGitLabReleases(t *testing.T) {
	testCases := []struct {
		name        string
		repoInfo    main.RepoInfo
		releases    []*gitlab.Release
		listErr     error
		deleteErr   error
		expectError bool
	}{
		{
			name: "No releases",
			repoInfo: main.RepoInfo{
				FullPath:  "group",
				RepoName: "repo",
				Token:    "token",
			},
			releases:    []*gitlab.Release{},
			expectError: false,
		},
		{
			name: "Multiple releases",
			repoInfo: main.RepoInfo{
				FullPath:  "group",
				RepoName: "repo",
				Token:    "token",
			},
			releases: []*gitlab.Release{
				{
					Name:    "v1.0.0",
					TagName: "v1.0.0",
				},
				{
					Name:    "v1.1.0",
					TagName: "v1.1.0",
				},
			},
			expectError: false,
		},
		{
			name: "List error",
			repoInfo: main.RepoInfo{
				FullPath:  "group",
				RepoName: "repo",
				Token:    "token",
			},
			listErr:     fmt.Errorf("API error"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock GitLab client
			oldNewGitLabClient := main.NewGitLabClient
			defer func() { main.NewGitLabClient = oldNewGitLabClient }()

			main.NewGitLabClient = func(token, baseURL string) (*gitlab.Client, error) {
				if tc.listErr != nil {
					return nil, tc.listErr
				}
				return &gitlab.Client{
					Releases: &mockGitLabClient{
						releases: tc.releases,
						err:      tc.deleteErr,
					},
				}, nil
			}

			err := main.DeleteGitLabReleases(tc.repoInfo)

            if tc.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }

            if err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
        })
    }
}