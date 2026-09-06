# 08 — Edit action (suspend + exec `sops <file>`)

**What to build:** Selecting a file and triggering the edit action suspends the TUI, hands the terminal to the `secrets` package's `Edit` function (foreground `sops <file>`, inherited stdio), and resumes the TUI with a fresh scan when the subprocess exits — whether the user saved changes or not.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Triggering edit suspends the TUI's screen and hands the terminal fully to the `sops` subprocess
- [ ] On the subprocess exiting normally, the TUI resumes and the list is refreshed
- [ ] If the edited file was plaintext before and is re-encrypted by `sops` during the edit, the row's status reflects Encrypted after resume
- [ ] If the subprocess exits non-zero (e.g. editor aborted, decrypt failed), the TUI resumes without crashing and surfaces that the edit did not complete
