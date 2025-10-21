set shell := ["zsh", "-lc"]

# Show available commands with descriptions
default:
  @just --list --unsorted

smoke:
	@echo "[smoke] go test ./..."
	@cd go && go test ./...
	@echo "[smoke] cargo nextest run"
	@cd rust && cargo nextest run
	@echo "[smoke] openspec validate --strict"
	@openspec validate --strict
