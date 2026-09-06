# 02 — secrets: file discovery + status detection

**What to build:** A `secrets` package function that recursively scans a root directory for `.yaml`/`.yml`/`.json` files (excluding `.git`, `node_modules`, `vendor`), and for each one determines whether it is SOPS-encrypted by checking for a top-level `sops` metadata key in the parsed content — no `sops` binary invocation required for this. Returns a list of file entries carrying at least the relative path and the encryption status.

**Blocked by:** None — can start immediately.

**Status:** done

- [x] Given a directory tree with a mix of encrypted, plaintext, and non-yaml/json files, listing returns only `.yaml`/`.yml`/`.json` files
- [x] Files under `.git`, `node_modules`, or `vendor` (at any depth) are excluded
- [x] A file with a top-level `sops` key (yaml or json) is reported as Encrypted
- [x] A file without that key is reported as Plaintext
- [x] Malformed yaml/json in a scanned file does not abort the whole scan — that file is reported with an error/unknown status, other files still listed
- [x] Unit tests use real fixture files on a temp directory tree, no mocking of the filesystem interface

**Follow-up fix (found while building ticket 05):** `.sops.yaml` itself matches the `.yaml` extension filter and was being listed as a scanned file (and sorted first alphabetically, ahead of real secrets). It's now excluded by exact filename, same as the ignored directories. Added `TestList_ExcludesSopsConfigFile` regression test.
