package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GQLOutput struct {
	Data GQLData `json:"data"`
}
type GQLData struct {
	Repository GQLRepo `json:"repository"`
}

type GQLPageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type GQLRepo struct {
	Ref      GQLRef           `json:"ref"`
	Object   GQLObjectRepo    `json:"object"`
	Releases GQLObjectRelease `json:"releases"`
}

type GQLObjectRelease struct {
	Nodes []GQLRelease `json:"nodes"`
}

type GQLRelease struct {
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	IsDraft      bool      `json:"isDraft"`
	IsPrerelease bool      `json:"isPrerelease"`
	Description  string    `json:"description"`
}

type GQLObjectRepo struct {
	History GQLHistoryRepo `json:"history"`
}

type GQLHistoryRepo struct {
	PageInfo GQLPageInfo `json:"pageInfo"`
	Nodes    []GQLCommit `json:"nodes"`
}

type GQLAssociatedPRs struct {
	Nodes []GQLPRNode `json:"nodes"`
}

type GQLAuthor struct {
	Login string `json:"login"`
}

type GQLPRNode struct {
	Author GQLAuthor `json:"author"`
	Number int       `json:"number"`
	Title  string    `json:"title"`
	Body   string    `json:"body"`
}

type GQLRef struct {
	Target GQLRefTarget `json:"target"`
}

type GQLCommit struct {
	Oid                    string           `json:"oid"`
	Message                string           `json:"message"`
	AssociatedPullRequests GQLAssociatedPRs `json:"associatedPullRequests"`
}

type GQLRefTarget struct {
	CommitUrl string `json:"commitUrl"`
	Oid       string `json:"oid"`
}

type GQLClient struct {
	Token string
}

func splitRepo(repo string) (string, string) {
	r := strings.Split(repo, "/")
	return r[0], r[1]
}

func (c GQLClient) releaseGraphQL(repo string) ([]GQLRelease, error) {
	owner, name := splitRepo(repo)
	res, err := c.graphqlQuery(`
query($name: String!, $owner: String!) {
  repository(owner: $owner, name: $name) {
    releases(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {
      nodes {
        name
        createdAt
        isDraft
        isPrerelease
        description
      }
    }
  }
}
`, map[string]interface{}{"owner": owner, "name": name})
	if err != nil {
		return nil, err
	}
	return res.Data.Repository.Releases.Nodes, nil
}

func (c GQLClient) historyGraphQl(repo, branch, commitLimit string) ([]GQLCommit, error) {
	owner, name := splitRepo(repo)
	var out []GQLCommit
	var err error
	var res GQLOutput
	for {
		cursorStr := ""
		if res.Data.Repository.Object.History.PageInfo.EndCursor != "" {
			cursorStr = fmt.Sprintf(`(after: "%s")`, res.Data.Repository.Object.History.PageInfo.EndCursor)
		}
		res, err = c.graphqlQuery(fmt.Sprintf(`
query($name: String!, $owner: String!, $branch: String!) {
  repository(owner: $owner, name: $name) {
    object(expression: $branch) {
      ... on Commit {
        history%s {
          pageInfo {
            hasNextPage
            endCursor
          }
          nodes {
            oid
            message
            associatedPullRequests(first: 1) {
              nodes {
                author {
                  login
                }
                number
                title
                body
              }
            }
          }
        }
      }
    }
  }
}
`, cursorStr), map[string]interface{}{"owner": owner, "name": name, "branch": branch})
		if err != nil {
			return out, err
		}
		for _, r := range res.Data.Repository.Object.History.Nodes {
			if commitLimit != "" && strings.HasPrefix(r.Oid, commitLimit) {
				return out, err
			}
			out = append(out, r)
		}
		if !res.Data.Repository.Object.History.PageInfo.HasNextPage {
			return out, err
		}
	}
}

func (c GQLClient) commitByRef(repo, tag string) (string, error) {
	owner, name := splitRepo(repo)
	res, err := c.graphqlQuery(`
query ($owner: String!, $name: String!, $ref: String!) {
  repository(name: $name, owner: $owner) {
    ref(qualifiedName: $ref) {
      target {
        commitUrl
        oid
      }
    }
  }
}
`, map[string]interface{}{"owner": owner, "name": name, "ref": fmt.Sprintf("refs/tags/%s", tag)})
	if err != nil {
		return "", err
	}
	return res.Data.Repository.Ref.Target.Oid, nil
}

func (c GQLClient) graphqlQuery(query string, variables map[string]interface{}) (GQLOutput, error) {
	var out GQLOutput
	var err error
	b2 := bytes.Buffer{}
	err = json.NewEncoder(&b2).Encode(map[string]interface{}{"query": query, "variables": variables})
	if err != nil {
		return out, err
	}
	var r *http.Request
	r, err = http.NewRequest(http.MethodPost, "https://api.github.com/graphql", &b2)
	if err != nil {
		return out, err
	}
	r.Header.Set("Authorization", fmt.Sprintf("bearer %s", c.Token))
	r.Header.Set("Content-Type", "application/json")
	var res *http.Response
	res, err = http.DefaultClient.Do(r)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		err = fmt.Errorf("got status: %d body:%s", res.StatusCode, b)
		return out, err
	}
	err = json.NewDecoder(res.Body).Decode(&out)
	return out, err
}
