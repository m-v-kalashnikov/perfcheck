set shell := ["zsh", "-lc"]

# Show available commands with descriptions
default:
  @just --list --unsorted

lint-go:
	@echo "[lint] building perfcheck-golangci bridge"
	@mkdir -p go/bin
	@cd go && go build -o bin/perfcheck-golangci ./cmd/perfcheck-golangci
	@echo "[lint] golangci-lint run"
	@cd go && PATH="$(pwd)/bin:$PATH" golangci-lint run --config .golangci.yml ./...
	@echo "[lint] perfcheck analyzers"
	@cd go && ./bin/perfcheck-golangci ./...

maintain-rust:
	@echo "[rust] cargo fmt --check"
	@cd rust && cargo fmt --check
	@echo "[rust] cargo clippy --all-targets --all-features -D warnings"
	@cd rust && cargo clippy --all-targets --all-features -- -D warnings
	@echo "[rust] cargo deny check"
	@cd rust && cargo deny check
	@echo "[rust] cargo audit"
	@cd rust && cargo audit
	@echo "[rust] cargo udeps --all-targets"
	@cd rust && cargo +nightly udeps --all-targets

smoke:
	@just lint-go
	@just maintain-rust
	@echo "[smoke] go test ./..."
	@cd go && go test ./...
	@echo "[smoke] cargo nextest run"
	@cd rust && cargo nextest run
	@echo "[smoke] openspec validate --strict"
	@openspec validate --strict --all
