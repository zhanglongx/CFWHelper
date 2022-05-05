# CFWHelper

CFWHelper is a helper program to provide some missing features for CFW.

## Background

[CFW](https://github.com/Fndroid/clash_for_windows_pkg/releases) is an excellent proxy software for Windows base on [Clash Core](https://github.com/Dreamacro/clash).

CFW or Clash (Clash Premium) is powerful (tun mode, external control), but it is not open sourced, which makes it very difficult to add some features. For example, I need to be notified when the proxy mode is changed to 'global' for an unexpected long time. 

Fortunately, Clash Core provides an external control API that allows us to achieve such purposes.

## Usage

1. Make Clash Core use the fixed external control port. See: https://github.com/Fndroid/clash_for_windows_pkg/issues/2409

2. Make modifies to the main.go (FIXME: It's no need to use configuration file for now):

    ```go
        // CFW Const
        // Config Url
        URL = "http://127.0.0.1:9090/configs"
    ```

3. Build the binary.

    ```powershell
        PS > go build .
    ```

    OR (to hide cmd windows):

    ```powershell
        PS > go build -ldflags -H=windowsgui .
    ```
4. Run it

    It will output log file (with rotation) under the execute directory.