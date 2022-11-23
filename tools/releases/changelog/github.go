package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var config struct {
	branch     string
	owner      string
	repo       string
	fromTag    string
	fromCommit string
	format     string
}

type OutFormat string

const (
	FormatMarkdown OutFormat = "md"
	FormatJson     OutFormat = "json"
)

var rootCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate the changelog.",
	Long:  `Generate the changelog.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must pass a subcommand")
	},
}

var autoChangelog = &cobra.Command{
	Use:   "changelog.md",
	Short: "Recreate the changelog.md using the changelog in each github release",
	Long: `
	We use whatever is after '## Changelog' to build the changelog 
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_API_TOKEN")
			if token == "" {
				return errors.New("need to set at least env GITHUB_TOKEN or GITHUB_API_TOKEN")
			}
		}
		gqlClient := GQLClient{Token: token}
		res, err := gqlClient.releaseGraphQL(config.owner, config.repo)
		if err != nil {
			return err
		}
		_, _ = cmd.OutOrStdout().Write([]byte("# Changelog\n<!-- Autogenerated with go run tools/releases/changelog.go changelog.md -->\n"))
		for _, release := range res {
			if strings.Contains(release.Description, "## Changelog") {
				changelog := strings.SplitN(release.Description, "## Changelog", 2)[1]
				_, _ = cmd.OutOrStdout().Write([]byte(fmt.Sprintf(`
## [%s]
> Released on %s%s
`, release.Name, release.CreatedAt.Format("2006/01/02"), changelog)))
			}

		}
		return nil
	},
}

