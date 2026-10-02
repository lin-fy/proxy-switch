# Proxy Switch review checklist

- Dependency direction remains `interfaces → application → domain`, with infrastructure implementing ports.
- Windows-only V1 scope is preserved; no platform or protocol expansion is hidden in a refactor.
- Route activation validates the provider and model before writing configuration.
- Active and profile Codex configuration files are backed up before writes and restored after partial failure.
- A failed reachability check cannot persist a new default route.
- API keys and credentials remain references; logs, errors, state files, and generated artifacts do not contain secrets.
- Windows process launch, restart, Credential Manager, paths, and file replacement handle failure safely.
- Wails bindings are regenerated from source and remain compatible with the frontend API.
- Non-trivial logic has focused tests, especially around rollback, ordering, concurrency, and filesystem errors.
- Changes do not weaken CI, release validation, signing, or artifact integrity.
