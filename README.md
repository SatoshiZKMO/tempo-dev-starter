# Tempo Stablecoin Analytics Engine

A persistent, multi-token, fee-aware analytics engine built for the Tempo Testnet.

This project listens to on-chain ERC-20 Transfer events and computes real-time and lifetime payment metrics for stablecoins on Tempo.

---

## Why This Exists

Tempo is a payment-first blockchain with:

- No native gas token
- Stablecoin-based transaction fees
- Deterministic fee behavior

Instead of just sending transactions, this engine analyzes:

- Total stablecoin volume
- Total fees paid
- Effective fee rate
- Per-token lifetime metrics

---

## Features

- Multi-token support (pathUSD, AlphaUSD, BetaUSD, ThetaUSD)
- Real-time block listener
- Persistent state (block tracking)
- Persistent lifetime analytics (JSON-based)
- Fee-aware event parsing
- Global protocol metrics
- Effective fee rate calculation

---

## Architecture Overview

1. Poll latest block
2. Scan new block range
3. Filter ERC-20 Transfer events
4. Separate:
   - Incoming transfers
   - Outgoing transfers
   - Fee transfers
5. Update lifetime stats
6. Compute global metrics

---

## Example Output

==== Lifetime Stats: AlphaUSD ====
Transfers: 9
Incoming: 0
Outgoing: 4.5
Fees Paid: 0.0045
===== GLOBAL METRICS =====
Total Transfers: 9
Total Volume: 4.5
Total Fees: 0.0045
Effective Fee Rate: 0.1000%


---

## What This Demonstrates

- Stablecoin-native gas accounting
- Deterministic fee modeling (~0.1%)
- Event-level economic analysis
- Payment flow validation

---

## How To Run

cd go-client
go run main.go

Replace the wallet address inside `main.go` with your own.

---

## Future Extensions

- WebSocket subscription instead of polling
- SQLite persistence
- REST API endpoint
- Prometheus metrics export
- Per-token fee comparison
- Dashboard visualization

---

## Built For

Developers exploring:

- Payment-first blockchains
- Stablecoin gas models
- On-chain analytics
- Event-driven infrastructure
