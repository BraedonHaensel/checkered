{
    description = "GEARS Game";

    inputs.nixpkgs.url = "github:nixos/nixpkgs/0cd7045799ff794bc9393c5ef94bda516f6cb0fc";

    outputs = { self, nixpkgs }:
        let
            system = "x86_64-linux";
            pkgs = import nixpkgs { inherit system; };
        in {
            devShells.${system}.default = import ./shell.nix { inherit pkgs; };
        };
}
