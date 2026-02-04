# Repository Guidelines

## Project Structure & Module Organization
- `handler/`: Gin HTTP handlers (e.g., OAuth callback flow).
- `route/`: router wiring for SDK endpoints.
- `middleware/`: reusable Gin middleware (auth checks, context injection).
- `startup/`: startup helpers (OAuth config bootstrap).
- `constant/`: environment keys and context keys.
- `utility/`: request-scope helpers (OAuth config/userinfo accessors).
- `api/`, `models/`, `repository/`: reserved for DTOs, shared models, and persistence (currently empty).

## Build, Test, and Development Commands
- `go mod download`: fetch module dependencies.
- `go fmt ./...`: format all Go files.
- `go vet ./...`: static analysis for common mistakes.
- `go test ./...`: run unit tests (SDK has no standalone server entrypoint).

## Coding Style & Naming Conventions
- Use `gofmt`-standard formatting (tabs for indentation).
- Keep GoDoc-style comments for exported symbols.
- Follow existing package naming patterns such as `bSdkHandler`, `bSdkRoute`, `bSdkUtil` to avoid API churn.
- Prefer clear, descriptive variable names; exported identifiers use `CamelCase`, unexported use `camelCase`.

## Testing Guidelines
- Use Go’s standard `testing` package.
- Place tests alongside code as `*_test.go` in the same package.
- Favor table-driven tests for handler and utility logic.
- No formal coverage target is defined; prioritize OAuth and error-path coverage.

## Commit & Pull Request Guidelines
- This checkout has no Git history, so commit conventions are not discoverable. Recommended format: `type(scope): summary` (e.g., `feat(oauth): add userinfo fetch`).
- PRs should include a concise summary, linked issue/ticket (if any), test evidence (`go test ./...`), and any config/env changes.

## Security & Configuration Tips
- Required env vars include: `SSO_CLIENT_ID`, `SSO_CLIENT_SECRET`, `SSO_REDIRECT_URI`, `SSO_ENDPOINT_AUTH_URI`, `SSO_ENDPOINT_TOKEN_URI`, `SSO_ENDPOINT_USERINFO_URI`.
- Optional: `SSO_WELL_KNOWN_URI` to auto-discover endpoints.
- Never commit secrets or `.env` files. Remove any local `replace` directives in `go.mod` before sharing.
