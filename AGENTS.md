<!-- TRELLIS:START -->
# Trellis Instructions

These instructions are for AI assistants working in this project.

This project is managed by Trellis. The working knowledge you need lives under `.trellis/`:

- `.trellis/workflow.md` — development phases, when to create tasks, skill routing
- `.trellis/spec/` — package- and layer-scoped coding guidelines (read before writing code in a given layer)
- `.trellis/workspace/` — per-developer journals and session traces
- `.trellis/tasks/` — active and archived tasks (PRDs, research, jsonl context)

If a Trellis command is available on your platform (e.g. `/trellis:finish-work`, `/trellis:continue`), prefer it over manual steps. Not every platform exposes every command.

If you're using Codex or another agent-capable tool, additional project-scoped helpers may live in:
- `.agents/skills/` — reusable Trellis skills
- `.codex/agents/` — optional custom subagents

Managed by Trellis. Edits outside this block are preserved; edits inside may be overwritten by a future `trellis update`.

<!-- TRELLIS:END -->

## Repo-Specific Guidance

### Current repo shape

- Primary application code lives under `backend/`; `backend/go.mod` is the active module manifest.
- Root `package.json` is empty and does not define repo-level scripts.
- `frontend/` is not a checked-in app yet; it currently contains only `node_modules/`, so do not assume established frontend structure or commands.

### Backend boundaries and entrypoints

- Main backend areas are `backend/common`, `backend/api`, and `backend/services`.
- Current Go service entrypoints are:
  - `backend/services/gateway/main.go`
  - `backend/services/user/main.go`
  - `backend/services/forum/main.go`
  - `backend/services/audit/main.go`
  - `backend/services/es/main.go`
- The RAG service is a separate Python service with entrypoint `backend/services/rag/main.py`.
- Service startup is config-driven; for example, `forum` loads `config/forum.yaml` from `main.go` and wires infrastructure clients there.

### Local run workflow

- Prefer the Docker wrapper at `./backend/deploy/bin/run ...` from the repo root.
- That wrapper enters `backend/deploy/docker/`, creates `.env` from `.env.example` if missing, and runs `docker compose -f docker-compose.service.yml ...`.
- The executable source of truth for service names, ports, and infra dependencies is `backend/deploy/docker/docker-compose.service.yml`.

### Runtime/config gotchas

- Many service configs are written for the Docker network and use hostnames like `postgres`, `redis`, `minio`, `rabbitmq`, `qdrant`, and `rag-model-server` rather than `localhost`.
- If you run a service directly on the host instead of through compose, expect to override those hosts or adjust config first.
- `gateway-service` proxies to fixed container hostnames such as `http://user-service:8081` and `http://forum-service:8082`.
- `rag-service` expects OpenAI-compatible local endpoints via Docker-provided env vars such as `RAG_LLM_BASE_URL` and `RAG_LLM_API_KEY`.

### Command and verification constraints

- Do not claim repo-wide npm, Make, lint, test, or CI commands unless you verify them in the relevant subproject first; none are defined at the repo root.
- No checked-in Go test files were found under `backend/`, so avoid assuming an existing backend test suite.
- `backend/deploy/bin/protoc_gen` is not a general proto regeneration script; it currently regenerates only `backend/api/user/user.proto`.
