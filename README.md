# tpm2-attester

`tpm2-attester` is proof of concept experimenting with performing **TPM remote attestation through the browser**.

Typically a [Remote Attestation framework](https://nedmsmith.github.io/draft-sh-rats-oidc-attest/draft-sh-rats-oidcatt.html) will involve some kind of Authorization Server (AS) reaching back to an Attester endpoint (RA), to retrieve Attestation Evidence.

If implemented as an enhancement to the OIDC code authorization protocol, this mechanism would leverage the User Agent (UA) - the browser on the user's terminal - acting in the capacity of the Attester.

The question becomes: can the browser window (presumably downloaded from the Authorization Server) actually interact with the TPM device on the user's terminal, for retrieving the Attestation Evidence. This is not so easy because typical browsers do not currently support TPM integration (as opposed to FIDO2/Webauthn integration which they support).

This tutorial shows a possible implementation, leveraging an Attester daemon running on the user's terminal:
- The Attester daemon caters for the interaction with the TPM device
- The "authentication" window in the user's browser communicate with the Attester daemon via some browser-extension glue

## Browser / Attester Daemon Glue

The design for glueing browser and daemon consists of:
- The "authentication" window in the user's browser shares its DOM with Javascript code injected by a browser extension[^1], which enables communication via DOM events.
- The injected Javascript code communicates with the extension's service-worker via the [Message Passing API](https://developer.chrome.com/docs/extensions/mv3/messaging/).
- The extension's service-worker communicates with the Attester daemon via the [Native Messaging API](https://developer.chrome.com/docs/extensions/mv3/nativeMessaging/)[^2].

[^1]: The browser extension has to be installed in the user's terminal in the first place.
[^2]: For the extension service-worker to native daemon communication, I have heavily borrowed from [John Farley's mini-project on Github](https://github.com/jfarleyx/chrome-native-messaging-golang).

## Assets used for the Demo

- Ubuntu-desktop guest OS with virtual TPM and measured boot support
- *Non-sandboxed* Chromium browser[^2]
- Golang code for driving the TPM device (adapted from [tpm2-lc](https://github.com/douzebis/tpm2-lc))

[^2]: Sandboxed browsers - such as the SNAP'ed version of chromium or firefox bundled with Ubuntu installation packages - prevent the Native Messaging API from working.

## Installation

### Ubuntu-desktop guest OS

The tutorial uses [jammy-desktop-arm64](https://cdimage.ubuntu.com/jammy/daily-live/current/jammy-desktop-arm64.iso) (22.04 LTS)

The installation is somewhat involved and is described in the [swtpm](./swtpm/README.md) page.

The complete installation procedure and subsequent demo are hosted on the Ubuntu-desktop guest OS.

***/!\ Don't forget to launch the `swtpm` device before you boot the Guest OS.***

### Non-sandboxed browser

A non-official [snap-free chromium package](https://fosspost.org/chromium-deb-package-ubuntu-22-04/) for Ubuntu 22.04 is being made available by `fosspost.org` with installation instructions.

```bash
sudo add-apt-repository ppa:saiarcot895/chromium-beta
sudo apt remove chromium-browser
sudo snap remove chromium
sudo apt install chromium-browser
```

Alternatively it may be possible to build chromium from source; I have to say this is probably long/complex.

### Golang >= 1.19

The Golang code for driving the TPM requires Golang 1.19, which is not installed by default with Ubuntu 22.04.

Install a more recent version as described by [techadmin.net](https://tecadmin.net/how-to-install-go-on-ubuntu-20-04/)
```bash
wget  https://go.dev/dl/go1.20.2.linux-arm64.tar.gz 
sudo tar -xvf go1.20.2.linux-arm64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/mes:/usr/local/games
go version
```

### Build an Start the Demo

#### TPM Ownership
On the host:
```bash
git clone https://github.com/douzebis/tpm2-attester.git
cd tpm2-attester/
(cd device && go build -o attester src/attest/main.go)
(cd device && go build -o init src/init/main.go)
```

Take ownership of the TPM and record the TPM/PCR reference values:
(make sure your `$USER` has access to `/dev/tpmrm0` and `/sys/kernel/security/tpm0/binary_bios_measurements`)
```bash
(cd device && ./init --alsologtostderr -v 5)
```

#### Chromium Extension
Install the chromium extension:
```bash
(cd device && ./install-dev.sh chromium)
```
Then:
- open `chromium`
- navigate to `chrome://extensions`
- activate `Developer mode` switch
- click `Load unpacked`
- select the `tpm2-attester/chrome` directory and load the extension

/!\ You may have to update the extension ID (e.g. `ID: omgdoephiiagfflfpomaobelgfflidha`) in the `device/com.douzebis.attester.json` extension config file. If this is the case, unload the extensions, fix the config file and redo the steps from `./install-dev.sh` on.

#### HTTP Server
Launch an HTTP server for the "authentication" webpage.
```bash
(cd chrome/ && python3 -m http.server &)
```

### Run the Demo

- In `chromium` navigate to `http://localhost:8000/index.html`:
<img src="./attest-1.png">
- click `[Get TPM quote]`: the browser window (i.e. the Attester) retrieves a PCR quote from the TPM daemon (the list of `PCRs` and `Nonce` are chosen by the Verifier)
<img src="./attest-2.png">
- Click `[Get AK pub]`: the browser window shows the Attestation Key that was registered with the Verifier in a provisioning step
<img src="./attest-3.png">
- Click `[Verify]`: the Verifier checks the PCR quote... all good.
<img src="./attest-4.png">
- Tamper (say) with the TPM Signature, and click `[Verify]` again... this time the Verifier complains!
<img src="./attest-5.png">
