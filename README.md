# harness-store

SQLite registry for harness instances and credential bindings. Part of the [llm-bridge](https://github.com/kayushkin/llm-bridge) ecosystem.

Tracks which harness instances are deployed across machines (local or SSH), which credentials are bound to each instance, and display metadata for harness types. This is configuration storage only -- runtime state (active sessions, credential slot occupancy) lives in [llm-bridge-server](https://github.com/kayushkin/llm-bridge-server).

## Install

```bash
go get github.com/kayushkin/harness-store
```

## Usage

```go
import store "github.com/kayushkin/harness-store"

// Open (creates DB and runs migrations if needed)
s, err := store.Open("~/.config/harness-store/harness.db")
defer s.Close()

// Register a harness instance
s.CreateInstance(&msg.Instance{
    ID:          "laptop-cc",
    HarnessType: msg.HarnessClaudeCode,
    Name:        "laptop-cc",
    Transport:   msg.TransportLocal,
})

// Bind a credential to it
s.BindCredential(&msg.InstanceCredential{
    InstanceID:   "laptop-cc",
    CredentialID: "anthropic-key-1",
    Priority:     0, // primary
    MaxConcurrent: 2,
})

// Query
instances, _ := s.ListInstancesByHarness(msg.HarnessClaudeCode)
creds, _ := s.ListInstanceCredentials("laptop-cc")
totalSlots, _ := s.TotalMaxConcurrent("laptop-cc")
```

## Schema

### `instances`

Harness deployments on specific machines.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PK | Unique instance identifier |
| `harness_type` | TEXT | Harness type (e.g. `claude_code`, `codex`) |
| `name` | TEXT | Human-readable name |
| `host` | TEXT | Hostname (default: `localhost`) |
| `transport` | TEXT | `local` or `ssh` |
| `ssh_user` | TEXT | SSH username |
| `ssh_key_path` | TEXT | Path to SSH private key |
| `ssh_port` | INTEGER | SSH port (default: 22) |
| `working_dir` | TEXT | Working directory |
| `max_concurrent_sessions` | INTEGER | Max concurrent sessions (default: 1) |
| `enabled` | INTEGER | 0/1 flag (default: 1) |

### `instance_credentials`

Binds credentials to instances with priority and concurrency limits.

| Column | Type | Description |
|--------|------|-------------|
| `instance_id` | TEXT | References `instances.id` (cascade delete) |
| `credential_id` | TEXT | Credential identifier |
| `priority` | INTEGER | Priority order (0 = primary) |
| `max_concurrent` | INTEGER | Max concurrent sessions using this credential (default: 1) |
| `enabled` | INTEGER | 0/1 flag (default: 1) |

### `harness_types`

Display metadata for each harness type.

| Column | Type | Description |
|--------|------|-------------|
| `name` | TEXT PK | Harness type identifier |
| `label` | TEXT | Human-readable label |
| `emoji` | TEXT | Unicode emoji |
| `image` | TEXT | Image URL or path |

## API

### Store lifecycle

| Function | Description |
|----------|-------------|
| `Open(dbPath)` | Open or create database, apply WAL mode, run migrations |
| `Close()` | Close database connection |
| `DB()` | Access underlying `*sql.DB` |

### Instance operations

| Method | Description |
|--------|-------------|
| `CreateInstance(inst)` | Register a new instance |
| `GetInstance(id)` | Get by ID |
| `GetInstanceByName(name)` | Get by name |
| `ListInstances()` | All instances, ordered by name |
| `ListInstancesByHarness(type)` | Enabled instances for a harness type |
| `ListInstancesByHost(host)` | Instances on a specific host |
| `UpdateInstance(inst)` | Update configuration |
| `DeleteInstance(id)` | Remove instance (cascades to credentials) |
| `SetInstanceEnabled(id, enabled)` | Toggle enabled status |
| `CountInstances()` | Total count |
| `CountEnabledInstances()` | Enabled count |

### Credential binding operations

| Method | Description |
|--------|-------------|
| `BindCredential(ic)` | Associate credential with instance (upsert) |
| `UnbindCredential(instanceID, credentialID)` | Remove binding |
| `ListInstanceCredentials(instanceID)` | All bindings for an instance, ordered by priority |
| `ListCredentialInstances(credentialID)` | All instances using a credential |
| `GetCredentialBinding(instanceID, credentialID)` | Get specific binding |
| `SetCredentialEnabled(instanceID, credentialID, enabled)` | Toggle binding |
| `UpdateCredentialPriority(instanceID, credentialID, priority)` | Update priority |
| `UpdateCredentialMaxConcurrent(instanceID, credentialID, max)` | Update concurrency limit |
| `ClearInstanceCredentials(instanceID)` | Remove all bindings for an instance |
| `CountCredentialBindings(instanceID)` | Count bindings |
| `TotalMaxConcurrent(instanceID)` | Sum of max_concurrent across enabled credentials |

### Harness type operations

| Method | Description |
|--------|-------------|
| `UpsertHarnessType(ht)` | Insert or update harness type metadata |
| `GetHarnessType(name)` | Get metadata |
| `ListHarnessTypes()` | All harness types, ordered by name |
| `DeleteHarnessType(name)` | Remove entry |

## Dependencies

- [`llm-bridge`](https://github.com/kayushkin/llm-bridge) -- canonical types (`msg.Instance`, `msg.InstanceCredential`, `msg.Harness`, `msg.Transport`)
- [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) -- pure-Go SQLite driver (no CGO)

## Part of the llm-bridge ecosystem

This store is one of several optional SQLite-backed libraries that [llm-bridge-server](https://github.com/kayushkin/llm-bridge-server) can compose. See the [llm-bridge README](https://github.com/kayushkin/llm-bridge) for the full picture.
