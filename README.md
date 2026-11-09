# nix-cache-discovery

nix-cache-discovery is a Nix binary cache proxy with automated local network sharing. Every host discovers other hosts running nix-cache-discovery on the same host; each daemon serves its own Nix store, discovers the other daemons on configured prefixes, and proxies paths it does not have locally. It links `libstore` and `libutil` through a C++ shim adapted from nix-serve-ng. 

nix-cache-discovery is explicitly limited to local addresses only, and has a public-ip filter to prevent cache sharing on unsafe networks (in which case, nix-cache-discovery will still run, but will only use the fallback substituters)

nix-cache-discovery verifies the signatures and hashes of the upstream cache to prevent MITM attacks, just like nix does by itself. 

## Status

Tested functional.

## Requirements

- NixOS with flakes enabled and a running `nix-daemon`.
- One TCP port reachable between hosts (default port 5000); all peers must use the same port to discover eachother and multiple independent sharing meshes can be set up on one network using multiple ports.
- `discovery.ranges` containing the addresses to scan, or `discovery.auto` to automatically configure the range to the local subnet on-connect. 
- Network access to `fallbackSubstituters` unless the local store and peers already hold every path.


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

Rebuild the system afterwards, The module starts the service, creates the
`nix-cache-discovery` system user, and points the host's Nix client at the local nix-cache-discovery cache. 

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

By default, nix-cache-discovery is set up to make itself the *only* substituter cache on your system. this allows it to ensure local caches are always prioritized over public caches, and gives it more control over how to handle timeouts in the upstream substituter caches. It will always check the local store first, then local peers, then the fallback cache. 

the daemon will disable nix's internal signature enforcement by default, allowing the daemon to handle trust. signature enforcement can be re-enabled with require-sigs = true; which will then require each peer to sign its cache with a trusted key. this needs explicit setup per peer, but is significantly more secure and recommended for professional or business use (alongside using an internal nix-cache-ng server as your primary fallback substituter)

the trustedNetworks key allows you to set specific public IPs on which nix-cache-discovery will activate its discovery functionality, if your device is on a different network with this configuration option set, it will not contact any other device on the network or run scans. 


## Limits

The cache has no user authentication. Any host inside `allowedClients` can read
every path in the local store that the daemon can see. Discovery ranges are
capped at 65536 addresses. TLS is not included in nix-cache-discovery, and peer probes accept self-signed
certificates. The default public IP lookup uses a third-party service (which to use is configurable in the options). Running
with `exclusive = false` and require-sigs = false can result in insecure cache access to regular substituters (not configured in nix-cache-discovery's settings) as they can be accessed and their signatures are not checked, please use require-sigs = true if you want to use exclusive = false outside of temporary troubleshooting. Windows and macOS are not supported.

## Releases

A `v*` tag, a manual workflow run, or a commit on `main` whose title starts with
a version such as `v0.1.0` builds `nix-cache-discovery.bin` and `nix-cache-discovery-flake.tar.gz`.

The release flake installs the prebuilt bundle without compiling:

```nix
inputs.nix-cache-discovery.url =
  "tarball+https://github.com/ojuny-project/nix-cache-discovery/releases/download/vX.Y.Z/nix-cache-discovery-flake.tar.gz";
```

Then add `inputs.nix-cache-discovery.packages.${pkgs.system}.nix-cache-discovery` to `environment.systemPackages`.

The generated flake exposes packages only. Use the source flake when you need
the NixOS module.


## Documentation
- [OPTIONS.md](OPTIONS.md) - generated options reference.
- [NOTICE](NOTICE) - records the nix-serve-ng code and its license.

## Contributing

Open issues at <https://github.com/ojuny-project/nix-cache-discovery/issues>. A patch
must run `go test ./...` and `nix build` from `nix develop`, and must keep the
`NOTICE` credit when it changes the adapted C++ files.

## License

GPL-3.0-only. See [LICENSE](LICENSE). The files
`internal/store/nix.cpp` and `internal/store/nix.h` remain under the BSD
3-clause terms in [NOTICE](NOTICE).

