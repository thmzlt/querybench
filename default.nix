{ pkgs ? import ./pkgs.nix { } }:

pkgs.buildGoModule {
  pname = "querybench";
  version = "0.0.1";
  src = builtins.path { path = ./.; name = "querybench"; };

  vendorSha256 = null;

  CGO_ENABLED = "0";
}
