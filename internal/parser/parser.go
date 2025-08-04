package parser

import (
	"fmt"
	"regexp"
)

var ownerRepoStr = `([a-z\d](?:[a-z\d-]{0,38}[a-z\d])?)/([a-z\d_.-]+)`

var ownerRepoPattern = regexp.MustCompile(`(?i)^` + ownerRepoStr + `$`)
var ownerRepoTagPattern = regexp.MustCompile(`(?i)^` + ownerRepoStr + `(?:@(.+))?$`)
var githubUrlPattern = regexp.MustCompile(`(?i)^(?:https://github\.com/|git@github\.com:)` + ownerRepoStr + `(?:\.git)?$`)
var githubUrlPatternWithRelease = regexp.MustCompile(`(?i)^(?:https://github\.com/|git@github\.com:)` + ownerRepoStr + `(?:\.git)?(?:@(.+))?$`)

func ParseRepoRef(ref string) (owner string, repo string, err error) {
	if matches := ownerRepoPattern.FindStringSubmatch(ref); matches != nil {
		return matches[1], matches[2], nil
	}
	return "", "", fmt.Errorf("Cannot validate owner/repository link: %q", ref)
}

func ParseRepoReleaseRef(ref string) (owner string, repo string, release string, err error) {
	if matches := ownerRepoTagPattern.FindStringSubmatch(ref); matches != nil {
		return matches[1], matches[2], matches[3], nil
	}
	return "", "", "", fmt.Errorf("Cannot validate owner/repository link: %q", ref)
}

func ParseGithubUrlPattern(ref string) (owner string, repo string, err error) {
	if matches := githubUrlPattern.FindStringSubmatch(ref); matches != nil {
		return matches[2], matches[3], nil
	}
	return "", "", fmt.Errorf("Cannot validate owner/repository link: %q", ref)
}

func ParseGithubUrlPatternWithRelease(ref string) (owner string, repo string, release string, err error) {
	if matches := githubUrlPatternWithRelease.FindStringSubmatch(ref); matches != nil {
		return matches[2], matches[3], matches[4], nil
	}
	return "", "", "", fmt.Errorf("Cannot validate owner/repository link: %q", ref)
}
