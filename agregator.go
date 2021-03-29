package aggregator

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/google/go-github/v34/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

type CaddyAggregator struct {
	configuration CaddyAggregatorConfiguration
	store         *ristretto.Cache
	restClient    *github.Client
	graphClient   *githubv4.Client
	clientCTX     context.Context
}

func (c *CaddyAggregator) initializeClient() {
	c.clientCTX = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.configuration.Pat},
	)
	baseClient := oauth2.NewClient(c.clientCTX, ts)
	c.restClient = github.NewClient(baseClient)
	c.graphClient = githubv4.NewClient(baseClient)
}

// It will load the configuration to get the Personal Access Token
func (c *CaddyAggregator) LoadConfiguration(channel chan int) {
	_ = c.configuration.ParseConfiguration()
	store, _ := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
		OnEvict: func(key uint64, u2 uint64, i interface{}, i2 int64) {
			channel <- 1
		},
	})

	c.store = store
	c.initializeClient()
}

func (c *CaddyAggregator) hasChanged(module CaddyModule) bool {
	if module.Version != "" && module.Hash != "" {
		key := module.Repository
		hash, _ := c.store.Get(key)

		if hash != module.Hash {
			c.store.Set(key, module.Hash, 1)
			return true
		}
	}

	return false
}

func (c *CaddyAggregator) redeploy(data []byte) {
	err := ioutil.WriteFile(c.configuration.FilePath, data, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *CaddyAggregator) pushToGit(data []byte) {
	commitOption := &github.RepositoryContentFileOptions{
		Branch:  github.String("master"),
		Message: github.String("Update from xcaddy-builder-aggregator"),
		Committer: &github.CommitAuthor{
			Name:  github.String(c.configuration.Owner),
			Email: github.String(c.configuration.Email),
		},
		Author: &github.CommitAuthor{
			Name:  github.String(c.configuration.Owner),
			Email: github.String(c.configuration.Email),
		},

		Content: data,
	}

	x, _, err := c.restClient.Repositories.UpdateFile(c.clientCTX, c.configuration.Owner, c.configuration.Repository, c.configuration.Path, commitOption)
	if err != nil {
		panic(err)
	}

	fmt.Println(x.SHA)
	fmt.Println(x)
}

func (c *CaddyAggregator) fetchCaddyfile(module CaddyModule, path string) string {
	if path != "" {
		f, _, e := c.restClient.Repositories.DownloadContents(c.clientCTX, module.Author, module.Name, path, nil)
		if e != nil {
			panic(e)
		}
		buf := new(bytes.Buffer)
		_, e = buf.ReadFrom(f)
		return buf.String()
	}

	return ""
}

func (c *CaddyAggregator) parseData(rr RepositoryRetriever) {
	var modules []CaddyModule
	e := Parse(c.configuration.FilePath, &modules)
	for _, m := range modules {
		_ = c.hasChanged(m)
	}
	time.Sleep(1 * time.Second)
	modules = []CaddyModule{}
	if e != nil {
		panic(e)
	}
	hasUpdated := false
	for _, edge := range rr.Search.Edges {
		r := edge.Node.Repository
		names := strings.Split(string(r.NameWithOwner), "/")
		s := strings.Split(string(r.LatestRelease.TagCommit.CommitUrl), "/")
		tags := []string{}
		for _, tag := range r.RepositoryTopics.Nodes {
			tags = append(tags, string(tag.Topic.Name))
		}

		commitHash := s[len(s)-1]
		module := CaddyModule{
			Author:      names[0],
			Config:      "",
			Name:        names[1],
			Description: string(r.Description),
			Hash:        commitHash,
			Repository:  fmt.Sprintf("github.com/%s", string(r.NameWithOwner)),
			Tags:        tags,
			Version:     string(r.LatestRelease.TagName),
		}

		cf := module.retrieveCaddyfile(r.LatestRelease)
		if c.hasChanged(module) {
			hasUpdated = true
			module.Config = c.fetchCaddyfile(module, cf)
		}
		modules = append(modules, module)
	}

	if hasUpdated {
		fmt.Println("Updating at " + time.Now().String())
		data, err := yaml.Marshal(&modules)
		if err != nil {
			panic(err)
		}
		c.redeploy(data)
		c.pushToGit(data)
	} else {
		fmt.Println("No update needed " + time.Now().String())
	}
}

func (c *CaddyAggregator) UpdateModulesList() {
	rr := newRepositoryRetriever()
	e := c.graphClient.Query(c.clientCTX, &rr, nil)
	if e != nil {
		panic(e)
	}

	c.parseData(rr)
	c.store.SetWithTTL("aggregator", "", 1, time.Minute*time.Duration(c.configuration.StoreTTL))
}
