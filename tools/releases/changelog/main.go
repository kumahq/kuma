package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

func main() {
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
}
