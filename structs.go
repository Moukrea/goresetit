package main

// GitProvider represents the supported Git hosting providers
type GitProvider int

const (
	GitHub GitProvider = iota
	GitLab
)

// RepoInfo contains all repository-related information
type RepoInfo struct {
	Provider  GitProvider
	FullPath  string
	RepoName  string
	Token     string
	GitLabURL string
	DryRun    bool
}

// CommandLineFlags holds all possible command line arguments
type CommandLineFlags struct {
	RepoPath      string
	Token         string
	Provider      string
	GitLabURL     string
	DryRun        bool
	NoInteractive bool
	CommitMsg     string
}