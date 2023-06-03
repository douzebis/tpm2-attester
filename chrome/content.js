// SPDX-License-Identifier: Apache-2.0

function appendMessage(msg) {
    let text = JSON.stringify(msg)
    console.log(text)
    document.getElementById('response').value += "<p>" + text + "</p>";
}

function updateUiState() {
    //document.getElementById('connect-button').style.display = 'none'
    document.getElementById('input-text').style.display = 'inline-block'
    document.getElementById('send-message-button').style.display = 'inline-block'
}

function sendMsg() {
    //message = {"query": document.getElementById('input-text').value.trim()}
    let text = document.getElementById('input-text').value
    console.log(text)
    let message = JSON.parse(text.trim())
    console.log(message)
    chrome.runtime.sendMessage(message, (response) => {
        appendMessage(response)
        console.log(response)
    });
}

document.getElementById('send-message-button').addEventListener('click', sendMsg)
document.getElementById('input-text').addEventListener('keypress', function(event) {
    // If the user presses the "Enter" key on the keyboard
    if (event.key === "Enter" && event.getModifierState("Control")) {
        // Cancel the default action, if needed
        event.preventDefault();
        // Trigger the button element with a click
        sendMsg()
    }
})

document.getElementById('get-ak-button').addEventListener('click', function() {
    let query = {
        "query": "get-ak-pub"
    }
    chrome.runtime.sendMessage(query, (response) => {
        document.getElementById('ak-text').value = response['ak-pub']
    });
})

document.getElementById('get-tpm-quote-button').addEventListener('click', function() {
    let query = {
        "query": "get-tpm-quote",
        "pcrs": JSON.parse(document.getElementById('pcrs-text').value)
    }
    chrome.runtime.sendMessage(query, (response) => {
        document.getElementById('nonce-text').value = JSON.stringify(response['nonce'])
        document.getElementById('tpm-attestation-text').value = JSON.stringify(response['attestation'])
        document.getElementById('tpm-signature-text').value = JSON.stringify(response['signature'])
    });
})

document.getElementById('verify-tpm-quote-button').addEventListener('click', function() {
    let query = {
        "query": "verify-tpm-quote",
        "pcrs": JSON.parse(document.getElementById('pcrs-text').value),
        "nonce": JSON.parse(document.getElementById('nonce-text').value),
        "attestation": JSON.parse(document.getElementById('tpm-attestation-text').value),
        "signature": JSON.parse(document.getElementById('tpm-signature-text').value),
        "ak-pub": document.getElementById('ak-text').value
    }
    chrome.runtime.sendMessage(query, (response) => {
        if (response['is-legit']) {
            document.getElementById('verification').value = "OK"
        } else {
            document.getElementById('verification').value = "/!\\/!\\/!\\ KO /!\\/!\\/!\\"
        }
        document.getElementById('tpm-message-text').value = response['message']
    });
})

updateUiState()
