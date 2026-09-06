{
  pkgs,
  lib,
  config,
  inputs,
  ...
}: {
  languages = {
    go = {
      enable = true;
      enableHardeningWorkaround = true;
      package = pkgs.go;
    };
  };
  dotenv.enable = true;

  # https://devenv.sh/packages/
  packages = with pkgs; [
    go-task
    gotestsum
    goreleaser
    gomarkdoc
    lefthook
    commitlint-rs
    python313Packages.mkdocs
    python313Packages.mkdocs-material
    python313Packages.mkdocs-mermaid2-plugin
    shellcheck
  ];

  enterShell = ''
    if [ -f .env ]; then
      cp env.example .env
    fi
    lefthook install
    go mod tidy
  '';

  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # https://devenv.sh/tests/
  enterTest = ''
    go test ./... -v
  '';

  # https://devenv.sh/git-hooks/
  # git-hooks.hooks.shellcheck.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
