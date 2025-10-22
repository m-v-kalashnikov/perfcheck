set shell := ["zsh", "-lc"]

toolchain_raw := `awk -F'"' '/^[[:space:]]*channel/ {print $2; found=1} END {exit (found ? 0 : 1)}' rust/rust-toolchain.toml`
toolchain := trim(toolchain_raw)
components := `awk '
  /^\s*components\s*=/ {
    line = $0
    sub(/.*\[/, "", line)
    while (index(line, "]") == 0 && getline nextline) {
      line = line " " nextline
    }
    sub(/\].*/, "", line)
    gsub(/"/, "", line)
    gsub(/,/, " ", line)
    gsub(/\s+/, " ", line)
    gsub(/^ /, "", line)
    gsub(/ $/, "", line)
    print line
    exit
  }
' rust/rust-toolchain.toml`
cargo_tools := "cargo-nextest cargo-deny cargo-audit cargo-udeps taplo-cli"
go_version := `awk '/^go [0-9]+\.[0-9]+(\.[0-9]+)?/ {print $2; exit}' go/go.mod`
go_tools := "golangci-lint govulncheck"
golangci_lint_version := "v2.5.0"
govulncheck_version := "v1.1.4"

# Show available commands with descriptions
default:
  @just --list --unsorted

# Configure both toolchains and prerequisite tooling
setup: _setup

# Install the required Rust toolchain and components
rust-ensure-toolchain: _rust_ensure_toolchain

# Validate rustup and cargo availability
rust-ensure-tools: _rust_ensure_tools

# Install required Rust cargo-based tooling
rust-tools: _rust_ensure_toolchain _rust_ensure_cargo_tools

# Ensure the configured Go toolchain version is installed
go-ensure-toolchain: _go_ensure_toolchain

# Install required Go-based tooling
go-tools: _go_ensure_tools

# Run Go formatting, linting, analyzers, and vulnerability checks
go-maintain:
	@echo "[go] gofmt check"
	@cd go && set -o pipefail; \
		fmt_out="$(find . -name '*.go' -not -path './vendor/*' -not -path './bin/*' -print0 | xargs -0 gofmt -l)" || exit 1; \
		if [ -n "$fmt_out" ]; then \
			echo "[go] gofmt found unformatted files:"; \
			printf '%s\n' "$fmt_out"; \
			echo "run 'gofmt -w' on the listed files"; \
			exit 1; \
		fi
	@echo "[go] building perfcheck-golangci bridge"
	@mkdir -p go/bin
	@just _checked_run go go build -o bin/perfcheck-golangci ./cmd/perfcheck-golangci
	@echo "[go] golangci-lint run"
	@just _checked_run go 'PATH="./bin:$PATH"' golangci-lint run --config .golangci.yml ./...
	@echo "[go] perfcheck analyzers"
	@just _checked_run go ./bin/perfcheck-golangci ./...
	@echo "[go] go mod verify"
	@just _checked_run go go mod verify
	@echo "[go] govulncheck ./..."
	@just _checked_run go govulncheck ./...

# Execute the Go test suite
go-test:
	@echo "[go] go test ./..."
	@just _checked_run go go test ./...

# Run Rust formatting, linting, audit, and dependency hygiene checks
rust-maintain:
	@echo "[rust] cargo fmt --check"
	@just _checked_run rust cargo +{{toolchain}} fmt --check
	@echo "[rust] cargo clippy --all-targets --all-features -D warnings"
	@just _checked_run rust cargo +{{toolchain}} clippy --all-targets --all-features -- -D warnings
	@echo "[rust] cargo deny check"
	@just _checked_run rust cargo +{{toolchain}} deny check
	@echo "[rust] cargo audit"
	@just _checked_run rust cargo +{{toolchain}} audit
	@echo "[rust] cargo udeps --all-targets"
	@just _checked_run rust cargo +{{toolchain}} udeps --all-targets

# Execute the Rust test suite with nextest
rust-test:
	@echo "[rust] cargo nextest run"
	@just _checked_run rust cargo +{{toolchain}} nextest run

# Run all maintenance and test commands before committing
pre-commit: go-maintain rust-maintain go-test rust-test
	@echo "[pre-commit] openspec validate --strict"
	@openspec validate --strict --all

_checked_run dir +command:
	@cd {{dir}} && { \
		set -o pipefail; \
		if ! output="$({{command}} 2>&1)"; then \
			printf '%s\n' "$output"; \
			exit 1; \
		fi; \
	}

_setup: _rust_ensure_toolchain _rust_ensure_tools _rust_ensure_cargo_tools _go_ensure

_rust_ensure_toolchain:
	@if ! rustup toolchain list | grep -q "{{toolchain}}"; then \
		echo "[setup] installing Rust toolchain {{toolchain}}"; \
		rustup toolchain install {{toolchain}} >/dev/null; \
	fi
	@for component in {{components}}; do \
		if [ -n "$component" ] && ! rustup component list --toolchain {{toolchain}} | grep -q "^$component (installed)"; then \
			echo "[setup] adding component $component for {{toolchain}}"; \
			rustup component add --toolchain {{toolchain}} "$component" >/dev/null; \
		fi; \
	done

_rust_ensure_tools:
	@if ! command -v rustup >/dev/null 2>&1; then \
		echo "rustup not installed; install Rustup to manage toolchains." >&2; \
		exit 1; \
	fi
	@if ! command -v cargo >/dev/null 2>&1; then \
		echo "cargo not installed; install the Rust toolchain." >&2; \
		exit 1; \
	fi

_rust_ensure_cargo_tools:
	@for crate in {{cargo_tools}}; do \
		bin="$crate"; \
		if [ "$crate" = "taplo-cli" ]; then \
			bin="taplo"; \
		fi; \
		if ! command -v "$bin" >/dev/null 2>&1; then \
			echo "[setup] installing $crate"; \
			cargo +{{toolchain}} install --locked "$crate"; \
		fi; \
	done

_go_ensure: _go_ensure_toolchain _go_ensure_tools

_go_ensure_toolchain:
	@if ! command -v go >/dev/null 2>&1; then \
		echo "go not installed; install Go {{go_version}} to build analyzers." >&2; \
		exit 1; \
	fi
	@installed="$(go env GOVERSION)"; \
	if [ "$installed" != "go{{go_version}}" ]; then \
		echo "Go {{go_version}} required, but found $installed. Install matching Go toolchain." >&2; \
		exit 1; \
	fi

_go_ensure_tools:
	@for tool in {{go_tools}}; do \
		if ! command -v "$tool" >/dev/null 2>&1; then \
			echo "[setup] installing $tool"; \
			case "$tool" in \
				golangci-lint) GOTOOLCHAIN=go{{go_version}} GO111MODULE=on go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{golangci_lint_version}} ;; \
				govulncheck) GOTOOLCHAIN=go{{go_version}} GO111MODULE=on go install golang.org/x/vuln/cmd/govulncheck@{{govulncheck_version}} ;; \
				*) echo "Unknown Go tool $tool" >&2; exit 1 ;; \
			esac; \
		fi; \
	done
