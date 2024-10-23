package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v38/github"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// For mocking in tests
var (
	newGitHubClient = func(token string) *github.Client {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		return github.NewClient(tc)
	}

	newGitLabClient = func(token, baseURL string) (*gitlab.Client, error) {
		return gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	}
)

type CommandError struct {
	Command string
	Output  string
	Err     error
}

func (e *CommandError) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("Command '%s' failed: %v\nOutput: %s", e.Command, e.Err, e.Output)
	}
	return fmt.Sprintf("Command '%s' failed: %v", e.Command, e.Err)
}

func RunGitCommandWithOutput(args ...string) error {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &CommandError{
			Command: "git " + strings.Join(args, " "),
			Output:  string(output),
			Err:     err,
		}
	}
	if len(output) > 0 {
		fmt.Print(string(output))
	}
	return nil
}

func ResetRepo(repoInfo RepoInfo, commitMessage string) error {
	// Prepare temporary directory
	tmpPath := filepath.Join(os.TempDir(), "git-tmp")
	if err := os.RemoveAll(tmpPath); err != nil {
		return fmt.Errorf("failed to remove existing temporary directory: %v", err)
	}
	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpPath)

	// Change to temporary directory
	if err := os.Chdir(tmpPath); err != nil {
		return fmt.Errorf("failed to change to temporary directory: %v", err)
	}

	// Clone the repository
	var cloneURL string
	switch repoInfo.Provider {
	case GitHub:
		cloneURL = fmt.Sprintf("https://github.com/%s/%s.git", repoInfo.FullPath, repoInfo.RepoName)
	case GitLab:
		cloneURL = fmt.Sprintf("%s/%s/%s.git", repoInfo.GitLabURL, repoInfo.FullPath, repoInfo.RepoName)
	}
	fmt.Println(info.Render("Cloning repository: git clone %s", cloneURL))

	if err := RunGitCommandWithOutput("clone", cloneURL); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}
	os.Chdir(repoInfo.RepoName)

	// Perform Git operations
	gitOperations := []struct {
		desc string
		args []string
	}{
		{"Creating new orphan branch", []string{"checkout", "--orphan", "temp_branch"}},
		{"Staging all files", []string{"add", "-A"}},
		{"Creating initial commit", []string{"commit", "-m", commitMessage}},
		{"Removing old main branch", []string{"branch", "-D", "main"}},
		{"Renaming branch to main", []string{"branch", "-m", "main"}},
	}

	for _, op := range gitOperations {
		fmt.Println(info.Render("Executing: git %s", strings.Join(op.args, " ")))
		if err := RunGitCommandWithOutput(op.args...); err != nil {
			if !strings.Contains(op.args[0], "branch -D") {
				return fmt.Errorf("failed to %s: %v", op.desc, err)
			}
		}
	}

	// Handle tags deletion
	tags, err := GetGitTags()
	if err != nil {
		return fmt.Errorf("failed to list tags: %v", err)
	}

	if len(tags) > 0 {
		fmt.Println(info.Render("Found %d tags to delete", len(tags)))
		for _, tag := range tags {
			fmt.Println(info.Render("Removing local tag: %s", tag))
			if err := RunGitCommandWithOutput("tag", "-d", tag); err != nil {
				fmt.Println(warning.Render("Warning: Failed to delete local tag %s: %v", tag, err))
			}
		}
	} else {
		fmt.Println(info.Render("No local tags found"))
	}

	// Handle remote operations
	if repoInfo.DryRun {
		if len(tags) > 0 {
			fmt.Printf(info.Render("\nWould delete %d remote tags: %v\n"), len(tags), tags)
		}
		fmt.Println(info.Render("Would execute: git push -f origin main"))
		return nil
	}

	// Delete remote tags
	if len(tags) > 0 {
		for _, tag := range tags {
			if err := RunGitCommandWithOutput("push", "origin", "--delete", fmt.Sprintf("refs/tags/%s", tag)); err != nil {
				fmt.Println(warning.Render("Warning: Failed to delete remote tag %s: %v", tag, err))
			} else {
				fmt.Println(success.Render("Deleted remote tag: %s", tag))
			}
		}
	}

	// Force push the new main branch
	if err := RunGitCommandWithOutput("push", "-f", "origin", "main"); err != nil {
		return fmt.Errorf("failed to push changes: %v", err)
	}

	// Delete releases
	switch repoInfo.Provider {
	case GitHub:
		return DeleteGitHubReleases(repoInfo)
	case GitLab:
		return DeleteGitLabReleases(repoInfo)
	default:
		return fmt.Errorf("unsupported git provider")
	}
}

