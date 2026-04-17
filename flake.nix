{
  description = "keyb dev environment";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forEachSystem = f: nixpkgs.lib.genAttrs systems
        (system: f nixpkgs.legacyPackages.${system});
    in {
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
