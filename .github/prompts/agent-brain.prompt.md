---
name: agent-brain
description: "Stage 1 kickoff: build services/diagnosis and services/event-store — GPT-4o RCA engine and SQLite event persistence."
---

You are the Brain agent for Stage 1 of the AI Infrastructure Resilience Copilot build.

Read `.github/instructions/brain.instructions.md` for your full task list and done criteria.

Start with event-store (diagnosis depends on it):
1. `services/event-store/internal/db.go` — SQLite setup and schema migration.
2. `services/event-store/internal/handlers/` — REST handlers for all tables.
3. Wire in `services/event-store/main.go`.
4. Test: POST and GET /incidents.

Then diagnosis:
1. `services/diagnosis/internal/llm/client.go` — GitHub Models GPT-4o client.
2. `services/diagnosis/internal/analyzer.go` — evidence fetcher + LLM prompt builder + response parser.
3. Wire `POST /diagnose` in `services/diagnosis/main.go`.
4. Test: POST /diagnose with a sample payload returns RCAPayload JSON.

GITHUB_TOKEN is in the environment — do not hardcode it. Use `os.Getenv("GITHUB_TOKEN")`.
