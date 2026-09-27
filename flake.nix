{
  description = "gocrawler - a web crawler in golang";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      treefmt-nix,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };

        treefmtEval = treefmt-nix.lib.evalModule pkgs {
          projectRootFile = "flake.nix";

          programs.nixfmt.enable = true;
          programs.gofmt.enable = true;

          settings.global.excludes = [
            "*.lock"
          ];
        };
      in
      {
        formatter = treefmtEval.config.build.wrapper;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            goreleaser
            chromium
          ];
        };
      }
    );
}
