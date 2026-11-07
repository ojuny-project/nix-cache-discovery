{
  description = "In-home Nix binary cache with LAN auto-discovery";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      forEachSystem = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      binaryInfo =
        if builtins.pathExists ./binary-url.json then
          builtins.fromJSON (builtins.readFile ./binary-url.json)
        else
          null;

      mkBinary =
        {
          system,
          tag ? (
            if binaryInfo == null then
              throw "nix-cache-discovery: binary-url.json is missing"
            else
              binaryInfo.latest
          ),
          package ? "nix-cache-discovery",
        }:
        let
          release =
            if binaryInfo == null then
              throw "nix-cache-discovery: binary-url.json is missing"
            else
              binaryInfo.releases.${tag}
                or (throw "nix-cache-discovery: release '${tag}' is not recorded in binary-url.json");
          packageInfo =
            release.packages.${package}
              or (throw "nix-cache-discovery: package '${package}' is not recorded in release '${tag}'");
          pkgs = import nixpkgs { inherit system; };
        in
        pkgs.callPackage ./nix/binary.nix {
          name = package;
          program = packageInfo.program or package;
          inherit (packageInfo) file;
          info = {
            inherit (binaryInfo) repo system;
            inherit (release) tag version;
            inherit (packageInfo) url hash;
          };
        };
    in
    {
      packages = forEachSystem (
        pkgs:
        rec {
          default = pkgs.callPackage ./nix/package.nix { };
          nix-cache-discovery = default;
        }
        //
          nixpkgs.lib.optionalAttrs
            (binaryInfo != null && binaryInfo.system == pkgs.stdenv.hostPlatform.system)
            {
              nix-cache-discovery-binary = mkBinary { system = pkgs.stdenv.hostPlatform.system; };
            }
      );

      devShells = forEachSystem (pkgs: {
        default = pkgs.callPackage ./nix/devshell.nix { };
      });

      lib.mkBinary = mkBinary;
      nixosModules.default = import ./nix/module.nix;
      nixosModules.nix-cache-discovery = self.nixosModules.default;

      checks = forEachSystem (pkgs: {
        package = self.packages.${pkgs.system}.default;
        vm = import ./tests/two-node.nix { inherit pkgs; };
      });
    };
}
