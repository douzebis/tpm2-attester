// SPDX-License-Identifier: Apache-2.0

function appendMessage(msg) {
    let text = JSON.stringify(msg)
    console.log(text)
    document.getElementById('response').innerHTML += "<p>" + text + "</p>";
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
    if (event.key === "Enter") {
        // Cancel the default action, if needed
        event.preventDefault();
        // Trigger the button element with a click
        sendMsg()
        }
    })
updateUiState()
