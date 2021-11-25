{ pkgs ? import (builtins.fetchTarball {
  name = "nixos-unstable-2021-11-25";
  url = "https://github.com/nixos/nixpkgs/archive/7c4bbc7cd008fb3802e3b5ea44118ebfc013578d.tar.gz";
  sha256 = "0sxppb7caw14a2ifi8p2wia3cgjb6vdf2ia5386s3y8l90m58a2z"; # Hash obtained using `nix-prefetch-url --unpack <url>`
}) {} }:

with pkgs;

mkShell {
  buildInputs = [
    gitlint
    gnupg
    go_1_17
    go-tools
    go-mockery
    gogetdoc
    golangci-lint
    goreleaser
    gosec
    gotools
    nodejs-16_x
    nodePackages.pnpm
    pre-commit
  ];

  shellHook =
    ''
      # Setup the terminal prompt.
      export PS1="(nix-shell) \W $ "

      # Setup aliases.
      alias init='pre-commit install'
      alias down='docker-compose down -p appy --remove-orphans'
      alias up='docker-compose up -p appy -d --remove-orphans'
      alias install='go mod download && pnpm i -C admin && pnpm i -C docs'

      # Clear the terminal screen.
      clear
    '';
}
