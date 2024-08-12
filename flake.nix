{
  inputs = {
    nix-go = {
      inputs.nixpkgs.follows = "nixpkgs";
      url = "github:matthewdargan/nix-go";
    };
    nixpkgs.url = "nixpkgs/nixos-unstable";
    parts.url = "github:hercules-ci/flake-parts";
    pre-commit-hooks = {
      inputs.nixpkgs.follows = "nixpkgs";
      url = "github:cachix/pre-commit-hooks.nix";
    };
  };
  outputs = inputs:
    inputs.parts.lib.mkFlake {inherit inputs;} {
      imports = [inputs.pre-commit-hooks.flakeModule];
      flake.homeModules.mooch = {
        config,
        lib,
        pkgs,
        ...
      }: let
        cfg = config.mooch;
      in {
        config = lib.mkIf cfg.enable {
          systemd.user.services.mooch = {
            Install.WantedBy = ["default.target"];
            Service = {
              ExecStart = "${cfg.package}/bin/mooch";
              Type = "oneshot";
            };
            Unit.Description = "Download and organize torrents from RSS feeds";
          };
          systemd.user.timers.mooch = {
            Install.WantedBy = ["timers.target"];
            Timer = {
              OnCalendar = cfg.timer.onCalendar;
              Persistent = "true";
            };
            Unit.Description = "mooch.service";
          };
        };
        options.mooch = {
          enable = lib.mkEnableOption "Enable mooch service";
          package = lib.mkOption {
            default = inputs.self.packages.${pkgs.system}.mooch;
            type = lib.types.package;
          };
          timer = {
            enable = lib.mkEnableOption "Enable mooch timer";
            onCalendar = lib.mkOption {
              default = "daily";
              type = lib.types.str;
            };
          };
        };
      };
      perSystem = {
        config,
        inputs',
        lib,
        pkgs,
        ...
      }: {
        devShells.default = pkgs.mkShell {
          packages = [inputs'.nix-go.packages.go inputs'.nix-go.packages.golangci-lint];
          shellHook = "${config.pre-commit.installationScript}";
        };
        packages.mooch = inputs'.nix-go.legacyPackages.buildGoModule {
          meta = with lib; {
            description = "Download and organize torrents from RSS feeds";
            homepage = "https://github.com/matthewdargan/mooch";
            license = licenses.bsd3;
            maintainers = with maintainers; [matthewdargan];
          };
          pname = "mooch";
          src = ./.;
          vendorHash = "sha256-DrTKoivInvri6QLOTG39Uus58LQceV8/IRwkhWpjdfg=";
          version = "0.1.0";
        };
        pre-commit = {
          check.enable = false;
          settings = {
            hooks = {
              alejandra.enable = true;
              deadnix.enable = true;
              golangci-lint = {
                enable = true;
                package = inputs'.nix-go.packages.golangci-lint;
              };
              gotest = {
                enable = true;
                package = inputs'.nix-go.packages.go;
              };
              statix.enable = true;
            };
            src = ./.;
          };
        };
      };
      systems = ["aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux"];
    };
}
