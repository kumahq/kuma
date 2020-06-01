package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

var config struct {
	gitHubRepo string
	startTag   string
	endTag     string
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
		Info("Clone %s", config.gitHubRepo)
		r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL: config.gitHubRepo,
		})
		CheckIfError(err)

		Info("Retreive tags")
		tagMap := map[string]string{}
		// List all tag references, both lightweight tags and annotated tags
		tagrefs, err := r.Tags()
		CheckIfError(err)
		err = tagrefs.ForEach(func(t *plumbing.Reference) error {
			tagMap[t.Hash().String()] = t.Name().String()
			return nil
		})
		CheckIfError(err)

		Info("Retreive commit history")
		// ... retrieves the commit history
		cIter, err := r.Log(&git.LogOptions{})
		CheckIfError(err)

		Info("Filter the commits by tag")
		generator := NewGenerator(config.startTag, config.endTag)
		currentTag := "master"
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&config.gitHubRepo, "repo", "https://github.com/Kong/kuma.git", "The GitHub repo to process")
	rootCmd.Flags().StringVar(&config.startTag, "start", "", "The start hash or tag")
	rootCmd.Flags().StringVar(&config.endTag, "end", "", "The end hash or tag")
}
