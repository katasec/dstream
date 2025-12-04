.PHONY: help up down restart

DEFAULT_GOAL := help

help:
	@echo "Available commands:"
	@echo "  make up       - Start MSSQL container"
	@echo "  make down     - Stop MSSQL container"
	@echo "  make restart  - Restart MSSQL container"

up:
	powershell -NoProfile -ExecutionPolicy Bypass -File "$(CURDIR)/scripts/mssql-up.ps1"

down:
	powershell -NoProfile -ExecutionPolicy Bypass -File "$(CURDIR)/scripts/mssql-down.ps1"

restart: down up
