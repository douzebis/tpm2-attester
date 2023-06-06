// SPDX-License-Identifier: Apache-2.0

var port = null;

function sendNativeMessage(text) {
  port.postMessage(text);
  console.log("Sent message: " + JSON.stringify(text));
}

function onDisconnected() {
  console.log("Failed to connect: " + chrome.runtime.lastError.message);
  port = null;
}

var hostName = "com.douzebis.attester";
console.log("Connecting to native messaging host <b>" + hostName + "</b>...")
port = chrome.runtime.connectNative(hostName);
port.onDisconnect.addListener(onDisconnected);

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    (async function () {
        console.log("From content: " + message)
        console.log("From content: " + JSON.stringify(message))

        function onNativeMessage(message) {
            console.log("From native: " + JSON.stringify(message));
            sendResponse(message);
            port.onMessage.removeListener(onNativeMessage)
        }
        port.onMessage.addListener(onNativeMessage);
        sendNativeMessage(message)
    })();
    // return true to indicate you want to send a response asynchronously
    return true;
});