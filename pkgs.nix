{ pkgsRef ? "f4e108408ffe2a56ae91857009aaa1f7352351a6" }:

import (fetchTarball "https://github.com/nixos/nixpkgs/archive/${pkgsRef}.tar.gz") {}
