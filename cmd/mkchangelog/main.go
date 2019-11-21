package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	flag "github.com/spf13/pflag"

	"github.com/simplesurance/mkchangelog/changelog"
	"github.com/simplesurance/mkchangelog/format/markdown"
	"github.com/simplesurance/mkchangelog/parser"
)

func exitOnError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %+v\n", msg, err)
		os.Exit(1)
	}
}

var buildCommit = "undefined-commit"
var buildVersion = "undefined-version"

var (
	githubPRURLFlag     *string
	jiraTicketURLFlag   *string
	jiraProjectKeysFlag *string
	jiraProjectKeys     []string
	showUsageFlag       *bool
	showVersionFlag     *bool
	repositoryPathFlag  *string
	releaseNameFlag     *string

	fromRevArg string
	toRevArg   string
)

func registerFlags() {
	githubPRURLFlag = flag.StringP(
		"github-pull-request-url",
		"g",
		"https://github.com/simplesurance/sisu/pull",
		"URL to the Github Pull-Request Base URL for the repository.\n"+
			"This is normally https://github.com/<ORGANIZATION>/<REPOSITORY>/pull.\n"+
			"If set, links to the Pull-Request for that repository will be added for (#<NUMBER>) occurrences in commit messages.",
	)

	jiraTicketURLFlag = flag.StringP(
		"jira-ticket-url",
		"j",
		"https://sisu-agile.atlassian.net/browse",
		"Base ticket URL of a IRA instance.\n"+
			"This is normally https://<DOMAIN>/browse\n"+
			"If set, links to JIRA tickets will be added for <PROJECT>-<NUMBER> occurrences in commit messages.",
	)

	jiraProjectKeysFlag = flag.StringP(
		"jira-project-keys",
		"k",
		"PLAT,CORE",
		"Comma-separated list of Jira Project Keys that are recognized\n"+
			"when searching for Jira Ticket references in commit messages.\n"+
			"The keys are case-sensitive.",
	)

	repositoryPathFlag = flag.StringP(
		"repository-path",
		"r",
		".",
		"Path to the cloned Git-Repository for that the Changelog is created.",
	)

	releaseNameFlag = flag.StringP(
		"release-name",
		"n",
		"UNDEFINED-VERSION",
		"Name of the release that will show up in the changelog.",
	)

	showUsageFlag = flag.BoolP("help", "h", false, "display this help and exit")
	showVersionFlag = flag.BoolP("version", "v", false, "display the version and exit")
}

func printHelpErr() {
	fmt.Fprintf(os.Stderr, "Try %s --help for more information.\n", path.Base(os.Args[0]))
}

func mustParseFlags() {
	cmdname := path.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [<Parameters>]... <FROM-REV> [<TO-REV>]\n", cmdname)
		fmt.Fprintf(os.Stderr, "Create a Changelog from Git log messages.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "If <TO-REV> is omitted, HEAD will be used as TO-REV.\n\n")
		fmt.Fprintf(os.Stderr, "Parameters:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showUsageFlag {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersionFlag {
		fmt.Printf("%s %s (%s)\n", cmdname, buildVersion, buildCommit)
		os.Exit(0)
	}

	switch flag.NArg() {
	case 1:
		fromRevArg = flag.Arg(0)
		toRevArg = "HEAD"

	case 2:
		fromRevArg = flag.Arg(0)
		toRevArg = flag.Arg(1)

	default:
		fmt.Fprintf(os.Stderr, "%s: invalid arguments\n", cmdname)
		printHelpErr()
		os.Exit(1)
	}

	if *jiraProjectKeysFlag != "" {
		jiraProjectKeys = strings.Split(*jiraProjectKeysFlag, ",")
	}
}

func changelogToMarkdownDoc(log *changelog.Changelog) *markdown.Document {
	doc := markdown.Document{}

	doc.AppendElement(&markdown.Heading{
		Level: markdown.HeadingLvl1,
		Text:  "Release " + *releaseNameFlag,
	})

	list := markdown.List{}

	for _, entry := range log.Entries {
		list.Entries = append(list.Entries, &markdown.ListItem{
			Lvl:     markdown.ListItemLevel1,
			Content: &markdown.Text{Text: entry.Headline},
		})

		for _, desc := range entry.Description {
			list.Entries = append(list.Entries, &markdown.ListItem{
				Lvl:     markdown.ListItemLevel2,
				Content: &markdown.Link{URL: desc},
			})
		}

	}

	doc.AppendElement(&list)

	return &doc
}

func main() {
	registerFlags()
	mustParseFlags()

	repo, err := git.PlainOpen(*repositoryPathFlag)
	exitOnError(err, "reading git repository failed")

	fromHash, err := repo.ResolveRevision(plumbing.Revision(fromRevArg))
	exitOnError(err, fmt.Sprintf("invalid FROM revision argument '%s'", fromRevArg))

	toHash, err := repo.ResolveRevision(plumbing.Revision(toRevArg))
	exitOnError(err, fmt.Sprintf("invalid TO revision argument '%s'", toRevArg))

	it, err := repo.Log(&git.LogOptions{
		From:  *toHash,
		Order: git.LogOrderCommitterTime,
	})
	exitOnError(err, "reading git repository log failed")

	var descriptionCreators []changelog.DescriptionCreator

	if len(jiraProjectKeys) != 0 && *jiraTicketURLFlag != "" {
		ticketRefparser, err := parser.NewJiraTicketRef(*jiraTicketURLFlag, jiraProjectKeys...)
		exitOnError(err, "initializing Jira Ticket Reference Parser failed")

		descriptionCreators = append(descriptionCreators, ticketRefparser)
	}

	if *githubPRURLFlag != "" {
		githubPRparser := parser.NewGithubPRRef(*githubPRURLFlag)
		exitOnError(err, "initializing GitHub Pull-Request Reference Parser failed")
		descriptionCreators = append(descriptionCreators, githubPRparser)
	}

	log := changelog.Changelog{
		ReleaseName:         *releaseNameFlag,
		DescriptionCreators: descriptionCreators,
	}

	// TODO: skip merge commit descriptions?

	err = it.ForEach(func(commit *object.Commit) error {
		if commit.Hash == *fromHash {
			return storer.ErrStop
		}

		log.Parse(commit.Message)

		return nil
	})
	exitOnError(err, "retrieving commits failed")

	markdownDoc := changelogToMarkdownDoc(&log)

	fmt.Println(markdownDoc.String())
}
