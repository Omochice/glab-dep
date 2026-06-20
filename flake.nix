{
  description = "A GitLab CLI extension that streamlines the review and merge workflow for automated dependency update MRs";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks = {
      url = "github:cachix/git-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    nur-packages = {
      url = "github:Omochice/nur-packages";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      treefmt-nix,
      flake-utils,
      git-hooks,
      nur-packages,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [
            nur-packages.overlays.default
          ];
        };
        treefmt = treefmt-nix.lib.evalModule pkgs (
          { ... }:
          let
            rumdlConfig = (pkgs.formats.toml { }).generate "rumdl.toml" {
              # keep-sorted start
              MD004.style = "dash";
              MD007.indent = 4;
              MD007.style = "fixed";
              MD041.enabled = false;
              MD049.style = "underscore";
              MD050.style = "asterisk";
              MD055.style = "leading-and-trailing";
              MD060.enabled = true;
              MD060.style = "aligned";
              MD077.enabled = false;
              global.line_length = 0;
              # keep-sorted end
            };
          in
          {
            settings.global.excludes = [ ];
            settings.formatter.rumdl-format.options = [
              "--config"
              (toString rumdlConfig)
            ];
            # treefmt-nix's programs.golangci-lint runs `golangci-lint run --fix`,
            # which rejects file lists spanning multiple directories. Use the
            # dedicated `fmt` subcommand so the formatters in .golangci.yml apply.
            settings.formatter.golangci-lint = {
              command = "${pkgs.lib.getExe pkgs.golangci-lint}";
              options = [ "fmt" ];
              includes = [ "*.go" ];
              excludes = [ "vendor/*" ];
            };
            programs = {
              # keep-sorted start block=yes
              keep-sorted.enable = true;
              nixfmt.enable = true;
              rumdl-format.enable = true;
              toml-sort.enable = true;
              yamlfmt = {
                enable = true;
                settings = {
                  formatter = {
                    type = "basic";
                    retain_line_breaks_single = true;
                  };
                };
              };
              # keep-sorted end
            };
          }
        );
        gitHooks = git-hooks.lib.${system}.run {
          src = self;
          hooks = {
            # keep-sorted start block=yes
            actionlint.enable = true;
            ghalint = {
              enable = true;
              name = "ghalint";
              entry = "${pkgs.lib.getExe pkgs.ghalint} run";
              files = "^\\.github/";
              pass_filenames = false;
            };
            golangci-lint = {
              enable = true;
              package = pkgs.golangci-lint;
            };
            treefmt = {
              enable = true;
              packageOverrides.treefmt = treefmt.config.build.wrapper;
            };
            zizmor.enable = true;
            # keep-sorted end
          };
        };
        devPackages = rec {
          # keep-sorted start block=yes
          actions = with pkgs; [
            actionlint
            ghalint
            zizmor
          ];
          # keep-sorted end
          default = [
            treefmt.config.build.wrapper
          ]
          ++ actions;
        };
      in
      {
        # keep-sorted start block=yes
        checks.git-hooks = gitHooks;
        devShells = pkgs.lib.pipe devPackages [
          (pkgs.lib.attrsets.mapAttrs (
            name: buildInputs:
            pkgs.mkShell (
              {
                inherit buildInputs;
              }
              // pkgs.lib.optionalAttrs (name == "default") { inherit (gitHooks) shellHook; }
            )
          ))
        ];
        formatter = treefmt.config.build.wrapper;
        # keep-sorted end
      }
    );
}
