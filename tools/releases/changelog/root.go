package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
)

var config struct {
	gitHubRepo string
	branch     string
	startTag   string
	endTag     string
	owner      string
	repo       string
	since      time.Duration
	fromTag    string
	fromCommit string
}

func gitHubOrgProject() string {
	u, _ := url.Parse(config.gitHubRepo)
	return u.Path[:len(u.Path)-4]
}

var rootCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate the changelog.",
	Long:  `Generate the changelog.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Clones the given repository, creating the remote, the local branches
		// and fetching the objects, everything in memory:
		Info("Clone %s branch %s", config.gitHubRepo, plumbing.ReferenceName(config.branch))
		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:           config.gitHubRepo,
			ReferenceName: plumbing.ReferenceName(config.branch),
			SingleBranch:  false,
		})
		CheckIfError(err)

		Info("Retrieve tags")
		tagMap := map[string]string{}
		// List all tag references, both lightweight tags and annotated tags
		tagrefs, err := r.Tags()
		CheckIfError(err)
		err = tagrefs.ForEach(func(t *plumbing.Reference) error {
			tagMap[t.Hash().String()] = t.Name().String()
			return nil
		})
		CheckIfError(err)

		Info("Retrieve commit history")
		// ... retrieves the commit history
		refHash := plumbing.Hash{}
		startRef, err := r.Reference(plumbing.ReferenceName(config.startTag), false)
		if err != nil {
			Info("error: %s", err)
		} else {
			refHash = startRef.Hash()
		}
		cIter, err := r.Log(&git.LogOptions{
			From:  refHash,
			Order: git.LogOrderCommitterTime,
		})
		CheckIfError(err)

		Info("Filter the commits by tag")
		generator := NewGenerator(config.startTag, config.endTag)
		currentTag := config.branch
		err = cIter.ForEach(func(c *object.Commit) error {
			if tag, found := tagMap[c.Hash.String()]; found {
				currentTag = tag
			}
			generator.addToLog(currentTag, c)
			return nil
		})
		CheckIfError(err)

		Info("Generate the formatted log")
		generator.Generate()
		fmt.Println(generator.Changelog())
	},
}

var graphqlCmd = &cobra.Command{
	Use:   "graphql",
	Short: "Generate the changelog using graphql api",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_API_TOKEN")
			if token == "" {
				return errors.New("need to set at least env GITHUB_TOKEN or GITHUB_API_TOKEN")
			}
		}
		since := time.Now().Add(-config.since)
		gqlClient := GQLClient{Token: token}

		// Retrieve data from github
		commitLimit := config.fromCommit
		if config.fromTag != "" {
			res, err := gqlClient.commitByRef(config.owner, config.repo, config.fromTag)
			if err != nil {
				return err
			}
			commitLimit = res.Oid
		}
		res, err := gqlClient.historyGraphQl(config.owner, config.repo, config.branch, since)
		if err != nil {
			return err
		}

		// Rollup changes together
		byChangelog := map[string]*Changelog{}
		for i := range res {
			if strings.HasPrefix(commitLimit, res[i].Oid) {
				break
			}
			ci := NewCommitInfo(res[i])
			if ci.ShouldIgnore() {
				continue
			}
			cl := ci.Changelog
			if ci.Changelog == "" {
				cl = ci.PrTitle
			}
			c := byChangelog[cl]
			if c == nil {
				c = &Changelog{
					Desc:         cl,
					PullRequests: []int{},
				}
				byChangelog[cl] = c
			}
			c.PullRequests = append(c.PullRequests, ci.PrNumber)
			c.Authors = append(c.Authors, "@"+ci.Author)
		}
		// Create a list to display
		var out []string
		for _, l := range byChangelog {
			out = append(out, l.String())
		}
		sort.Strings(out)
		fmt.Fprintf(cmd.OutOrStdout(), "%s", strings.Join(out, "\n"))
		return nil
	},
}

type Changelog struct {
	Desc         string
	Authors      []string
	PullRequests []int
}

func (c Changelog) String() string {
	var prLinks []string
	for _, n := range c.PullRequests {
		prLinks = append(prLinks, fmt.Sprintf("[%d](https://github.com/%s/%s/pull/%d)", n, config.owner, config.repo, n))
	}
	seen := map[string]interface{}{}
	var authors []string
	for _, a := range c.Authors {
		if _, ok := seen[a]; !ok {
			authors = append(authors, a)
			seen[a] = nil
		}
	}
	sort.Strings(authors)
	return fmt.Sprintf("%s %s %s", c.Desc, strings.Join(prLinks, " "), strings.Join(authors, ","))
}

type CommitInfo struct {
	Sha       string
	Author    string
	PrNumber  int
	PrTitle   string
	Changelog string
}

func (ci CommitInfo) ShouldIgnore() bool {
	if ci.Changelog != "" {
		return ci.Changelog == "skip"
	}
	for _, v := range []string{"ci(", "test(", "refactor(", "fix(ci)", "fix(test)", "tests("} {
		if strings.HasPrefix(ci.PrTitle, v) {
			return true
		}
	}
	return false
}

func NewCommitInfo(commit GQLCommit) CommitInfo {
	pr := commit.AssociatedPullRequests.Nodes[0]
	changelog := ""
	for _, l := range strings.Split(pr.Body, "\n") {
		if strings.HasPrefix(l, "> Changelog: ") {
			changelog = strings.TrimSpace(strings.TrimPrefix(l, "> Changelog: "))
		}
	}
	// TODO extract changelog string
	return CommitInfo{
		Author:    pr.Author.Login,
		Sha:       commit.Oid,
		PrNumber:  pr.Number,
		PrTitle:   pr.Title,
		Changelog: changelog,
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Older stuff
	rootCmd.Flags().StringVar(&config.gitHubRepo, "repo", "https://github.com/kumahq/kuma.git", "The GitHub repo to process")
	rootCmd.Flags().StringVar(&config.branch, "branch", "master", "The branch to process")
	rootCmd.Flags().StringVar(&config.startTag, "start", "0.7.1", "The start hash or tag")
	rootCmd.Flags().StringVar(&config.endTag, "end", "", "The end hash or tag")

	graphqlCmd.Flags().StringVar(&config.owner, "owner", "kumahq", "The owner")
	graphqlCmd.Flags().StringVar(&config.repo, "name", "kuma", "The repo")
	graphqlCmd.Flags().StringVar(&config.branch, "branch", "master", "The branch to look for the start on")
	graphqlCmd.Flags().DurationVar(&config.since, "since", time.Hour*24*90, "When to get the data from can either be a timestamp of a tag (90 days ago)")
	graphqlCmd.Flags().StringVar(&config.fromCommit, "from-commit", "", "If set only show commits after this commit sha")
	graphqlCmd.Flags().StringVar(&config.fromTag, "from-tag", "", "If set only show commits after this tag (must be on the same branch)")
	rootCmd.AddCommand(graphqlCmd)
}
