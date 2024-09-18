import RFB from "@novnc/novnc";

const connectButton = document.getElementById("connectButton");
const disconnectButton = document.getElementById("disconnectButton");
const statusElement = document.getElementById("status");
let rfb;

connectButton.addEventListener("click", () => {
  const url = "https://term-and-vnc.toquinha.online/vnc-ticket";
  const wss = "https://term-and-vnc.toquinha.online/vnc";

  fetch(url)
    .then((response) => response.json())
    .then((data) => {
      const ticket = data.ticket;
      connect(`${wss}?ticket=${ticket}`, atob(ticket));
    });

  statusElement.textContent = "Conectando ao VNC Parte 1";
});

disconnectButton.addEventListener("click", () => {
  if (rfb) {
    rfb.disconnect();
    statusElement.textContent = "Desconectado manualmente";
  }
});

function connect(url, password) {
  statusElement.textContent = "Conectando ao VNC Parte 2";
  // Inicia a conexão VNC
  rfb = new RFB(document.getElementById("vncContainer"), url, {
    shared: false,
    repeaterID: false,
    credentials: {
      password: password, // Substitua pela senha do VNC
    },
  });

  rfb.addEventListener("connect", () => {
    statusElement.textContent = "Conectado ao VNC";
  });

  rfb.addEventListener("disconnect", () => {
    statusElement.textContent = "Desconectado do VNC";
  });

  rfb.addEventListener("credentialsrequired", () => {
    alert("Credenciais necessárias");
    // Você pode solicitar credenciais aqui se ainda não forneceu
  });

  rfb.addEventListener("securityfailure", (e) => {
    statusElement.textContent = `Falha de segurança: ${e.detail.reason}`;
  });
}
