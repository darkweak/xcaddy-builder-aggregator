package aggregator

import (
	"fmt"
	"github.com/shurcooL/githubv4"
	"regexp"
)

type CaddyModule struct {
	Author      string   `yaml:"author"`
	Config      string   `yaml:"caddyfile"`
	Description string   `yaml:"description"`
	Hash        string   `yaml:"hash"`
	Name        string   `yaml:"name"`
	Repository  string   `yaml:"repository"`
	Tags        []string `yaml:"tags"`
	Version     string   `yaml:"version"`
}

const filename = "Caddyfile"

var authorizedSubfolders = regexp.MustCompile(`((plugins)?/caddy)$`)

func (c *CaddyModule) eventuallyUpdateCaddyFolder(s githubv4.String) {
	if authorizedSubfolders.MatchString(string(s)) {
		c.Repository = fmt.Sprintf("%s/%s", c.Repository, string(s))
	}
}

func (c *CaddyModule) retrieveCaddyfile(release LatestRelease) string {
	for _, v := range release.TagCommit.Commit.Tree.Entries {
		if v.Name == filename {
			return string(v.Path)
		} else {
			c.eventuallyUpdateCaddyFolder(v.Path)
			if v.Object.Tree.Entries != nil {
				for _, v2 := range v.Object.Tree.Entries {
					c.eventuallyUpdateCaddyFolder(v2.Path)
					if v2.Name == filename {
						return string(v2.Path)
					} else {
						if v2.Object.Tree.Entries != nil {
							for _, v3 := range v2.Object.Tree.Entries {
								c.eventuallyUpdateCaddyFolder(v3.Path)
								if v3.Name == filename {
									return string(v3.Path)
								} else {
									if v.Object.Tree.Entries != nil {
										for _, v4 := range v3.Object.Tree.Entries {
											c.eventuallyUpdateCaddyFolder(v4.Path)
											if v4.Name == filename {
												return string(v4.Path)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}
