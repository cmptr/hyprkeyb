{
  description = "hyprkeyb — Hyprland-aware keybind cheatsheet overlay";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forEachSystem = f: nixpkgs.lib.genAttrs systems
        (system: f nixpkgs.legacyPackages.${system});
    in {
      packages = forEachSystem (pkgs: {
        default = pkgs.buildGoModule {
          pname = "hyprkeyb";
          version = self.shortRev or "dev";
          src = ./.;
          vendorHash = "sha256-fQkYYrMC48NbDW0yqNuV1VAC0XbRIQad7Iy2/P8Yubw=";
          ldflags = [ "-s" "-w" "-X main.version=${self.shortRev or "dev"}" ];
          meta = {
            description = "Hyprland-aware keybind cheatsheet overlay";
            mainProgram = "hyprkeyb";
          };
        };
      });

      devShells = forEachSystem (pkgs: {
        default = pkgs.mkShell {
          packages = [
            (if pkgs ? go_1_26 then pkgs.go_1_26 else pkgs.go)
            pkgs.gopls
            pkgs.gotools
            pkgs.golangci-lint
            pkgs.hyprland
          ];
        };
      });
    };
}
