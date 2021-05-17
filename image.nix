{ pkgs ? import ./pkgs.nix { } }:

let
  package = import ./default.nix { };
in
pkgs.dockerTools.buildLayeredImage {
  name = "thmzlt/querybench";
  tag = "latest";

  config = {
    Cmd = [ "${package}/bin/querybench" ];
  };
}
