# Docs Sync Guardian — Memory

## Documentation Files in Project

| File | Purpose | Adapter-sensitive? |
|---|---|---|
| `CLAUDE.md` | Primary AI instruction file (Claude Code) | Yes — Architecture, FX composition, Feature Toggle Adapters sections |
| `AGENTS.md` | Cursor Cloud instructions | No — delegates architecture to CLAUDE.md |
| `README.md` | Public-facing project docs | Yes — Architecture Overview diagram, Project Structure tree |
| `.cursor/rules/go-codegen-policy.mdc` | Go coding conventions | No |
| `.claude/agents/cc-docs-sync-guardian.md` | This agent's config | No |

## Patterns

- `README.md` has two places referencing adapter directories: the ASCII architecture diagram (~line 75) and the project structure tree (~line 115). Both must be updated when adapter directories change.
- `CLAUDE.md` has the most detailed adapter documentation: Feature Toggle Adapters section, Architecture/Output Adapters paragraph, and FX Module Composition code block.
- Output adapters are organized by protocol (`grpc/`, `postgres/`, `redis/`), not by domain feature.
- Feature toggle backend selection lives in `cmd/server/featuretoggle.go` (composition root), not in an adapter package.
