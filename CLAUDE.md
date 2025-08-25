# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

udppunch is a UDP hole punching tool for WireGuard, inspired by natpunch-go. It consists of:

- **Server** (`server/server.go`): UDP server that maintains a registry of peers and facilitates peer discovery
- **Client** (`client/client.go`): Client that connects to WireGuard interface and updates peer endpoints through the server

## Architecture

The project uses a simple client-server architecture:

1. **Core Types** (`data.go`): Defines `Key` (32-byte WireGuard public key) and `Peer` (38-byte structure with key + IP + port)
2. **Protocol** (`const.go`): Two message types - HandshakeType (0x01) for peer registration, ResolveType (0x02) for peer discovery
3. **Server**: Maintains LRU cache of peers, handles handshake and resolve requests
4. **Client**: Integrates with WireGuard interface via `wg` commands, performs handshakes and endpoint resolution

## Build Commands

```bash
# Build for local platform
make build

# Build for all platforms (cross-compilation)
make build_all

# Clean build artifacts
make clean

# Build everything (clean + build_all)
make all
```

Built binaries are placed in `dist/` directory with naming pattern: `punch-{server|client}-{os}-{arch}`

## WireGuard Integration

- **WireGuard Module** (`client/wg/`): Wraps `wg` command-line tool for interface operations
- **Network Module** (`client/netx/`): Handles UDP socket operations for handshake packets
- Client requires WireGuard interface to be up before running
- Uses `wg show` commands to get interface info and `wg set` to update peer endpoints

## Usage

Server: `./punch-server-linux-amd64 -port 19993`
Client: `./punch-client-linux-amd64 -server xxxx:19993 -iface wg0`

The client performs handshakes every 25 seconds and can run in continuous mode for ongoing peer resolution.