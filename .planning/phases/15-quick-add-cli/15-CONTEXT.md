# Phase 15: Quick-Add CLI - Context

**Gathered:** 2026-02-02
**Status:** Ready for planning

<vision>
## How This Should Work

Add tasks from the terminal without opening the TUI. Type `dflow add "fix the bug"` and you're done.

When you add a task, you get confirmation — the task echoed back plus the current count. Something like "✓ Added: fix the bug (3 tasks)". Not fire-and-forget, but not verbose either. Just enough feedback to know it worked.

</vision>

<essential>
## What Must Be Nailed

- **Zero friction** — Must be fast to type, no flags or ceremony. Just `dflow add "task"` and done.
- **Reliable sync** — Task must actually end up in the system correctly every time.
- **Good CLI feel** — Output should look polished, feel like a proper CLI tool.

</essential>

<specifics>
## Specific Ideas

- Output should feel like GitHub CLI (gh) — clean, colorful, professional
- Confirmation shows task + count: "✓ Added: fix the bug (3 tasks)"
- Match the quality bar of modern CLI tools

</specifics>

<notes>
## Additional Context

No additional notes

</notes>

---

*Phase: 15-quick-add-cli*
*Context gathered: 2026-02-02*
