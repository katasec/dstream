SHELL := /bin/bash

.PHONY: help mssql-up mssql-up-force mssql-down mssql-test mssql-insert-person
.DEFAULT_GOAL := help

help: ## Show available make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


mssql-up: ## Start SQL Server and verify TestDB+CDC (preserves existing DB)
	./scripts/mssql-init-safe.sh
	./scripts/mssql-verify.sh


mssql-up-force: ## Start SQL Server, drop and recreate TestDB+CDC
	./scripts/mssql-init.sh
	./scripts/mssql-verify.sh


mssql-down: ## Stop and remove local SQL Server container
	./scripts/mssql-stop.sh


mssql-test: ## Run the configured DStream MSSQL task
	go run . run mssql-test --log-level debug


mssql-insert-person: ## Insert one random row into dbo.Persons
	./scripts/mssql-insert-person.sh
