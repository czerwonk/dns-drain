{
  pkgs,
  lib,
  buildGo123Module,
}:

buildGo123Module {
  pname = "dns-drainctl";
  version = "1.0.2";

  src = lib.cleanSource ./.;

  vendorHash = pkgs.lib.fileContents ./go.mod.sri;

  env.CGO_ENABLED = 0;

  meta = with lib; {
    description = "Drain by removing/replacing IP/net from DNS records with ease";
    homepage = "https://github.com/czerwonk/dns-drain";
    license = licenses.mit;
  };
}