func GetGitTags() ([]string, error) {
	cmd := exec.Command("git", "tag")
	output, err := cmd.Output()
	if err != nil {
		return nil, &CommandError{
			Command: "git tag",
			Output:  string(output),
			Err:     err,
		}
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		return []string{}, nil
	}
	return tags, nil
}

func DeleteGitHubReleases(repoInfo RepoInfo) error {
	client := newGitHubClient(repoInfo.Token)
	// ... [previous functions.go content] ...

func DeleteGitHubReleases(repoInfo RepoInfo) error {
	client := newGitHubClient(repoInfo.Token)
	ctx := context.Background()

	releases, _, err := client.Repositories.ListReleases(ctx, repoInfo.FullPath, repoInfo.RepoName, nil)
	if err != nil {
		return fmt.Errorf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		fmt.Println(info.Render("No releases found to delete"))
		return nil
	}

	fmt.Println(info.Render("Found %d releases", len(releases)))

	if repoInfo.DryRun {
		fmt.Println(info.Render("\nThe following releases would be deleted:"))
		for _, release := range releases {
			fmt.Printf(info.Render("- Release %d: %s (tag: %s)\n"),
				*release.ID,
				*release.Name,
				*release.TagName)
		}
	} else {
		for _, release := range releases {
			_, err := client.Repositories.DeleteRelease(ctx, repoInfo.FullPath, repoInfo.RepoName, *release.ID)
			if err != nil {
				fmt.Println(warning.Render("Warning: Failed to delete release %d: %v", *release.ID, err))
			} else {
				fmt.Println(success.Render("Deleted release %d: %s", *release.ID, *release.Name))
			}
		}
	}

	return nil
}

func DeleteGitLabReleases(repoInfo RepoInfo) error {
	client, err := newGitLabClient(repoInfo.Token, repoInfo.GitLabURL)
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %v", err)
	}

	fullPath := repoInfo.FullPath + "/" + repoInfo.RepoName
	releases, resp, err := client.Releases.ListReleases(fullPath, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("failed to list releases (status %d): %v", resp.StatusCode, err)
		}
		return fmt.Errorf("failed to list releases: %v", err)
	}

	if len(releases) == 0 {
		fmt.Println(info.Render("No releases found to delete"))
		return nil
	}

	fmt.Println(info.Render("Found %d releases", len(releases)))

	if repoInfo.DryRun {
		fmt.Println(info.Render("\nThe following releases would be deleted:"))
		for _, release := range releases {
			fmt.Printf(info.Render("- Release: %s (tag: %s)\n"),
				release.Name,
				release.TagName)
		}
	} else {
		for _, release := range releases {
			_, resp, err := client.Releases.DeleteRelease(fullPath, release.TagName)
			if err != nil {
				if resp != nil {
					fmt.Println(warning.Render("Warning: Failed to delete release %s (status %d): %v",
						release.TagName, resp.StatusCode, err))
				} else {
					fmt.Println(warning.Render("Warning: Failed to delete release %s: %v",
						release.TagName, err))
				}
			} else {
				fmt.Println(success.Render("Deleted release: %s", release.Name))
			}
		}
	}

	return nil
}