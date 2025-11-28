# Repository Guidelines

## Project Structure & Module Organization
- `cmd/k8v`: CLI entrypoint that wires Kubernetes client, caches, watchers, and HTTP server.
- `internal/k8s`: Informer setup, resource cache, relationships, transformers, and log streaming.
- `internal/server`: HTTP handlers, WebSocket hubs, log relay, and embedded static UI in `internal/server/static`.
- `internal/types`: Shared resource models sent to the frontend.
- `k8v-poc`: Older proof-of-concept; reference only.
- Frontend lives in `internal/server/static/index.html`; keep assets minimal and framework-free.

## Build, Test, and Development Commands
- `go build -o k8v ./cmd/k8v`: Build the single binary with embedded UI.
- `go run ./cmd/k8v -port 8080`: Run locally against the active kubeconfig context.
- `go test ./...`: Run the Go test suite; prefer adding fast unit tests.
- `go fmt ./...` and `go vet ./...`: Format and vet before sending a PR.

## Coding Style & Naming Conventions
- Go defaults: tabs for indentation, `gofmt` output is the source of truth.
- Keep functions small; pass contexts; avoid global state beyond the hubs/cache singletons already present.
- Name resources clearly (e.g., `deploymentTransformer`, `logHub`) and keep filenames aligned with contained types.
- Frontend: plain HTML/CSS/JS, no bundler—prefer small, readable modules and descriptive IDs/classes.

## Testing Guidelines
- Favor table-driven tests and fakes over live clusters; use `client-go` fakes when stubbing informer behavior.
- Scope tests to specific transformers/handlers; ensure event/order expectations are deterministic.
- If adding new handlers, validate both HTTP status and WebSocket payload shapes.
- Keep tests parallel-safe; avoid mutating shared package-level state without guards.

## Commit & Pull Request Guidelines
- Follow the repo’s short imperative style (e.g., `feat: log viewer`, `fix ns filtering`); keep subjects under ~72 chars.
- PRs should state intent, how to run/verify (`go test ./...`, manual UI steps), and call out Kubernetes context requirements.
- Include screenshots/GIFs for UI tweaks and summarize risk/rollback notes for cluster-facing changes.

## Security & Configuration Tips
- Never commit kubeconfigs, tokens, or cluster metadata; use local kubeconfig for development.
- Confirm the active context with `kubectl config current-context` before running or demoing.
- When logging, avoid printing full secrets or ConfigMap contents; redact sensitive fields at the transformer layer.