var github = &cobra.Command{
	Use:   "github",
	Short: "Generate the changelog using the github graphql api",
	Long: `Generate the changelog using the github graphql api.
This will get all the commits in the branch after '--from-tag' or '--from-commit'
It will retrieve all the associated PRs to these commits and extract a changelog entry following these rules:

- If there's in the PR description an entry '> Changelog:'
	- If it's 'skip' --> This PR won't be listed in the changelog
	- Use this as the value for the changelog
- If the PR title starts with ci, test, refactor, build... skip the entry (if you still want it add a '> Changelog:' line in the PR description.
- Else use the PR title in the changelog

It will then output a changelog with all PRs with the same changelog grouped together
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.fromTag == "" && config.fromCommit == "" {
			return errors.New("You must set either --from-tag or --from-commit")
		}
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_API_TOKEN")
			if token == "" {
				return errors.New("need to set at least env GITHUB_TOKEN or GITHUB_API_TOKEN")
			}
		}
		gqlClient := GQLClient{Token: token}

		// Retrieve data from github
		commitLimit := config.fromCommit
		if config.fromTag != "" {
			res, err := gqlClient.commitByRef(config.owner, config.repo, config.fromTag)
			if err != nil {
				return err
			}
			commitLimit = res
		}
		byChangelog := map[string][]*CommitInfo{}
		// Deal with pagination
		res, err := gqlClient.historyGraphQl(config.owner, config.repo, config.branch, commitLimit)
		if err != nil {
			return err
		}

		// Rollup changes together
		for i := range res {
			ci := NewCommitInfo(res[i])
			if ci == nil {
				continue
			}
			byChangelog[ci.Changelog] = append(byChangelog[ci.Changelog], ci)
		}
		// Create a list to display
		var out []Changelog
		for changelog, commits := range byChangelog {
			uniqueAuthors := map[string]interface{}{}
			var authors []string
			var prs []int
			var minVersion, maxVersion string
			for _, c := range commits {
				if minVersion == "" || c.MinDependency < minVersion {
					minVersion = c.MinDependency
				}
				if maxVersion == "" || c.MaxDependency > maxVersion {
					maxVersion = c.MaxDependency
				}
				prs = append(prs, c.PrNumber)
				if _, exists := uniqueAuthors[c.Author]; !exists {
					authors = append(authors, fmt.Sprintf("@%s", c.Author))
					uniqueAuthors[c.Author] = nil
				}
			}
			sort.Ints(prs)
			sort.Strings(authors)
			if minVersion != "" && maxVersion != "" {
				changelog = fmt.Sprintf("%s from %s to %s", changelog, minVersion, maxVersion)
			}
			out = append(out, Changelog{Desc: changelog, Authors: authors, PullRequests: prs})
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i].Desc < out[j].Desc
		})
		switch OutFormat(config.format) {
		case FormatMarkdown:
			for _, v := range out {
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "* %s\n", v)
				if err != nil {
					return err
				}
			}
		case FormatJson:
			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(out)
		}
		return nil
	},
}

type Changelog struct {
	Desc         string   `json:"desc"`
	Authors      []string `json:"authors"`
	PullRequests []int    `json:"pull_requests"`
}

func (c Changelog) String() string {
	var prLinks []string
	for _, n := range c.PullRequests {
		prLinks = append(prLinks, fmt.Sprintf("[#%d](https://github.com/%s/%s/pull/%d)", n, config.owner, config.repo, n))
	}
	seen := map[string]struct{}{}
	var authors []string
	for _, a := range c.Authors {
		if _, ok := seen[a]; !ok {
			authors = append(authors, a)
			seen[a] = struct{}{}
		}
	}
	sort.Strings(authors)
	return fmt.Sprintf("%s %s %s", c.Desc, strings.Join(prLinks, " "), strings.Join(authors, ","))
}

type CommitInfo struct {
	Sha           string
	Author        string
	PrNumber      int
	PrTitle       string
	Changelog     string
	MinDependency string
	MaxDependency string
}

// titles look like: chore(deps): bump github.com/lib/pq from 1.10.6 to 1.10.7
var dependabotPRTitleRegExp = regexp.MustCompile(`(chore\(deps\): bump [^ ]+) from ([^ ]+) to ([^ ]+)`)

func NewCommitInfo(commit GQLCommit) *CommitInfo {
	if len(commit.AssociatedPullRequests.Nodes) == 0 {
		return nil
	}
	pr := commit.AssociatedPullRequests.Nodes[0]
	res := &CommitInfo{
		Author:   pr.Author.Login,
		Sha:      commit.Oid,
		PrNumber: pr.Number,
		PrTitle:  pr.Title,
	}
	changelog := ""
	for _, l := range strings.Split(pr.Body, "\n") {
		if strings.HasPrefix(l, "> Changelog: ") {
			changelog = strings.TrimSpace(strings.TrimPrefix(l, "> Changelog: "))
		}
	}
	switch changelog {
	case "skip":
		return nil
	case "":
		// Ignore prs with usually ignored prefix
		for _, v := range []string{"build:", "ci:", "ci(", "test(", "refactor(", "chore(ci)", "fix(ci)", "fix(test)", "tests(", "build(", "docs(madr)"} {
			if strings.HasPrefix(commit.Message, v) {
				return nil
			}
		}
		matches := dependabotPRTitleRegExp.FindStringSubmatch(pr.Title)
		// Rollup dependabot issues with the same dependency into just one so we can rebuild a single line with all update PRs.
		if matches != nil {
			res.Changelog = matches[1]
			res.MinDependency = matches[2]
			res.MaxDependency = matches[3]
		} else {
			// Use the pr.Title as a changelog entry
			res.Changelog = pr.Title
		}
	default:
		res.Changelog = changelog
	}
	return res
}

func init() {
	github.Flags().StringVar(&config.owner, "owner", "kumahq", "The owner org to query")
	github.Flags().StringVar(&config.repo, "name", "kuma", "The repository to query")
	autoChangelog.Flags().StringVar(&config.owner, "owner", "kumahq", "The owner org to query")
	autoChangelog.Flags().StringVar(&config.repo, "name", "kuma", "The repository to query")
	github.Flags().StringVar(&config.branch, "branch", "master", "The branch to look for the start on")
	github.Flags().StringVar(&config.fromCommit, "from-commit", "", "If set only show commits after this commit sha")
	github.Flags().StringVar(&config.fromTag, "from-tag", "", "If set only show commits after this tag (must be on the same branch)")
	github.Flags().StringVar(&config.format, "format", string(FormatMarkdown), fmt.Sprintf("The output format (%s, %s)", FormatJson, FormatMarkdown))
	rootCmd.AddCommand(github)
	rootCmd.AddCommand(autoChangelog)
}
