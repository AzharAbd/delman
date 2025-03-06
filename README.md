# Delman

## Overview
Delman is a simple key-value database with transactional support, built in Go. It ensures thread-safe read/write operations and provides a locking mechanism to manage concurrent access.

## Dependencies
This project uses the following dependency:
- [gorilla/mux](https://github.com/gorilla/mux) (`v1.8.1`): A powerful HTTP router and dispatcher for building REST APIs in Go.

## Installation
### Prerequisites
- Go 1.22 or later installed on your system.

### Steps to Install
1. Clone the repository:
   ```sh
   git clone https://github.com/azharabd/delman.git
   cd delman
   ```
2. Initialize Go modules (if not already initialized):
   ```sh
   go mod tidy
   ```
3. Build the project:
   ```sh
   go build -o delman
   ```
4. Run the project:
   ```sh
   ./delman
   ```

## Database Locking Mechanism
The `database.go` file implements a thread-safe key-value store with a transaction system. The key aspects of the locking mechanism include:

### 1. **Per-Key Locks**
- A `sync.Mutex` is created for each key dynamically.
- This prevents multiple transactions from modifying the same key simultaneously.

```go
func (db *Database[K, V]) getLock(key K) *sync.Mutex {
    db.lockMu.RLock()

    lock, exists := db.locks[key]
    if !exists {
        db.lockMu.RUnlock()
        return db.initLock(key)
    }

    defer db.lockMu.RUnlock()
    return lock
}
```

### 2. **Concurrent Reads (`Get` Method)**
- Uses a `sync.RWMutex` for concurrent read access.
- Ensures multiple reads can happen in parallel without blocking each other.

```go
func (db *Database[K, V]) Get(key K) (V, bool) {
    db.mu.RLock()
    defer db.mu.RUnlock()
    val, exists := db.data[key]
    return val, exists
}
```

### 3. **Safe Writes (`Set` Method)**
- Uses a per-key lock to ensure only one goroutine modifies a key at a time.
- Prevents conflicts between transactions and direct writes.

```go
func (db *Database[K, V]) Set(key K, value V) error {
    lock := db.getLock(key)
    lock.Lock()
    defer lock.Unlock()

    db.mu.Lock()
    defer db.mu.Unlock()
    db.data[key] = value
    return nil
}
```

### 4. **Transaction Locking (`StartTransaction` Method)**
- Acquires locks for all specified keys before processing a transaction.
- Locks are acquired in a sorted order without duplicates to prevent deadlocks.

```go
func (db *Database[K, V]) StartTransaction(keys []K) *Transaction[K, V] {
    tx := &Transaction[K, V]{
        db:   db,
        temp: make(map[K]V),
    }
	slices.Sort(keys)
    keys = slices.Compact(keys)
    for _, key := range keys {
        lock := db.getLock(key)
        lock.Lock()
        tx.locks = append(tx.locks, lock)
        if val, exist := db.Get(key); exist {
            tx.temp[key] = val
        }
    }
    return tx
}
```

### 5. **Commit & Rollback Mechanism**
- Commits apply changes atomically while holding `db.mu.Lock()`.
- Rollback releases all locks without modifying the database.

```go
func (tx *Transaction[K, V]) Commit() error {
    switch tx.txStatus {
    case TxCommitted:
        return errors.New("transaction already commited")
    case TxRolledBack:
        return errors.New("transaction already rolled back")
    }
    tx.txStatus = TxCommitted

    tx.db.dataMu.Lock()
    defer tx.db.dataMu.Unlock()

    for key, value := range tx.temp {
        tx.db.data[key] = value
    }

    for _, lock := range tx.locks {
        lock.Unlock()
    }

    return nil
}
```

## Contributing
Feel free to submit issues or pull requests if you'd like to contribute! 

## License
This project is licensed under the MIT License.

