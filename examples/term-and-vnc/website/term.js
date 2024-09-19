import { Terminal } from "@xterm/xterm";
import { AttachAddon } from "@xterm/addon-attach";

const term = new Terminal({
  cursorBlink: true,
  theme: {
    foreground: "#00ff00",
    background: "#000000",
    cursor: "#00ff00",
  },
});
term.open(document.getElementById("terminal"));

const url = "wss://term-and-vnc.toquinha.online/term";
const socket = new WebSocket(url);
const attachAddon = new AttachAddon(socket);
term.loadAddon(attachAddon);

// term.onKey(key => {
//   socket.send(key.key);
// });
