package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/google/go-github/v67/github"
	"github.com/hashicorp/go-version"
	"golang.org/x/oauth2"
)

var token = os.Getenv("GITHUB_TOKEN")
var versionRegexp = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
var goModRequireSDKRegexp = regexp.MustCompile(`github\.com/terraform-linters/tflint-plugin-sdk v(.+)`)
var goModRequireBundledPluginRegexp = regexp.MustCompile(`github.com/terraform-linters/tflint-ruleset-terraform v(.+)`)

func main() {
	currentVersion := getCurrentVersion()
	log.Printf("current version: %s", currentVersion)

	newVersion := getNewVersion()
	log.Printf("new version: %s", newVersion)

	releaseNotePath := "tools/release/release-note.md"

	log.Println("checking requirements...")
	if err := checkRequirements(currentVersion, newVersion); err != nil {
		log.Fatal(err)
	}

	log.Println("rewriting files with new version...")
	if err := rewriteFileWithNewVersion("tflint/meta.go", currentVersion, newVersion); err != nil {
		log.Fatal(err)
	}
	if err := rewriteFileWithNewVersion(".github/ISSUE_TEMPLATE/bug.yml", currentVersion, newVersion); err != nil {
		log.Fatal(err)
	}

	log.Println("generating release notes...")
	if err := generateReleaseNote(currentVersion, newVersion, releaseNotePath); err != nil {
		log.Fatal(err)
	}
	if err := editFileInteractive(releaseNotePath); err != nil {
		log.Fatal(err)
	}

	log.Println("running tests...")
	if err := execCommand(os.Stdout, "make", "test"); err != nil {
		log.Fatal(err)
	}
	if err := execCommand(os.Stdout, "make", "e2e"); err != nil {
		log.Fatal(err)
	}

	log.Println("committing and tagging...")
	if err := execCommand(os.Stdout, "git", "add", "."); err != nil {
		log.Fatal(err)
	}
	if err := execCommand(os.Stdout, "git", "commit", "-m", fmt.Sprintf("Bump up version to v%s", newVersion)); err != nil {
		log.Fatal(err)
	}
	if err := execCommand(os.Stdout, "git", "tag", fmt.Sprintf("v%s", newVersion)); err != nil {
		log.Fatal(err)
	}

	if err := execCommand(os.Stdout, "git", "push", "origin", "master", "--tags"); err != nil {
		log.Fatal(err)
	}
	log.Printf("pushed v%s", newVersion)
}

func getCurrentVersion() string {
	stdout := &bytes.Buffer{}
	if err := execCommand(stdout, "git", "describe", "--tags", "--abbrev=0"); err != nil {
		log.Fatal(err)
	}
	return strings.TrimPrefix(strings.TrimSpace(stdout.String()), "v")
}

func getNewVersion() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(`Enter new version (without leading "v"): `)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read user input: %w", err))
	}
	version := strings.TrimSpace(input)

	if !versionRegexp.MatchString(version) {
		log.Fatal(fmt.Errorf("invalid version: %s", version))
	}
	return version
}

func checkRequirements(old string, new string) error {
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is not set. Required to generate release notes")
	}

	oldVersion, err := version.NewVersion(old)
	if err != nil {
		return fmt.Errorf("failed to parse current version: %w", err)
	}
	newVersion, err := version.NewVersion(new)
	if err != nil {
		return fmt.Errorf("failed to parse new version: %w", err)
	}
	if !newVersion.GreaterThan(oldVersion) {
		return fmt.Errorf("new version must be greater than current version")
	}

	if err := checkGitStatus(); err != nil {
		return fmt.Errorf("failed to check Git status: %w", err)
	}

	if err := checkGoModules(); err != nil {
		return fmt.Errorf("failed to check Go modules: %w", err)
	}
	return nil
}

func checkGitStatus() error {
	stdout := &bytes.Buffer{}
	if err := execCommand(stdout, "git", "status", "--porcelain"); err != nil {
		return err
	}
	if strings.TrimSpace(stdout.String()) != "" {
		return fmt.Errorf("the current working tree is dirty. Please commit or stash changes")
	}

	stdout = &bytes.Buffer{}
	if err := execCommand(stdout, "git", "rev-parse", "--abbrev-ref", "HEAD"); err != nil {
		return err
	}
	if strings.TrimSpace(stdout.String()) != "master" {
		return fmt.Errorf("the current branch is not master, got %s", strings.TrimSpace(stdout.String()))
	}

	stdout = &bytes.Buffer{}
	if err := execCommand(stdout, "git", "config", "--get", "remote.origin.url"); err != nil {
		return err
	}
	if !strings.Contains(strings.TrimSpace(stdout.String()), "terraform-linters/tflint") {
		return fmt.Errorf("remote.origin is not terraform-linters/tflint, got %s", strings.TrimSpace(stdout.String()))
	}
	return nil
}

func checkGoModules() error {
	bytes, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}
	content := string(bytes)

	matches := goModRequireSDKRegexp.FindStringSubmatch(content)
	if len(matches) != 2 {
		return fmt.Errorf(`failed to parse go.mod: did not match "%s"`, goModRequireSDKRegexp.String())
	}
	if !versionRegexp.MatchString(matches[1]) {
		return fmt.Errorf(`failed to parse go.mod: SDK version "%s" is not stable`, matches[1])
	}

	matches = goModRequireBundledPluginRegexp.FindStringSubmatch(content)
	if len(matches) != 2 {
		return fmt.Errorf(`failed to parse go.mod: did not match "%s"`, goModRequireBundledPluginRegexp.String())
	}
	if !versionRegexp.MatchString(matches[1]) {
		return fmt.Errorf(`failed to parse go.mod: bundled plugin version "%s" is not stable`, matches[1])
	}
	return nil
}

func rewriteFileWithNewVersion(path string, old string, new string) error {
	log.Printf("rewrite %s", path)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}
	content := string(bytes)

	replaced := strings.ReplaceAll(content, old, new)
	if replaced == content {
		return fmt.Errorf("%s is not changed", path)
	}

	if err := os.WriteFile(path, []byte(replaced), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func generateReleaseNote(old string, new string, savedPath string) error {
	tagName := fmt.Sprintf("v%s", new)
	previousTagName := fmt.Sprintf("v%s", old)
	targetCommitish := "master"

	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})))

	note, _, err := client.Repositories.GenerateReleaseNotes(
		context.Background(),
		"terraform-linters",
		"tflint",
		&github.GenerateNotesOptions{
			TagName:         tagName,
			PreviousTagName: &previousTagName,
			TargetCommitish: &targetCommitish,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %w", err)
	}

	if err := os.WriteFile(savedPath, []byte(note.Body), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", savedPath, err)
	}
	return err
}

func editFileInteractive(path string) error {
	editor := "vi"
	if e := os.Getenv("EDITOR"); e != "" {
		editor = e
	}
	return execShellCommand(os.Stdout, fmt.Sprintf("%s %s", editor, path))
}

func execShellCommand(stdout io.Writer, command string) error {
	shell := "sh"
	if s := os.Getenv("SHELL"); s != "" {
		shell = s
	}

	return execCommand(stdout, shell, "-c", command)
}

func execCommand(stdout io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		commands := append([]string{name}, args...)
		return fmt.Errorf(`failed to exec "%s": %w`, strings.Join(commands, " "), err)
	}
	return nil
}
