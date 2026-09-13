{ pkgs, lib, config, inputs, ... }:

{
  packages = [ pkgs.git ];
  languages.go.enable = true;
  languages.go.version = "1.27.0";
}
