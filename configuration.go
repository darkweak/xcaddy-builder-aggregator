package aggregator

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type CaddyAggregatorConfiguration struct {
	Email      string `yaml:"email"`
	FilePath   string `yaml:"file_path"`
	Owner      string `yaml:"owner"`
	Pat        string `yaml:"pat"`
	Path       string `yaml:"path"`
	Repository string `yaml:"repository"`
	StoreTTL   int    `yaml:"store_ttl"`
}

func Parse(s string, i interface{}) error {
	data, err := ioutil.ReadFile(s)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(data, i)
	return err
}

func (c *CaddyAggregatorConfiguration) ParseConfiguration() error {
	return Parse("configuration.yml", c)
}
