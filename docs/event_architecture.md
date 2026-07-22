# Chorus Immutable Event Architecture & Auditability Specification

## Overview

Chorus is built around the fundamental guarantee of **immutable history**. In accordance with our storage philosophy, no post, message, or metadata update is physically deleted or destructive.

Every write operation is stored as an append-only event inside the Git repository persistence layer.

---

## Event Log Structure

For each board and thread, state projections are derived from an immutable stream of events stored in `events.ndjson`:

```text
data/repository/
└── boards/
    └── general/
        └── threads/
            └── thd_xxxxx/
                ├── thread.json          # Materialized View / State Projection
                ├── messages.ndjson       # Materialized View / Message Stream
                └── events.ndjson         # Immutable Append-Only Event Log
```

---

## Core Event Envelope Schema

All events inherit a standard JSON envelope format:

```json
{
  "event_id": "evt_8f1a92b3c4d",
  "event_type": "MessageCreated",
  "aggregate_id": "thd_xxxxx",
  "timestamp": "2026-07-22T16:20:00Z",
  "payload": {}
}
```

---

## Event Catalog

### 1. `ThreadCreated`
Emitted when a new discussion thread is initialized.
```json
{
  "event_id": "evt_01H...",
  "event_type": "ThreadCreated",
  "aggregate_id": "thd_8ade181a55615fc0",
  "timestamp": "2026-07-22T16:20:00Z",
  "payload": {
    "title": "Architecture Discussion",
    "author_id": "usr_7a18f50e29b1",
    "country": "US"
  }
}
```

### 2. `MessageCreated`
Emitted when a user appends a new reply to a thread.
```json
{
  "event_id": "evt_02H...",
  "event_type": "MessageCreated",
  "aggregate_id": "thd_8ade181a55615fc0",
  "timestamp": "2026-07-22T16:21:00Z",
  "payload": {
    "message_id": "msg_3c92e10a8b",
    "author_id": "usr_90b3a1cf44e",
    "content": "I agree with this design approach.",
    "country": "DE"
  }
}
```

### 3. `MessageReported`
Emitted when a user flags a message for moderation audit.
```json
{
  "event_id": "evt_03H...",
  "event_type": "MessageReported",
  "aggregate_id": "thd_8ade181a55615fc0",
  "timestamp": "2026-07-22T16:22:00Z",
  "payload": {
    "message_id": "msg_3c92e10a8b",
    "reason": "spam",
    "reporter_ip_hash": "a1b2c3d4"
  }
}
```

### 4. `MessageHidden`
Emitted when a message is hidden from public projections. **The message payload is never deleted from disk.**
```json
{
  "event_id": "evt_04H...",
  "event_type": "MessageHidden",
  "aggregate_id": "thd_8ade181a55615fc0",
  "timestamp": "2026-07-22T16:23:00Z",
  "payload": {
    "message_id": "msg_3c92e10a8b",
    "moderation_reason": "policy_violation"
  }
}
```

### 5. `MessageRestored`
Emitted if a hidden message is un-hidden following audit review.
```json
{
  "event_id": "evt_05H...",
  "event_type": "MessageRestored",
  "aggregate_id": "thd_8ade181a55615fc0",
  "timestamp": "2026-07-22T16:24:00Z",
  "payload": {
    "message_id": "msg_3c92e10a8b"
  }
}
```

---

## Architectural Principles & Guarantees

1. **Physical Immutability**: No event line is ever deleted or mutated in `events.ndjson`.
2. **Deterministic Projection Rebuilding**: `GitStore.RecoverAndVerify` can replay `events.ndjson` from position 0 to reconstruct `thread.json` and `messages.ndjson` at any point in time.
3. **Git Commit Alignment**: Each appended event creates a corresponding Git commit in local version history (`git commit -m "event: MessageHidden msg_3c92e10a8b"`).
