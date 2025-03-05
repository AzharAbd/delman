package utils

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// DatabaseInterface defines a concurrent-safe key-value store with transactional support.
type DatabaseInterface[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	StartTransaction(keys []K) *Transaction[K, V]
}

// TransactionInterface defines an atomic unit of work on the Database.
type TransactionInterface[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	Commit() error
	Rollback()
}

type Database[K comparable, V any] struct {
	mu    sync.RWMutex
	data  map[K]V
	locks map[K]*sync.Mutex
}

type Transaction[K comparable, V any] struct {
	db         *Database[K, V]
	locks      []*sync.Mutex
	temp       map[K]V
	rolledBack bool
}

// **Create a new Database**
func NewDatabase[K comparable, V any]() *Database[K, V] {
	return &Database[K, V]{
		data:  make(map[K]V),
		locks: make(map[K]*sync.Mutex),
	}
}

// **Helper: Get per-key lock**
func (db *Database[K, V]) getLock(key K) *sync.Mutex {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, exists := db.locks[key]; !exists {
		db.locks[key] = &sync.Mutex{}
	}
	return db.locks[key]
}

// **Get: Read-only access (concurrent-safe)**
func (db *Database[K, V]) Get(key K) (V, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, exists := db.data[key]
	return val, exists
}

// **Set: Write operation (outside transactions), waits if the key is locked**
func (db *Database[K, V]) Set(key K, value V) error {
	// Acquire per-key lock to prevent conflicts with transactions
	lock := db.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[key] = value

	return nil
}

// **StartTransaction: Locks specified keys for modification**
func (db *Database[K, V]) StartTransaction(keys []K) *Transaction[K, V] {
	tx := &Transaction[K, V]{
		db:   db,
		temp: make(map[K]V),
	}

	//Sort keys before acquiring locks to prevent deadlock
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
	})

	for _, key := range keys {
		lock := db.getLock(key)
		lock.Lock()
		tx.locks = append(tx.locks, lock)

		// Copy original values into transaction temp storage
		if val, exist := db.Get(key); exist {
			tx.temp[key] = val
		}
	}

	return tx
}

// **GetForUpdate: Read inside a transaction (modifies temp storage)**
func (tx *Transaction[K, V]) Get(key K) (V, bool) {
	val, exists := tx.temp[key]
	return val, exists
}

// **Set inside transaction**
func (tx *Transaction[K, V]) Set(key K, value V) error {
	tx.temp[key] = value
	return nil
}

// **Commit: Apply all changes atomically**
func (tx *Transaction[K, V]) Commit() error {
	if tx.rolledBack {
		return errors.New("transaction already rolled back")
	}

	tx.db.mu.Lock()
	defer tx.db.mu.Unlock()

	// Apply transaction changes to the main database
	for key, value := range tx.temp {
		tx.db.data[key] = value
	}

	// Unlock all keys
	for _, lock := range tx.locks {
		lock.Unlock()
	}
	tx.locks = nil

	return nil
}

// **Rollback: Discard transaction changes**
func (tx *Transaction[K, V]) Rollback() {
	if tx.rolledBack {
		return
	}
	tx.rolledBack = true

	// Unlock all keys
	for _, lock := range tx.locks {
		lock.Unlock()
	}
	tx.locks = nil
}
