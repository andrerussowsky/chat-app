document.addEventListener("DOMContentLoaded", () => {
    const messagesContainer = document.getElementById("messages");
    const messageInput = document.getElementById("message");
    const sendButton = document.getElementById("send");

    const socket = new WebSocket("ws://localhost:8080/ws");
    var token = "{{ .Token }}";
    var teste = "{{ .Teste }}";

    sendButton.addEventListener("click", () => {
        const message = messageInput.value;
        if (message.trim() !== "") {
            console.log(token)
            console.log(teste)
            console.log("{{ .Name }}")
            console.log("{{ . }}")
            socket.send(JSON.stringify({ token, content: message }));
            messageInput.value = "";
        }
    });

    socket.addEventListener("message", (event) => {
        const message = JSON.parse(event.data);
        messagesContainer.innerHTML += `<p><strong>${message.username}:</strong> ${message.content}</p>`;
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    });
});

console.log("{{ .Name }}")

window.onload = function() {
    console.log("{{ .Name }}")
};