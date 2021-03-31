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
	"regexp"
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
	if _, err := time.Parse(time.RFC3339, module.Hash); err != nil && module.Version != "" && module.Hash != "" {
		key := module.Repository
		hash, _ := c.store.Get(key)

		if hash != module.Hash {
			c.store.Set(key, module.Hash, 1)
			return true
		}
	}

	return false
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

func containsModule(list []CaddyModule, module CaddyModule) bool {
	if regexp.MustCompile("caddyserver/caddy/v2/modules").MatchString(module.Repository) {
		return true
	}
	for _, v := range list {
		if v.Repository == module.Repository {
			return true
		}
	}
	return false
}

func (c *CaddyAggregator) parseData(r Retriever) {
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
	for _, cm := range r.Caddy {
		sPath := strings.Split(cm.Repository, "/")
		description := ""
		if len(cm.Modules) > 0 {
			description = cm.Modules[0].Description
		}
		name := sPath[2]
		if name == "caddy-ext" {
			name = sPath[3]
		}
		module := CaddyModule{
			Author:      sPath[1],
			Config:      "",
			Name:        name,
			Description: description,
			Hash:        cm.Hash,
			Repository:  cm.Repository,
			Tags:        []string{},
			Version:     "latest",
		}

		if c.hasChanged(module) {
			hasUpdated = true
		}
		if !containsModule(modules, module) {
			modules = append(modules, module)
		}
	}

	for _, edge := range r.Github.Search.Edges {
		r := edge.Node.Repository
		names := strings.Split(string(r.NameWithOwner), "/")
		s := strings.Split(string(r.LatestRelease.TagCommit.CommitUrl), "/")
		tags := []string{}
		for _, tag := range r.RepositoryTopics.Nodes {
			tags = append(tags, string(tag.Topic.Name))
		}
		version := string(r.LatestRelease.TagName)
		if version == "" {
			version = "latest"
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
			Version:     version,
		}

		cf := module.retrieveCaddyfile(r.LatestRelease)
		if c.hasChanged(module) {
			hasUpdated = true
			module.Config = c.fetchCaddyfile(module, cf)
		}
		if !containsModule(modules, module) {
			modules = append(modules, module)
		}
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
	c.parseData(retrieveAllData(c))
	c.store.SetWithTTL("aggregator", "", 1, time.Minute*time.Duration(c.configuration.StoreTTL))
}
