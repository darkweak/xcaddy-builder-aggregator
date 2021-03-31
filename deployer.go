package aggregator

import (
	"fmt"
	"github.com/google/go-github/v34/github"
	"io/ioutil"
	"log"
)

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
