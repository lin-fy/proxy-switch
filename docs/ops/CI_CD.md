# CI/CD

## Scope

Codex Provider Hub is a Windows desktop application. CI validates the Go, frontend, and Wails build on a Windows runner. CD creates signed Windows release artifacts; it does not deploy a server environment.

## Continuous integration

`.github/workflows/ci.yml` runs for pushes to `dev`, pull requests targeting `dev` or `main`, and manual dispatches. It runs:

```text
npm ci
npm run format:check
npm run lint
npm run build
gofmt check
go mod tidy check
go vet . ./internal/...
go test . ./internal/... -count=1
wails3 build
```

The Windows executable and portable ZIP are retained as workflow artifacts for seven days.

## Agent-independent local verification

All coding agents use the same PowerShell verifier; the selected model does not change the checks. From the repository root:

Use PowerShell 7.3+, Go 1.25 and Node.js 22, then install the locked frontend dependencies with `npm ci` in `frontend/`. A fresh checkout needs `task check:frontend` before `task check:backend`, because the Go entry point embeds `frontend/dist`. The complete `task ci` already runs them in this order and additionally requires Wails `v3.0.0-beta.26`.

```text
task check:frontend   # format, lint, type-check, production frontend build
task check:backend    # gofmt, go mod tidy check, vet, tests
task ci               # both scopes plus Wails Windows build and portable ZIP
```

`Taskfile.yml` and GitHub Actions both delegate to `scripts/ci/verify.ps1`. CI never calls an LLM. If an AI reviewer or failure diagnostician is added later, it may analyze the logs after these checks fail, but it must not approve or publish artifacts by itself.

When Go Task is not installed, invoke the same implementation directly, for example `pwsh -NoProfile -File ./scripts/ci/verify.ps1 -Scope backend`.

For Windows builds, the verifier also searches `GOPATH\bin` for an installed Wails CLI. It adjusts only the verification process's PATH; user and system PATH settings are unchanged.

## Release

After M6-C and M6-D pass, create the release version commit on `dev`, merge it into `main`, and create a tag matching the version in `build/config.yml`. The planned V1 release is:

```text
v1.0.0
```

The current development baseline remains `0.1.0` until the release gates pass; do not create the `v1.0.0` tag early.

`.github/workflows/release.yml` rejects tags that are not based on `main` or do not match `build/config.yml`. It builds the user-scope NSIS installer and portable ZIP, signs the executable and installer, generates `SHA256SUMS.txt`, and creates a GitHub Release.

Configure the `production` environment with these secrets before the first formal release:

```text
WINDOWS_CERTIFICATE_BASE64
WAILS_WINDOWS_CERT_PASSWORD
```

The certificate is written only to the temporary runner directory and is not included in artifacts.

## Review

The project review workflow is in `.agents/skills/review-pr`. GitHub Actions provide hard build and test checks. Codex Review provides semantic and security review, while a human remains responsible for merge and release approval.
