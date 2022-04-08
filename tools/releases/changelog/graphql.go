package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type GQLOutput struct {
	Data GQLData `json:"data"`
}
type GQLData struct {
	Repository GQLRepo `json:"repository"`
}

type GQLRepo struct {
	Ref    GQLRef        `json:"ref"`
	Object GQLObjectRepo `json:"object"`
}

type GQLObjectRepo struct {
	History GQLHistoryRepo `json:"history"`
}

type GQLHistoryRepo struct {
	Nodes []GQLCommit `json:"nodes"`
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
	CommitUrl string    `json:"commitUrl"`
	Name      string    `json:"name"`
	Target    GQLCommit `json:"target"`
}

type GQLClient struct {
	Token string
}

func (c GQLClient) historyGraphQl(owner, name, branch string, since time.Time) ([]GQLCommit, error) {
	res, err := c.graphqlQuery(`
query($name: String!, $owner: String!, $branch: String!, $since: GitTimestamp!) {
  repository(owner: $owner, name: $name) {
    object(expression: $branch) {
      ... on Commit {
        history(since: $since) {
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
`, map[string]interface{}{"owner": owner, "name": name, "branch": branch, "since": since.Format(time.RFC3339)})
	if err != nil {
		return nil, err
	}
	return res.Data.Repository.Object.History.Nodes, nil
}

func (c GQLClient) commitByRef(owner, name, tag string) (GQLCommit, error) {
	res, err := c.graphqlQuery(`
query ($owner: String!, $name: String!, $ref: String!) {
  repository(name: $name, owner: $owner) {
    ref(qualifiedName: $ref) {
      target {
        commitUrl
        ... on Tag {
          name
          target {
            ... on Commit {
              oid
              message
            }
          }
        }
      }
    }
  }
}
`, map[string]interface{}{"owner": owner, "name": name, "ref": fmt.Sprintf("refs/tags/%s", tag)})
	if err != nil {
		return GQLCommit{}, err
	}
	return res.Data.Repository.Ref.Target.Target, nil
}

func (c GQLClient) graphqlQuery(query string, variables map[string]interface{}) (out GQLOutput, err error) {
	b2 := bytes.Buffer{}
	err = json.NewEncoder(&b2).Encode(map[string]interface{}{"query": query, "variables": variables})
	if err != nil {
		return
	}
	var r *http.Request
	r, err = http.NewRequest(http.MethodPost, "https://api.github.com/graphql", &b2)
	if err != nil {
		return
	}
	r.Header.Set("Authorization", fmt.Sprintf("bearer %s", c.Token))
	r.Header.Set("Content-Type", "application/json")
	var res *http.Response
	res, err = http.DefaultClient.Do(r)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := ioutil.ReadAll(res.Body)
		err = fmt.Errorf("got status: %d body:%s", res.StatusCode, b)
		return
	}
	err = json.NewDecoder(res.Body).Decode(&out)
	return
}
