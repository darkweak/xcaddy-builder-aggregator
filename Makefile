.PHONY: prepare run

DC=docker-compose
DC_BUILD=$(DC) build
DC_EXEC=$(DC) exec

prepare: ## Prepare the app
	cp configuration.yml.dist configuration.yml
	echo "Update the configuration.yml file with a valid Github Personal Access Token"

run: ## Run the aggregator
	go run ./runner/main.go
