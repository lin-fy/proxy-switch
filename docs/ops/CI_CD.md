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

## Release

Create a version commit on `dev`, merge it into `main`, and create a tag matching the version in `build/config.yml`:

```text
v0.1.0
```

`.github/workflows/release.yml` rejects tags that are not based on `main` or do not match `build/config.yml`. It builds the user-scope NSIS installer and portable ZIP, signs the executable and installer, generates `SHA256SUMS.txt`, and creates a GitHub Release.

Configure the `production` environment with these secrets before the first formal release:

```text
WINDOWS_CERTIFICATE_BASE64
WAILS_WINDOWS_CERT_PASSWORD
```

The certificate is written only to the temporary runner directory and is not included in artifacts.

## Review

The project review workflow is in `.agents/skills/review-pr`. GitHub Actions provide hard build and test checks. Codex Review provides semantic and security review, while a human remains responsible for merge and release approval.
