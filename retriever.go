package aggregator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func retrieveGithubModules(c *CaddyAggregator) GithubRepositoryRetriever {
	rr := newRepositoryRetriever()
	e := c.graphClient.Query(c.clientCTX, &rr, nil)
	if e != nil {
		panic(e)
	}

	return rr
}

type CaddyModuleItem struct {
	Repository string `json:"path"`
	Modules []struct {
		Description string `json:"docs"`
	} `json:"modules"`
	Hash string `json:"updated"`
}

type CaddyModuleResponse struct {
	Result []CaddyModuleItem `json:"result"`
}

func retrieveCaddyModules() []CaddyModuleItem {
	r, e := http.Get("https://caddyserver.com/api/packages")
	if e != nil {
		panic(e)
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var list CaddyModuleResponse

	err = json.Unmarshal(b, &list)
	if err != nil {
		log.Fatalln(err)
	}

	return list.Result
}

type Retriever struct {
	Caddy  []CaddyModuleItem
	Github GithubRepositoryRetriever
}

func retrieveAllData(c *CaddyAggregator) Retriever {
	return Retriever{
		Caddy: retrieveCaddyModules(),
		Github: retrieveGithubModules(c),
	}
}
