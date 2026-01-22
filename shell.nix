{ pkgs ? import <nixpkgs> {} }:
  pkgs.mkShell {
    # nativeBuildInputs is usually what you want -- tools you need to run
    nativeBuildInputs = with pkgs; [ 
        nodejs
        go
        glib
        electron-bin
    ];

    shellHook = let 
        runtimeLibraries = with pkgs; [
            glib
            nspr
            nss
            dbus
            atk
            cups
            cairo
            gtk3
            pango
            libx11
            libxcomposite
            libxdamage
            libxext
            libxfixes
            libxrandr
            libgbm
            expat
            libxcb
            libxkbcommon
            alsa-lib
        ];
    in ''
        export LD_LIBRARY_PATH=${pkgs.lib.makeLibraryPath runtimeLibraries}:LD_LIBRARY_PATH
        export ELECTRON_OVERRIDE_DIST_PATH=${pkgs.electron}/bin
        export ELECTRON_DISABLE_SANDBOX=1
    '';
}

