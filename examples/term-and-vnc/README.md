# Proxmox VNC and Terminal Proxy

This project provides a Go-based server for interacting with Proxmox virtual machines via VNC and terminal proxies. The server supports secure WebSocket connections and can be accessed through a simple web client.

## Getting Started

### Prerequisites

- Go 1.18 or higher
- OpenSSL for generating SSL certificates
- Proxmox environment with a valid configuration

### Installation

1. **Generate SSL Certificates**

   Create self-signed SSL certificates for secure communication:

   ```bash
   openssl req -newkey rsa:2048 -nodes -keyout server.key -x509 -days 365 -out server.crt
   ```

2. **Create the `.env` File**

   Add the following environment variables to a `.env` file in the root directory:

   ```plaintext
   PROXMOX_USERNAME=root@pam
   PROXMOX_PASSWORD=password
   PROXMOX_URL=https://192.168.0.200:8006/api2/json
   PROXMOX_NODE=proxmox5
   PROXMOX_VM=216
   ```

3. **Install Dependencies**

   Install the necessary Go modules:

   ```bash
   go mod tidy
   ```

### Running the Server

Start the server with SSL enabled:

```bash
go run main.go
```

The server will be accessible at `https://localhost:8523`.

### Endpoints

- **`/`**: Returns "hello world".
- **`/term`**: WebSocket endpoint for terminal access.
- **`/vnc`**: WebSocket endpoint for VNC access. Requires a valid ticket.
- **`/vnc-ticket`**: Generates and returns a VNC ticket.

### Code Overview

- **`main.go`**: The main application file that sets up the server and routes.
- **`impl/term.go`**: Contains the implementation for the terminal WebSocket endpoint.
- **`impl/vnc.go`**: Contains the implementation for the VNC WebSocket endpoint and ticket generation.

### Code Overview

- **`main.go`**: The main application file that sets up the server and routes. It configures logging, loads environment variables, and starts the server with SSL.

- **`impl/term.go`**: Contains the implementation for the terminal WebSocket endpoint. It sets up a WebSocket connection for terminal access and handles communication between the client and the Proxmox virtual machine.

- **`impl/vnc.go`**: Contains the implementation for the VNC WebSocket endpoint and ticket generation. It manages VNC connections and tickets:
  - The `tickets` map is used for learning purposes to store VNC tickets temporarily. In a real-world scenario, you should store tickets in a distributed cache or a persistent storage solution.
  - It is also possible to use a password instead of a ticket for authentication.

### Acknowledgments

A big thank you to ChatGPT for helping with the documentation and code improvements!
