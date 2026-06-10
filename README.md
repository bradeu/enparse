<div align="center">

# enparse

**enparse** is an open-source, encrypted, decentralized secret-sharing platform for small teams, enabling local deployments without exposing secrets.

[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Solidity](https://img.shields.io/badge/Solidity-0.8.24-363636?style=flat-square&logo=solidity)](https://soliditylang.org)
[![Network](https://img.shields.io/badge/Network-Sepolia-purple?style=flat-square)](https://sepoliafaucet.com)

</div>

No central server to compromise. Secrets are encrypted client-side with [NaCl box](https://nacl.cr.yp.to/box.html) before leaving your machine and stored as ciphertext on Ethereum. Each team member gets their own encrypted copy so that the chain never sees plaintext.

**Jump to:** [Installation](#installation) · [Owner setup](#setup--owner) · [Member setup](#setup--member)

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

One deployment per team. Multiple projects live inside the same vault.

---

## Installation

Pick your platform and run one command:

```bash
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

<details>
<summary><strong>Build from source</strong></summary>

Requires Go 1.22+.

```bash
cd cli
GOTOOLCHAIN=local go build -o enparse .
mv enparse /usr/local/bin/
```

</details>

---

## Setup — Owner

> Do this once for the whole team. Share your wallet address and the two contract addresses with teammates when done.

### 1. Generate your identity

```bash
enparse init
```

Creates `~/.enparse/identity.json` with your Ethereum address and NaCl keypair.

```bash
enparse status
```

Copy your `0x...` address — you need it in the next step.

### 2. Get free Sepolia ETH

1. Go to the **[Google Web3 Faucet](https://cloud.google.com/application/web3/faucet/ethereum/sepolia)**
2. Sign in with a Google account
3. Paste your `0x...` address
4. Click **Send 0.05 ETH**

Funds arrive in under a minute. Verify at [sepolia.etherscan.io](https://sepolia.etherscan.io) by searching your address.

### 3. Configure RPC

```bash
enparse config set rpc_url https://ethereum-sepolia-rpc.publicnode.com
```

### 4. Deploy contracts

```bash
enparse deploy
```

Deploys `IdentityRegistry` and `ProjectVault` to Sepolia and saves both addresses to `~/.enparse/config.json` automatically.

Run `enparse config show` to see the addresses — share them with your teammates.

### 5. Register on-chain

```bash
enparse register
```

### 6. Create a project and store secrets

```bash
enparse project create myapp

# Add teammates by their Ethereum address (after they complete their setup)
enparse project add myapp 0x<teammate address>

# Store secrets — encrypted per member, written on-chain
enparse set myapp DB_URL=postgres://user:pass@host/db
enparse set myapp API_KEY=sk-...
```

---

## Setup — Member

> Ask your owner for their two contract addresses before starting.

### 1. Generate your identity

```bash
enparse init
```

Creates `~/.enparse/identity.json` with your Ethereum address and NaCl keypair.

```bash
enparse status
```

Copy your `0x...` address — share it with the owner so they can add you to projects.

### 2. Get free Sepolia ETH

1. Go to the **[Google Web3 Faucet](https://cloud.google.com/application/web3/faucet/ethereum/sepolia)**
2. Sign in with a Google account
3. Paste your `0x...` address
4. Click **Send 0.05 ETH**

Funds arrive in under a minute. Verify at [sepolia.etherscan.io](https://sepolia.etherscan.io) by searching your address.

### 3. Configure RPC and contract addresses

```bash
enparse config set rpc_url https://ethereum-sepolia-rpc.publicnode.com
enparse config set identity_registry_addr 0x<address from owner>
enparse config set project_vault_addr 0x<address from owner>
```

### 4. Register on-chain

```bash
enparse register
```

### 5. Use secrets

Once the owner has added you to a project:

```bash
# Inject all secrets as env vars and run your command
enparse run myapp -- go run .
enparse run myapp -- npm start

# Or pull all secrets to a local .env cache
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
| `enparse deploy` | Deploy IdentityRegistry + ProjectVault on-chain (owner only, once) |
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

## License

MIT • [LICENSE](LICENSE)
