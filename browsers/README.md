# Browsers Support for the Native Messaging API

On recent Ubuntu distributions, the browsers (`chromium`, `firefox`) are packaged via `snap` in a sandboxed mode. This breaks the Native Messaging API that enables extension/host communication.

## Fixing the Firefox installation
Simply install the `xdg-desktop-portal` package as [explained here](https://ubuntuhandbook.org/index.php/2022/11/firefox-ubuntu-22-04-native-messaging/).
```bash
sudo apt-get install xdg-desktop-portal
```

## Fixing the Chromium installation
There is currently no easy way to fix the Chromium Snap installation.
Instead, we need to remove and purge it, and replace with basic installation packages from the debian distribution.
This is [explained here](https://askubuntu.com/questions/1179273/).

- Remove snap's chromium installation
    ```bash
    snap remove chromium
    sudo apt purge chromium-browser chromium-chromedriver
    ```
- Create (e.g. `sudo nano`) an /etc/apt/sources.list.d/debian-stable.list file containing:
    ```
    deb [signed-by=/usr/share/keyrings/debian-archive-keyring.gpg] http://deb.debian.org/debian stable main
    deb-src [signed-by=/usr/share/keyrings/debian-archive-keyring.gpg] http://deb.debian.org/debian stable main

    deb [signed-by=/usr/share/keyrings/debian-archive-keyring.gpg] http://deb.debian.org/debian-security/ stable-security main
    deb-src [signed-by=/usr/share/keyrings/debian-archive-keyring.gpg] http://deb.debian.org/debian-security/ stable-security main

    deb [signed-by=/usr/share/keyrings/debian-archive-keyring.gpg] http://deb.debian.org/debian stable-updates main
    deb-src [signed-by=/usr/share/keyrings/debian-archive-keyring.gpg] http://deb.debian.org/debian stable-updates main
    ```
- Create an /etc/apt/preferences.d/debian-chromium file containing:
    ```
    Explanation: Allow installing chromium from the debian repo.
    Package: chromium*
    Pin: origin "*.debian.org"
    Pin-Priority: 100

    Explanation: Avoid other packages from the debian repo.
    Package: *
    Pin: origin "*.debian.org"
    Pin-Priority: 1
    ```
- Install debian's `chromium` package
    ```bash
    sudo apt update
    sudo apt install chromium
    ```

You should be all set.