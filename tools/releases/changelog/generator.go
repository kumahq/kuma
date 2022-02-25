package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/iancoleman/orderedmap"
)

type Generator struct {
	log       *orderedmap.OrderedMap
	startTag  string
	endTag    string
	changelog string
}

func NewGenerator(startTag, endTag string) *Generator {
	return &Generator{
		log:      orderedmap.New(),
		startTag: startTag,
		endTag:   endTag,
	}
}

func (g *Generator) filteredTags() []string {
	tags := g.log.Keys()
	if g.endTag != "" {
		for _, tag := range tags {
			if g.formatTag(tag) == g.endTag {
				break
			}
			tags = tags[1:]
		}
	}

	if g.startTag != "" {
		for i, tag := range tags {
			if g.formatTag(tag) == g.startTag {
				tags = tags[:i]
				break
			}
		}
	}
	return tags
}

func (g *Generator) formatTitle(c *object.Commit) string {
	// Split the message, taking various new line formats
	splitMessage := strings.Split(strings.ReplaceAll(c.Message, "\r\n", "\n"), "\n")
	// The title is the first lien
	title := splitMessage[0]

	// generate a Markdown link to the pull request
	re := regexp.MustCompile(`\(#(?P<num>[0-9]*)\)`)
	mdPullString := fmt.Sprintf("[#$num](https://github.com/%s/pull/$num)", gitHubOrgProject())
	formattedTitle := re.ReplaceAllString(title, mdPullString)

	return formattedTitle
}

func (g *Generator) formatTag(tag string) string {
	return strings.ReplaceAll(tag, "refs/tags/", "")
}

func (g *Generator) formatTime(c *object.Commit) string {
	t := c.Author.When
	return fmt.Sprintf("%d/%02d/%02d", t.Year(), t.Month(), t.Day())
}

func (g *Generator) addToLog(tag string, c *object.Commit) {
	if _, found := g.log.Get(tag); !found {
		g.log.Set(tag, []*object.Commit{})
	}
	entry, _ := g.log.Get(tag)
	entry = append(entry.([]*object.Commit), c)
	g.log.Set(tag, entry)
}

func (g *Generator) addChangelog(add string) {
	g.changelog += add
}

func (g *Generator) Generate() {
	g.changelog = ""

	g.addChangelog("# CHANGELOG \n")

	tags := g.filteredTags()

	for _, tag := range tags {
		g.addChangelog(fmt.Sprintf("\n\n## [%s]\n", g.formatTag(tag)))

		value, _ := g.log.Get(tag)
		for i, c := range value.([]*object.Commit) {
			if i == 0 {
				g.addChangelog(fmt.Sprintln("> Released on ", g.formatTime(c)))
				g.addChangelog("\nChanges:\n")
			}
			g.addChangelog(fmt.Sprintln("*", g.formatTitle(c), "\n 👍contributed by", c.Author.Name))
		}
	}
}

func (g *Generator) Changelog() string {
	return g.changelog
}
