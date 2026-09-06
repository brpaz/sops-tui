# 01 — Prefactor: CLI wiring cleanup

**What to build:** Remove the scaffolded `hello` sample sub-command and its wiring, and give the root command an optional positional `path` argument (defaults to the current working directory) that later tickets will use as the scan root. The root command's action remains a no-op placeholder for now — it just needs to correctly resolve and expose the chosen path.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] `internal/commands/hello` (and its registration in the composition root) is deleted; no reference to it remains
- [ ] Root command accepts an optional positional argument; when omitted, the resolved path is the current working directory
- [ ] When given, the positional argument is resolved to an absolute path
- [ ] Existing `--version`/`--help` behavior from the scaffold still works unchanged
- [ ] Unit test covers: no argument → cwd; explicit relative/absolute argument → correct absolute path
