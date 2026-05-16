---
name: task
description: "Manage the Forge task system — create, move, and track tasks in .github/tasks/. Use when a design decision produces actionable work that must be tracked."
user-invocable: false
disable-model-invocation: true
allowed-tools: Read, Write, Edit, Bash
---

Manage the Forge task system in `.github/tasks/`.

## Task Locations

- **Backlog**: `.github/tasks/backlog/` — ready to start, dependencies met
- **In-progress**: `.github/tasks/in-progress/` — actively worked
- **Done**: `.github/tasks/done/` — merged to main
- **Index**: `.github/tasks/index.md` — single source of truth (Architect updates this)

## Actions

**`list`**:
- Read `.github/tasks/index.md` and display it.
- List any files in `in-progress/` with their first 5 lines.

**`new <title>`**:
1. Read `index.md` to find the next T-XX number.
2. Create `backlog/T-XX-<kebab-title>.md` using `.github/tasks/TASK_TEMPLATE.md`.
3. Fill in the ID, title, and today's date.
4. Add a new row to `index.md`: `| T-XX | <title> | Code Expert | 🔲 backlog | — |`

**`start T-XX`**:
1. Find `backlog/T-XX-*.md`.
2. Move it to `in-progress/` (copy + delete original).
3. Update its frontmatter: `status: in-progress`.
4. Update `index.md` row: change emoji to `▶️ in-progress`.

**`done T-XX`**:
1. Find `in-progress/T-XX-*.md`.
2. Move it to `done/` (copy + delete original).
3. Update its frontmatter: `status: done`.
4. Update `index.md` row: change emoji to `✅ done`.

## Index Row Format

```
| T-XX | Title | Agent | 🔲 backlog | Depends on |
```

Status emoji: `🔲 backlog` · `▶️ in-progress` · `✅ done` · `⏸ blocked`
