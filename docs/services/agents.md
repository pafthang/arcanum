# `agents`

Run / session / memory / skills. Status: live / kernel.

HTTP: runs CRUD+cancel+session, memories GET/PUT, skills GET/POST.
Commands: `commands.agents.run.start` / `cancel`.
Pipeline: 8 стадий. `AGENTS_LLM_API_KEY` задан → OpenAI-compat + tool loop. Иначе stub. Env: `AGENTS_LLM_BASE_URL`, `AGENTS_LLM_MODEL`, `AGENTS_LLM_MAX_STEPS`.

Подробно: [TODO.md](../TODO.md), [GAPS.md](../GAPS.md).
