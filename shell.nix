{ pkgs ? import ./pkgs.nix { } }:
let
  gopls = pkgs.buildGoModule rec {
    pname = "gopls";
    version = "0.6.8";

    src = pkgs.fetchgit {
      url = "https://go.googlesource.com/tools";
      rev = "gopls/v${version}";
      sha256 = "KPAo+s8zDN3wkxGu/k91m/ahHF4Ephgbj/KQxdsm0Mk=";
    };

    modRoot = "gopls";
    subPackages = [ "." ];

    vendorSha256 = "7E9AwGYyIJisBuD9kx4OX7YOdI/ZDaKi0bTJPaCYqOA=";

    doCheck = false;
  };
in
(pkgs.callPackage ./default.nix { }).overrideAttrs (
  attrs: {
    src = null;

    nativeBuildInputs = attrs.nativeBuildInputs ++ [
      gopls
      pkgs.go_1_16
      pkgs.golint
      pkgs.postgresql_13
    ];

    shellHook = ''
      export GO111MODULE="on"
      export GOCACHE=""
      export GOPATH="$(pwd)/.go"
      export PATH=$PATH:$GOPATH/bin

      set +v
    '';
  }
)
