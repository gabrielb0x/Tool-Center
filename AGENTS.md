# AGENTS Guidelines for Tool Center

This repository contains a Go API and a static web frontend.

## Directory structure
- `api/` – Go backend
- `frontend/` – static HTML/JS/CSS

## Programmatic checks
- After modifying Go code in `api/`, run `go vet ./...`. If the command cannot finish due to missing dependencies or network restrictions, record the output in the PR.

## Coding style
- Provide complete, working code. Expose configurable parameters through variables or a JSON configuration file (`api/example config.json`) so users can adjust behaviour easily.
- Avoid placeholder code whenever possible.

## Documentation
- Update `README.md` when adding or modifying features.

## Collaboration notes
- When creating tasks, specify which parts of the repo the changes should target and share any relevant configs or examples. This helps the assistant deliver better-quality work.
