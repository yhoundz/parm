package cmdparser

import (
	"fmt"
	"regexp"
	"strings"
)

var ownerRepoStr = `([a-z\d](?:[a-z\d-]{0,38}[a-z\d])*)/([a-z\d_.-]+?)`
var ownerRepoPattern = regexp.MustCompile(`(?i)^` + ownerRepoStr + `$`)
var ownerRepoTagPattern = regexp.MustCompile(`(?i)^` + ownerRepoStr + `(?:@(.+))?$`)
var githubUrlPattern = regexp.MustCompile(`(?i)^(?:https://github\.com/|git@github\.com:)` + ownerRepoStr + `(?:\.git)?$`)
var githubUrlPatternWithRelease = regexp.MustCompile(`(?i)^(?:https://github\.com/|git@github\.com:)` + ownerRepoStr + `(?:\.git)?(?:@(.+))?$`)

// Shorthand pattern for single name without slash (owner == repo)
var shorthandPattern = regexp.MustCompile(`(?i)^([a-z\d](?:[a-z\d-]{0,38}[a-z\d])*)$`)
var shorthandTagPattern = regexp.MustCompile(`(?i)^([a-z\d](?:[a-z\d-]{0,38}[a-z\d])*)@(.+)$`)

// general purpose
func ParseRepoRef(ref string) (owner string, repo string, err error) {
	// Try shorthand syntax first: "neovim" -> "neovim/neovim"
	if matches := shorthandPattern.FindStringSubmatch(ref); matches != nil {
		return matches[1], matches[1], nil
	}
	if matches := ownerRepoPattern.FindStringSubmatch(ref); matches != nil {
		return matches[1], matches[2], nil
	}
	return "", "", fmt.Errorf("cannot validate owner/repository link: %q", ref)
}

// specifically parsing tag args
func ParseRepoReleaseRef(ref string) (owner string, repo string, release string, err error) {
	// Try shorthand syntax with tag: "neovim@v0.11.0" -> "neovim/neovim@v0.11.0"
	if matches := shorthandTagPattern.FindStringSubmatch(ref); matches != nil {
		return matches[1], matches[1], matches[2], nil
	}
	if matches := ownerRepoTagPattern.FindStringSubmatch(ref); matches != nil {
		return matches[1], matches[2], matches[3], nil
	}
	return "", "", "", fmt.Errorf("cannot validate owner/repository link: %q", ref)
}

// general purpose
func ParseGithubUrlPattern(ref string) (owner string, repo string, err error) {
	if matches := githubUrlPattern.FindStringSubmatch(ref); matches != nil {
		owner, repo := matches[1], matches[2]
		if repo != ".git" {
			return owner, repo, nil
		}
	}
	return "", "", fmt.Errorf("cannot validate owner/repository link: %q", ref)
}

// specifically parsing tag args
func ParseGithubUrlPatternWithRelease(ref string) (owner string, repo string, release string, err error) {
	if matches := githubUrlPatternWithRelease.FindStringSubmatch(ref); matches != nil {
		owner, repo, tag := matches[1], matches[2], matches[3]
		if repo != ".git" {
			return owner, repo, tag, nil
		}
	}
	return "", "", "", fmt.Errorf("cannot validate owner/repository link: %q", ref)
}

func BuildGitLink(owner string, repo string) (httpsLink string, sshLink string) {
	httpCloneLink := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	sshCloneLink := fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	return httpCloneLink, sshCloneLink
}

func StringToString(in string) (string, string, error) {
	parts := strings.SplitN(in, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid argument format: %q. Expected key=value", in)
	}
	return parts[0], parts[1], nil
}
