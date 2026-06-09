<div align="center">

# enparse

**enparse** is an open-source, encrypted, decentralized secret-sharing platform for small teams, enabling local deployments without exposing secrets.

[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.24-363636?style=flat-square&logo=solidity)](https://soliditylang.org)
[![Network](https://img.shields.io/badge/Network-Sepolia-purple?style=flat-square)](https://sepoliafaucet.com)

</div>

No central server to compromise. Secrets are encrypted client-side with [NaCl box](https://nacl.cr.yp.to/box.html) before leaving your machine and stored as ciphertext on Ethereum. Each team member gets their own encrypted copy so that the chain never sees plaintext.

---

## How it works

```
Owner                                     Teammate
  │                                          │
  enparse init                               enparse init
  enparse register                           enparse register
  │                                          │
  enparse project create myapp               │
  enparse project add myapp 0x<teammate>     │
  enparse set myapp DB_URL=postgres://...    │
  │  (encrypts per-member, writes on-chain)  │
  │                                          │
  │                                          enparse run myapp -- go run .
  │                                          (fetches + decrypts → env vars → exec)
```

**Contracts (Sepolia):**

| Contract | Purpose |
|----------|---------|
| `IdentityRegistry` | Maps each Ethereum address to a NaCl public key |
| `ProjectVault` | Stores per-member encrypted blobs keyed by project + secret name + address |

Only the project owner can write secrets or add members. Any member can read their own blobs.

---

## Prerequisites

- Go 1.22+
- Sepolia RPC URL — use the free public endpoint `https://ethereum-sepolia-rpc.publicnode.com`
- Sepolia ETH for each user — free from the [Google Web3 Faucet](https://cloud.google.com/application/web3/faucet/ethereum/sepolia)
- Node.js 18+ and npm — contracts deployment

---

## Install



```bash
# YOLO

# macOS (Apple Silicon)
curl -L https://github.com/bradeu/enparse/releases/latest/download/enparse-darwin-arm64 -o enparse
chmod +x enparse && mv enparse /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/bradeu/enparse/releases/latest/download/enparse-darwin-amd64 -o enparse
chmod +x enparse && mv enparse /usr/local/bin/

# Linux
curl -L https://github.com/bradeu/enparse/releases/latest/download/enparse-linux-amd64 -o enparse
chmod +x enparse && mv enparse /usr/local/bin/
```

**Build from source:**

```bash
cd cli
GOTOOLCHAIN=local go build -o enparse .
mv enparse /usr/local/bin/
```

> `GOTOOLCHAIN=local` prevents Go 1.21+'s toolchain manager from auto-bumping the `go` directive in `go.mod`.

---

## Quick Start

### 1 — Deploy contracts (owner only, once)

```bash
cd contracts
npm install

# Point at Sepolia
enparse config set rpc_url https://ethereum-sepolia-rpc.publicnode.com

# Deploy — reads your identity key, writes contract addresses back to ~/.enparse/config.json
npm run deploy
```

### 2 — Set up identity (every user)

```bash
# Generate Ethereum + NaCl keypair, saved to ~/.enparse/identity.json
enparse init

# Confirm your address and config
enparse status

# Fund your address with Sepolia ETH, then register on-chain
enparse register
```

### 3 — Create a project and store secrets (owner)

```bash
enparse project create myapp
enparse project add myapp 0x<teammate address>

enparse set myapp API_KEY=... # Insert secrets 
```

### 4 — Use secrets (any member)

```bash
# Inject all secrets as env vars and exec your command
enparse run myapp -- go run .
enparse run myapp -- npm start

# Or fetch to a local cache file
enparse pull myapp    # writes ~/.enparse/cache/myapp.env

# Or read a single value
enparse get myapp DB_URL
```

---

## Configuration

Config lives at `~/.enparse/config.json`. Set values with `enparse config set` or via environment variables.

| Key | Env var | Description |
|-----|---------|-------------|
| `rpc_url` | `ENPARSE_RPC_URL` | Sepolia RPC endpoint |
| `identity_registry_addr` | `ENPARSE_REGISTRY_ADDR` | Deployed `IdentityRegistry` address |
| `project_vault_addr` | `ENPARSE_VAULT_ADDR` | Deployed `ProjectVault` address |

```bash
enparse config show   # print current values
```

---

## Command Reference

| Command | Description |
|---------|-------------|
| `enparse init [--force]` | Generate identity (NaCl + Ethereum keypair) |
| `enparse register` | Register NaCl public key on-chain (you pay gas) |
| `enparse status` | Show identity, config, and chain registration status |
| `enparse config set <key> <value>` | Set a config value |
| `enparse config show` | Print current config |
| `enparse project create <name>` | Create project (caller becomes owner) |
| `enparse project add <project> <address>` | Add registered member; auto-re-encrypts existing secrets |
| `enparse project list` | List projects you belong to |
| `enparse set <project> KEY=value` | Encrypt and store secret for all members (owner only) |
| `enparse get <project> KEY` | Fetch and decrypt a single secret |
| `enparse pull <project>` | Fetch all secrets → `~/.enparse/cache/<project>.env` |
| `enparse run <project> -- <cmd>` | Inject secrets as env vars and exec command |

---

## Security Model

- **Client-side encryption.** Secrets are encrypted with NaCl box (Curve25519 + XSalsa20 + Poly1305) before leaving your machine. The chain stores only ciphertext.
- **Per-member copies.** Each member gets their own encrypted blob. Decryption requires their private NaCl key, which never leaves `~/.enparse/identity.json`.
- **Append-only chain.** Old secret versions persist in transaction history. Rotate secrets after removing a member.
- **No process persistence.** `enparse run` uses `syscall.Exec` — secrets live only in the child process environment and are never written to disk.

---

## Repository Layout

```
enparse/
├── contracts/                  Solidity contracts + Hardhat config
│   ├── contracts/
│   │   ├── IdentityRegistry.sol
│   │   └── ProjectVault.sol
│   ├── scripts/deploy.ts       Deploys to Sepolia, updates ~/.enparse/config.json
│   └── test/                   36 Hardhat tests
└── cli/                        Go CLI — single binary after go build
    ├── cmd/                    Cobra commands
    └── internal/
        ├── chain/              go-ethereum contract bindings
        ├── config/             ~/.enparse/config.json
        ├── crypto/             NaCl box encrypt/decrypt
        └── identity/           ~/.enparse/identity.json
```

---

## License

MIT • [LICENSE](LICENSE)
