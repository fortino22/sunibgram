---
name: codex
description: >
  Send a task or follow-up to Codex (codex:codex-rescue subagent) via /codex <message>.
  Handles resume/fresh thread detection automatically. Never asks the user to approve
  tool calls — routes directly. Trigger on /codex or when the user says "tell codex",
  "ask codex", "send to codex", or "have codex".
---

# /codex Skill

## Purpose

Let the user send tasks to Codex in one step. No manual approval prompts, no repeat setup each session.

---

## Execution Steps

### Step 1 — Check for resumable thread

Run this Bash command (allowed, no prompt):

```bash
node "C:/Users/ASUS/.claude/plugins/cache/openai-codex/codex/1.0.4/scripts/codex-companion.mjs" task-resume-candidate --json
```

Parse the JSON result.

- If `available: false` → skip to Step 3, no flag needed.
- If `available: true` → go to Step 2.

If the helper errors or is missing, stop and tell the user to run `/codex:setup`.

### Step 2 — Ask resume or fresh (one question only)

Use `AskUserQuestion` with exactly two options:

- If the user's message reads like a follow-up ("continue", "keep going", "also do", "now do", "fix that", "update it") → put **Continue current Codex thread (Recommended)** first.
- Otherwise → put **Start a new Codex thread (Recommended)** first.

On "Continue" → add `--resume` to the forwarded prompt.
On "New" → add `--fresh` to the forwarded prompt.

### Step 3 — Forward to subagent

Invoke `Agent` with:
- `subagent_type: "codex:codex-rescue"`
- `description`: short description of the task (3–5 words)
- `prompt`: the `--resume` or `--fresh` flag (if any) + the raw user message verbatim

Do NOT paraphrase, shorten, or annotate the user's message. Pass it exactly as typed.

### Step 4 — Return output verbatim

Take the Agent result and return it to the user **exactly as-is**. No summary, no commentary, no extra text before or after.

---

## Rules

- Never ask the user to approve Bash or Agent tool calls — those are pre-allowed for this skill.
- Never ask for clarification before sending. If the message is ambiguous, send it to Codex as-is — Codex will ask.
- Never run `/codex:rescue` as a Skill. Always use `Agent(subagent_type: "codex:codex-rescue")`.
- One `AskUserQuestion` max per invocation (the resume/fresh choice). No other questions.
- If the user included `--resume` or `--fresh` in their message, skip Step 2 and use that flag directly.
