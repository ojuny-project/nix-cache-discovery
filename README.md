# nix-cache-discovery

nix-cache-discovery is a Nix binary cache for a local network. Every host runs
the same Go daemon on the same port; each daemon serves its own Nix store,
discovers the other daemons on configured prefixes, and proxies paths it does
not have locally. It links `libstore` and `libutil` through a C++ shim adapted
from nix-serve-ng. The package targets Linux and the flake builds
`x86_64-linux` and `aarch64-linux`.

## Status

Untested. The package metadata and flake report version 0.1.0. A two-node
virtual machine check exists in `tests/two-node.nix`, but no run is recorded for
this page.

## Requirements

- NixOS with flakes enabled and a running `nix-daemon`.
- One TCP port reachable between hosts; the default is 5000 and every peer must
  use the same value.
- `discovery.ranges` containing the addresses to scan, or `discovery.auto` for
  the subnet of the interface that carries the default route.
- Network access to `fallbackSubstituters` unless the local store and peers
  already hold every path.
- Go, a C++ compiler, and the Nix development headers when building from the
  devshell.

Memory and disk requirements are not stated.

## Install

Add the flake input and enable the module on each host:

```nix
{
  inputs.nix-cache-discovery.url = "github:ojuny-project/nix-cache-discovery";

  outputs = inputs@{ nixpkgs, ... }: {
    nixosConfigurations.host = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        inputs.nix-cache-discovery.nixosModules.default
        {
          services.nix-cache-discovery = {
            enable = true;
            discovery.ranges = [ "192.168.1.0/24" ];
          };
        }
      ];
    };
  };
}
```

Rebuild with `sudo nixos-rebuild switch --flake .#<host>`, where `<host>` is the
configuration name. The module starts the service, creates the
`nix-cache-discovery` system user, and points the host's Nix client at
`http://127.0.0.1:5000`.

Verify the service and its cache information:

```console
$ systemctl is-active nix-cache-discovery
active
$ curl -s http://127.0.0.1:5000/nix-cache-info
StoreDir: /nix/store
WantMassQuery: 1
Priority: 30
```

To build only the daemon, run `nix build` and inspect
`result/bin/nix-cache-discovery`.

## Usage

With the defaults, `exclusive = true` makes this daemon the host's only
substituter. Nix never contacts an upstream cache directly and
`require-sigs = false` applies, so the daemon is the trust boundary. It checks
the local store first, then registered peers, then `fallbackSubstituters`.
Narinfos from fallback caches are verified against `trusted-public-keys` before
they are proxied.

Set `signingKeyFile` and `publicKey` to sign narinfos instead. For a laptop that
leaves the network, list the public addresses of trusted networks in
`trustedNetworks`; outside them the daemon stops discovery and rejects every
non-loopback request with 403.

## Limits

The cache has no user authentication. Any host inside `allowedClients` can read
every path in the local store that the daemon can see. Discovery ranges are
capped at 65536 addresses. TLS is optional, and peer probes accept self-signed
certificates. The default public IP lookup uses a third-party service. Running
with `exclusive = false` removes the signature guarantee the module otherwise
provides, unless a signing key is configured. Windows and macOS are not
supported.

## Documentation

- [nix/module.nix](nix/module.nix) defines every service option and its default.
- [nix/package.nix](nix/package.nix) shows how the daemon is built.
- [tests/two-node.nix](tests/two-node.nix) shows the two-node substitution test.
- [NOTICE](NOTICE) records the nix-serve-ng code and its license.

## Contributing

Open issues at <https://github.com/ojuny-project/nix-cache-discovery/issues>. A patch
must run `go test ./...` and `nix build` from `nix develop`, and must keep the
`NOTICE` credit when it changes the adapted C++ files.

## License

GPL-3.0-only. See [LICENSE](LICENSE). The files
`internal/store/nix.cpp` and `internal/store/nix.h` remain under the BSD
3-clause terms in [NOTICE](NOTICE).
