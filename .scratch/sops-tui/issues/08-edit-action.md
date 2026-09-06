# 08 — Edit action (suspend + exec `sops <file>`)

**What to build:** Selecting a file and triggering the edit action suspends the TUI, hands the terminal to the `secrets` package's `Edit` function (foreground `sops <file>`, inherited stdio), and resumes the TUI with a fresh scan when the subprocess exits — whether the user saved changes or not.

**Blocked by:** 04

**Status:** done

- [x] Triggering edit suspends the TUI's screen and hands the terminal fully to the `sops` subprocess (bound to `E`; uses `tea.ExecProcess` + a new `secrets.EditCommand`, since bubbletea must own the terminal release/restore around the subprocess rather than `secrets.Edit`'s own direct `cmd.Run()`)
- [x] On the subprocess exiting normally, the TUI resumes and the list is refreshed
- [x] If the edited file was plaintext before and is re-encrypted by `sops` during the edit, the row's status reflects Encrypted after resume
- [x] If the subprocess exits non-zero (e.g. editor aborted, decrypt failed), the TUI resumes without crashing and surfaces that the edit did not complete (error pane naming the file)

Verified with a real end-to-end pty smoke test (this sandbox has no `/dev/tty` for a true interactive session): built binary, pre-encrypted fixture, a stub `$EDITOR` script, launched under a Python pty, sent `E` then `q`. Confirmed the file gained the stub editor's appended line and was re-encrypted (two `ENC[...]` fields on exit vs one before), and the process exited 0 — real proof of suspend → decrypt → edit → re-encrypt → resume, not just a mocked unit test.
