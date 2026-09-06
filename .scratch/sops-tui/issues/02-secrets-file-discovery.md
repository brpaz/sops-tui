# 02 — secrets: file discovery + status detection

**What to build:** A `secrets` package function that recursively scans a root directory for `.yaml`/`.yml`/`.json` files (excluding `.git`, `node_modules`, `vendor`), and for each one determines whether it is SOPS-encrypted by checking for a top-level `sops` metadata key in the parsed content — no `sops` binary invocation required for this. Returns a list of file entries carrying at least the relative path and the encryption status.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] Given a directory tree with a mix of encrypted, plaintext, and non-yaml/json files, listing returns only `.yaml`/`.yml`/`.json` files
- [ ] Files under `.git`, `node_modules`, or `vendor` (at any depth) are excluded
- [ ] A file with a top-level `sops` key (yaml or json) is reported as Encrypted
- [ ] A file without that key is reported as Plaintext
- [ ] Malformed yaml/json in a scanned file does not abort the whole scan — that file is reported with an error/unknown status, other files still listed
- [ ] Unit tests use real fixture files on a temp directory tree, no mocking of the filesystem interface
