package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// version will be set during build
var version = "dev"

const (
	tmpDir           = "git-tmp"
	defaultCommitMsg = "Initial commit"
)

func parseFlags() CommandLineFlags {
	flags := CommandLineFlags{}

	fs := flag.NewFlagSet("goresetit", flag.ExitOnError)

	// Add version flag
	showVersion := fs.Bool("version", false, "Show version information")
	fs.BoolVar(showVersion, "v", false, "Show version information")

	// Repository path
	fs.StringVar(&flags.RepoPath, "repo", "", "")
	fs.StringVar(&flags.RepoPath, "r", "", "Repository path (e.g., owner/repo or group/subgroup/repo)")

	// Token
	fs.StringVar(&flags.Token, "token", "", "")
	fs.StringVar(&flags.Token, "t", "", "Personal access token")

	// Provider
	fs.StringVar(&flags.Provider, "provider", "github", "")
	fs.StringVar(&flags.Provider, "p", "github", "Git provider (github or gitlab)")

	// GitLab URL
	fs.StringVar(&flags.GitLabURL, "gitlab-url", "https://gitlab.com", "")
	fs.StringVar(&flags.GitLabURL, "g", "https://gitlab.com", "GitLab instance URL (for private instances)")

	// Dry run
	fs.BoolVar(&flags.DryRun, "dry-run", false, "")
	fs.BoolVar(&flags.DryRun, "d", false, "Perform a dry run without making actual changes")

	// No interactive
	fs.BoolVar(&flags.NoInteractive, "no-interactive", false, "")
	fs.BoolVar(&flags.NoInteractive, "n", false, "Run without interactive prompts")

	// Commit message
	fs.StringVar(&flags.CommitMsg, "message", "", "")
	fs.StringVar(&flags.CommitMsg, "m", "", "Specify commit message (skips message prompt if provided)")

	// Custom usage message
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of GoresetIT:\n")
		fmt.Fprintf(os.Stderr, "  goresetit -r owner/repo -t <token> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -v, --version            Show version information\n")
		fmt.Fprintf(os.Stderr, "  -r, --repo string        Repository path (e.g., owner/repo or group/subgroup/repo)\n")
		fmt.Fprintf(os.Stderr, "  -t, --token string       Personal access token\n")
		fmt.Fprintf(os.Stderr, "  -p, --provider string    Git provider (github or gitlab) (default: github)\n")
		fmt.Fprintf(os.Stderr, "  -g, --gitlab-url string  GitLab instance URL (for private instances) (default: https://gitlab.com)\n")
		fmt.Fprintf(os.Stderr, "  -d, --dry-run           Perform a dry run without making actual changes\n")
		fmt.Fprintf(os.Stderr, "  -n, --no-interactive    Run without interactive prompts (uses default commit message if -m not provided)\n")
		fmt.Fprintf(os.Stderr, "  -m, --message string     Specify commit message (skips message prompt if provided)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Interactive mode with custom commit message:\n")
		fmt.Fprintf(os.Stderr, "  goresetit -r owner/repo -t <token> -m \"feat: fresh start\"\n\n")
		fmt.Fprintf(os.Stderr, "  # Non-interactive mode with custom commit message:\n")
		fmt.Fprintf(os.Stderr, "  goresetit -r owner/repo -t <token> -n -m \"feat: fresh start\"\n\n")
		fmt.Fprintf(os.Stderr, "  # Dry run with default commit message:\n")
		fmt.Fprintf(os.Stderr, "  goresetit -r owner/repo -t <token> -d -n\n")
	}

	fs.Parse(os.Args[1:])

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("GoresetIT version %s\n", version)
		os.Exit(0)
	}

	return flags
}

func main() {
	ShowLogo()

	flags := parseFlags()

	if flags.RepoPath == "" || flags.Token == "" {
		fmt.Println(errorStyle.Render("Error: Missing required arguments."))
		flag.Usage()
		os.Exit(1)
	}

	parts := strings.Split(flags.RepoPath, "/")
	if len(parts) < 2 {
		fmt.Println(errorStyle.Render("Error: Invalid repository format. Please use full path format (e.g., owner/repo or group/subgroup/repo)."))
		flag.Usage()
		os.Exit(1)
	}

	repoName := parts[len(parts)-1]
	fullPath := strings.Join(parts[:len(parts)-1], "/")

	var repoInfo RepoInfo
	repoInfo.FullPath = fullPath
	repoInfo.RepoName = repoName
	repoInfo.Token = flags.Token
	repoInfo.DryRun = flags.DryRun

	switch strings.ToLower(flags.Provider) {
	case "github":
		repoInfo.Provider = GitHub
	case "gitlab":
		repoInfo.Provider = GitLab
		repoInfo.GitLabURL = flags.GitLabURL
	default:
		fmt.Println(errorStyle.Render("Error: Invalid provider. Use 'github' or 'gitlab'."))
		os.Exit(1)
	}

	var commitMessage string

	// Determine commit message source
	if flags.CommitMsg != "" {
		// Use provided message from flag
		commitMessage = flags.CommitMsg
		fmt.Printf(info.Render("Using provided commit message: '%s'\n"), commitMessage)
	} else if flags.NoInteractive {
		// Use default message in non-interactive mode
		commitMessage = defaultCommitMsg
		fmt.Printf(info.Render("Using default commit message: '%s'\n"), commitMessage)
	}

	// Show confirmation unless in non-interactive mode
	if !flags.NoInteractive {
		// Show confirmation prompt
		confirmed, err := PromptConfirmation(flags.DryRun)
		if err != nil {
			fmt.Println(errorStyle.Render("Error during confirmation:", err))
			os.Exit(1)
		}
		if !confirmed {
			fmt.Println(info.Render("Operation cancelled by user"))
			os.Exit(0)
		}

		// Only prompt for commit message if not provided via flag
		if commitMessage == "" {
			message, err := PromptCommitMessage()
			if err != nil {
				fmt.Println(errorStyle.Render("Error getting commit message:", err))
				os.Exit(1)
			}
			if message == "" {
				fmt.Println(info.Render("Operation cancelled by user"))
				os.Exit(0)
			}
			commitMessage = message
		}
	} else {
		// Still show what's going to happen in non-interactive mode
		if flags.DryRun {
			fmt.Println(warning.Render("DRY RUN: Will simulate squashing all commits on main branch"))
		} else {
			fmt.Println(warning.Render("WARNING: Will squash all commits on main branch (no interactive confirmation requested)"))
		}
	}

	if err := ResetRepo(repoInfo, commitMessage); err != nil {
		fmt.Println(errorStyle.Render("Error:", err))
		os.Exit(1)
	}

	if flags.DryRun {
		fmt.Println(info.Render("\nDry run completed. No changes were pushed to remote."))
	} else {
		fmt.Printf(success.Render("\nRepository %s/%s has been reset with message: '%s'\n"),
			repoInfo.FullPath, repoInfo.RepoName, commitMessage)
		fmt.Println(success.Render("All tags and releases have been deleted."))
	}
}
