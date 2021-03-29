package main

import (
	aggregator "github.com/darkweak/xcaddy-builder-agregator"
)

func main()  {
	c := make(chan int)
	a := &aggregator.CaddyAggregator{}
	a.LoadConfiguration(c)
	a.UpdateModulesList()

	initializeNotifiedChannel(c, a)
}

func initializeNotifiedChannel(c chan int, aggregator *aggregator.CaddyAggregator) {
	for {
		<-c
		aggregator.UpdateModulesList()
	}
}
