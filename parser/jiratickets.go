package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type JiraTicketRef struct {
	jiraTicketURL string
	regex         *regexp.Regexp
}

func NewJiraTicketRef(jiraTicketURL string, projectKeys ...string) (*JiraTicketRef, error) {
	var re string

	for _, project := range projectKeys {
		if len(re) != 0 {
			re += "|"
		}

		re += strings.ToUpper(project) + `-[0-9]+`
	}

	rec, err := regexp.Compile(re)
	if err != nil {
		return nil, fmt.Errorf("compiling regex '%s' failed, %w", re, err)
	}

	return &JiraTicketRef{
		jiraTicketURL: jiraTicketURL,
		regex:         rec,
	}, nil
}

func (j *JiraTicketRef) Parse(in string) []string {
	ticketIDs := dedupStringSlice(j.regex.FindAllString(in, -1))

	result := make([]string, 0, len(ticketIDs))

	for _, id := range ticketIDs {
		result = append(result, j.jiraTicketURL+"/"+id)
	}

	return result
}
