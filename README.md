# mcp-wire: Zero-Config P2P Mesh for Model Context Protocol (MCP)

`mcp-wire` is an open-source, single-binary daemon written in Go that enables zero-config, peer-to-peer (P2P) distribution of Model Context Protocol (MCP) tools across machines.

By leveraging **go-libp2p** with automatic NAT hole punching (DCUtR), Kademlia DHT peer discovery, and end-to-end **Noise encryption**, `mcp-wire` allows local AI agent environments (**Claude Code**, **OpenCode**, **Cursor**, **Claude Desktop**) to securely discover and invoke MCP servers running on remote devices (home GPU workstations, cloud VMs, edge devices) without public IP addresses, domain names, SSH tunnels, or centralized relay servers.

---

## 🏗 Architecture

```text
┌──────────────────────────────────────────────┐        ┌──────────────────────────────────────────────┐
│ MACHINE A (Laptop / Agent Environment)      │        │ MACHINE B (Remote GPU / Dev Workstation)     │
│                                              │        │                                              │
│  [ AI Agent ]                                │        │                                              │
│  (Claude Code / Cursor / OpenCode)           │        │                                              │
│        │ stdio (JSON-RPC)                    │        │                                              │
│        ▼                                     │        │                                              │
│  [ mcp-wire daemon (Connect Mode) ]          │  P2P   │  [ mcp-wire daemon (Host Mode) ]             │
│   - Peer ID: 12D3KooW...                     │◄──────►│   - Peer ID: 12D3KooX...                     │
│   - libp2p Host + Kademlia DHT               │ Stream │   - Target Exec: `python server.py`          │
│   - E2E Noise Handshake                      │(Noise) │   - ACL & Pre-shared Token Validator         │
└──────────────────────────────────────────────┘        └──────────────────────┬───────────────────────┘
                                                                               │ stdio
                                                                               ▼
                                                                     [ Local Target MCP Server ]
```

---

## ✨ Features

- **Zero-Config P2P Mesh**: Connect behind symmetric NATs and firewalls using `go-libp2p`, STUN, AutoNAT, and Circuit Relay v2.
- **End-to-End Cryptographic Security**: Mandatory Ed25519 node identities (`~/.mcp-wire/identity.key`) and E2E Noise transport encryption.
- **Access Control & Authorization**: Peer whitelist (`~/.mcp-wire/allowed_peers.json`) and pre-shared secret token validation (`--token`).
- **Read-Only Mode Guardrail**: Block mutating MCP requests (`tools/call`) with `-32601 Method not allowed` error responses (`--read-only`).
- **Transparent JSON-RPC Proxying**: Low-latency `<10ms` buffer-pooled stdio proxying supporting streaming and context cancellation.

---

## 📦 Installation

### From Source
```bash
git clone https://github.com/kanqzkokelo/mcp-wire.git
cd mcp-wire
make build
sudo cp mcp-wire /usr/local/bin/
```

### Pre-requisites
- Go 1.23+

---

## 🚀 Quickstart Guide

### 1. Host a Local MCP Tool (Remote Machine)
Expose a local MCP server (e.g. GPU Whisper model or private database tool) over the P2P wire:

```bash
mcp-wire host \
  --name "gpu-whisper" \
  --cmd "python3 server.py" \
  --token "my-secret-key"
```

**Output**:
```text
🚀 mcp-wire daemon running!
   Peer ID: 12D3KooX7...
   Service: /mcp-wire/v1/gpu-whisper
   Connection String: mcp://12D3KooX7.../gpu-whisper?token=my-secret-key
```

### 2. Connect from AI Agent (Local Machine)
Bridge the remote tool directly into your local agent's stdio environment:

```bash
mcp-wire connect mcp://12D3KooX7.../gpu-whisper --token "my-secret-key"
```

---

## ⚙️ AI Client Integration (`claude_desktop_config.json` / `opencode.jsonc`)

Generate a ready-to-use JSON configuration snippet:

```bash
mcp-wire config mcp://12D3KooX7.../gpu-whisper?token=my-secret-key
```

**Output**:
```json
{
  "mcpServers": {
    "gpu-whisper": {
      "command": "mcp-wire",
      "args": [
        "connect",
        "mcp://12D3KooX7.../gpu-whisper?token=my-secret-key",
        "--token",
        "my-secret-key"
      ]
    }
  }
}
```

---

## 🛡 Security & Authorization Management

### Peer Whitelisting (ACL)
Authorize specific remote client Peer IDs:

```bash
# Add peer to whitelist
mcp-wire auth add 12D3KooW...

# List allowed peers
mcp-wire auth list

# Remove peer
mcp-wire auth remove 12D3KooW...
```

### Read-Only Guardrail
Host remote tools in read-only mode to prevent state modifications:

```bash
mcp-wire host --name "database-tool" --cmd "node db.js" --read-only
```

---

## 📊 Diagnostics & Status

Display active P2P connections, bytes transferred, and local node identity:

```bash
mcp-wire status
```

Output as JSON:
```bash
mcp-wire status --json
```

---

## 📄 License
MIT License
