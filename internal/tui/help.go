package tui

const helpOverlay = `sops-tui keybindings

  j / down       move down
  k / up         line up
  g / home       go to start
  G / end        go to end

  v / enter      view decrypted content (read-only)
  E              edit (suspends to sops <file>)
  e              encrypt selected plaintext file (asks to confirm)
  d              decrypt selected file in place (asks to confirm)
  r              refresh (re-scan)
  tab            cycle view: encrypted / plaintext / all

  /              filter by path
  :              command mode (view/encrypt/decrypt/edit/refresh)

  q / ctrl+c     quit
  esc / q        back to list / close overlay

press esc or q to close this help
`
