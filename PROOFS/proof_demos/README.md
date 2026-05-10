# UniDB-Go — Gap Proof Demos

## What This Folder Is

Each subfolder contains a **runnable Go program** that **proves a specific gap** exists in Go's `database/sql` — and shows how UniDB-Go fills it.

These are not slide decks or claims. You run them in front of the mentor and she sees the output live.

---

## Setup (Run Once)

```bash
# From this folder (proof_demos/)
go mod init proof_demos
go get github.com/lib/pq
go get github.com/go-sql-driver/mysql
go get go.mongodb.org/mongo-driver/mongo
go get github.com/xwb1989/sqlparser
```

---

## Demo Order (Recommended)

| # | Folder | Gap Proven | Run Command |
|---|--------|-----------|-------------|
| 1 | `gap1_no_unified_api/` | No single API for Postgres + MySQL + MongoDB | `go run gap1_no_unified_api/main.go` |
| 2 | `gap2_no_ast/` | Queries are opaque strings — zero introspection | `go run gap2_no_ast/main.go` |
| 3 | `gap3_no_security/` | No injection detection — runs any query | `go run gap3_no_security/main.go` |
| 4 | `gap4_no_federation/` | Cannot JOIN across two databases | `go run gap4_no_federation/main.go` |
| 5 | `gap5_no_middleware/` | No middleware / interceptor pattern | `go run gap5_no_middleware/main.go` |
| 6 | `gap6_no_error_standard/` | Each driver throws different error types | `go run gap6_no_error_standard/main.go` |
| 7 | `gap7_static_pool/` | Static pool — never adapts to load | `go run gap7_static_pool/main.go` |

---

## What Each Demo Shows

### Gap 1 — No Unified API
- Postgres and MySQL use `sql.Open()` — but with completely different DSN formats
- MongoDB cannot use `database/sql` at all — requires a separate package and different query API
- **Output**: Proves 3 different connection styles, 3 different query patterns

### Gap 2 — No AST / Query Parsing
- `database/sql` passes queries as raw strings directly to the driver
- You cannot extract which tables, columns, joins, or conditions are in a query
- **Output**: Shows that only fragile string operations are possible vs UniDB's structured AST

### Gap 3 — No Security Layer
- `database/sql` executes `SELECT * FROM users WHERE id = 1 OR 1=1` without complaint
- No pattern scanning, no anomaly detection, no exfiltration detection
- **Output**: Shows 5 malicious queries — raw Go sends them all; UniDB blocks/flags them

### Gap 4 — No Cross-DB Federation
- `database/sql` cannot JOIN across Postgres and MySQL in one query
- Manual approach: 2 connections + 2 queries + manual scan + nested loop join = ~60 lines
- **Output**: Shows the 7-step manual boilerplate vs UniDB's 1 query

### Gap 5 — No Middleware Pipeline
- No `db.Use()` hook exists in `database/sql`
- Every security/logging/metrics concern must be manually wrapped at every call site
- **Output**: Shows the copy-paste problem vs UniDB's composable middleware chain

### Gap 6 — No Standardized Errors
- `pq.Error` (Postgres) vs `mysql.MySQLError` (MySQL) — different types, different codes
- Switching databases = rewriting all error handling
- **Output**: Shows duplicate-key error from both drivers; UniDB wraps both into `ErrUniqueViolation`

### Gap 7 — Static Connection Pool
- `db.SetMaxOpenConns(20)` is set once and never changes
- No auto-scale up on load, no auto-scale down on idle, no per-database pools
- **Output**: Live demo of AutoTune goroutine scaling pool up under load and down on idle

---

## What to Say During the Demo

> *"For each gap, I'm going to show you two things: first, what happens in standard Go's database/sql — either it fails, produces wrong output, or requires massive boilerplate. Then I'll show the UniDB-Go equivalent that solves the same problem cleanly. The terminal output is the proof."*
