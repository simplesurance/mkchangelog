package parser

import (
	"regexp"
)

type GithubPRRef struct {
	pullRequestURL string
}

func NewGithubPRRef(basePullRequestURL string) *GithubPRRef {
	return &GithubPRRef{
		pullRequestURL: basePullRequestURL,
	}
}

var githubPRRegex = regexp.MustCompile(`(?:\(#)([0-9]+)(?:\))`)

func (g *GithubPRRef) Parse(in string) []string {
	var result []string

	matches := githubPRRegex.FindAllStringSubmatch(in, -1)
	result = make([]string, 0, len(matches))

	for _, m := range matches {
		if len(m) != 2 {
			continue
		}

		result = append(result, g.pullRequestURL+"/"+m[1])
	}

	return dedupStringSlice(result)
}
