## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**
```bash
bd ready --json
```

**Create new issues:**
```bash
bd create "Issue title" -t bug|feature|task -p 0-4 --json
bd create "Issue title" -p 1 --deps discovered-from:bd-123 --json
bd create "Subtask" --parent <epic-id> --json  # Hierarchical subtask (gets ID like epic-id.1)
```

**Claim and update:**
```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**
```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`
6. **Commit together**: Always commit the `.beads/issues.jsonl` file together with the code changes so issue state stays in sync with code state

### Parallel Async Subagents for Maximum Efficiency

**Use parallel async subagents often** to maximize productivity and minimize wait time.

**When to use parallel agents:**
- Reading multiple independent files for analysis
- Running multiple independent grep/search operations
- Executing independent git commands (e.g., `git status` and `git diff`)
- Performing multiple web searches or API calls
- Any set of operations with no dependencies between them

**How to launch parallel agents:**
Launch multiple tool calls in a single message block when tasks are independent. For example, when you need to read three files and run a search, launch all four operations together rather than sequentially.

**Benefits:**
- Faster task completion through concurrent execution
- Better resource utilization
- Reduced overall latency
- More efficient workflow for complex multi-step tasks

**Examples of good parallelization:**
- Analyzing a codebase: Read multiple source files + search for patterns in parallel
- Pre-commit checks: Run `git status`, `git diff`, and `git log` together
- Research tasks: Multiple web searches or file reads simultaneously
- Code review: Read changed files + search for related code in parallel

**When NOT to parallelize:**
- Operations with dependencies (e.g., read file before editing it)
- Sequential workflows where later steps depend on earlier results
- Operations that modify state (commits, file writes) that must happen in order

### Auto-Sync

bd automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### GitHub Copilot Integration

If using GitHub Copilot, also create `.github/copilot-instructions.md` for automatic instruction loading.
Run `bd onboard` to get the content, or see step 2 of the onboard instructions.

### MCP Server (Recommended)

If using Claude or MCP-compatible clients, install the beads MCP server:

```bash
pip install beads-mcp
```

Add to MCP config (e.g., `~/.config/claude/config.json`):
```json
{
  "beads": {
    "command": "beads-mcp",
    "args": []
  }
}
```

Then use `mcp__beads__*` functions instead of CLI commands.

### Claude Code Integration and Automatic Service Startup

This project integrates Claude Code with the Happy daemon for seamless AI-powered development workflows.

**What's Included:**
- Claude Code is installed via npm (`@anthropic-ai/claude-code`, compatible with Claude Code versions)
- Happy daemon and Claude Code automatically start after authentication
- One-time authentication step via `happy --no-qr`
- Credentials persist on PVC for automatic service startup on pod restart
- Background authentication watcher polls every 5 seconds for completion

**Getting Started:**

1. **Authenticate Once** (first time in pod):
   ```bash
   happy --no-qr
   ```
   Follow the prompts to authenticate with your credentials.

2. **Services Start Automatically**:
   Once authenticated, the pod's entrypoint automatically starts:
   - Happy daemon (credentials persisted on PVC)
   - Claude Code service
   - Background watcher monitoring authentication

3. **On Pod Restart**:
   Both services automatically start without requiring re-authentication, thanks to persisted credentials.

**Requirements:**
- Happy CLI version 0.13.0 or later
- Authentication credentials stored on persistent volume claim (PVC)
- Node.js environment for Claude Code npm package

**Architecture:**
- Entrypoint script detects authentication status
- Background watcher polls auth completion every 5 seconds (non-blocking)
- Both services start concurrently once auth is confirmed
- Services remain running for the duration of the pod

This setup enables AI agents and developers to use Claude Code seamlessly within the Happy environment without managing multiple authentication flows.

### Managing AI-Generated Planning Documents

AI assistants often create planning and design documents during development:
- PLAN.md, IMPLEMENTATION.md, ARCHITECTURE.md
- DESIGN.md, CODEBASE_SUMMARY.md, INTEGRATION_PLAN.md
- TESTING_GUIDE.md, TECHNICAL_DESIGN.md, and similar files

**Best Practice: Use a dedicated directory for these ephemeral files**

**Recommended approach:**
- Create a `history/` directory in the project root
- Store ALL AI-generated planning/design docs in `history/`
- Keep the repository root clean and focused on permanent project files
- Only access `history/` when explicitly asked to review past planning

**Example .gitignore entry (optional):**
```
# AI planning documents (ephemeral)
history/
```

**Benefits:**
- ✅ Clean repository root
- ✅ Clear separation between ephemeral and permanent documentation
- ✅ Easy to exclude from version control if desired
- ✅ Preserves planning history for archeological research
- ✅ Reduces noise when browsing the project

### CLI Help

Run `bd <command> --help` to see all available flags for any command.
For example: `bd create --help` shows `--parent`, `--deps`, `--assignee`, etc.

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ✅ Store AI planning docs in `history/` directory
- ✅ Run `bd <cmd> --help` to discover available flags
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems
- ❌ Do NOT clutter repo root with planning documents

For more details, see README.md and QUICKSTART.md.
