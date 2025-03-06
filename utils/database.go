package utils

import (
	"cmp"
	"errors"
	"slices"
	"sync"
)

const (
	TxPending    = 0
	TxCommitted  = 1
	TxRolledBack = 2
)

// DatabaseInterface defines a concurrent-safe key-value store with transactional support.
type DatabaseInterface[K cmp.Ordered, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	StartTransaction(keys []K) *Transaction[K, V]
}

// TransactionInterface defines an atomic unit of work on the Database.
type TransactionInterface[K cmp.Ordered, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V) error
	Commit() error
	Rollback()
}

type Database[K cmp.Ordered, V any] struct {
	dataMu sync.RWMutex
	data   map[K]V

	lockMu sync.Mutex
	locks  map[K]*sync.Mutex
}

type Transaction[K cmp.Ordered, V any] struct {
	db       *Database[K, V]
	locks    []*sync.Mutex
	temp     map[K]V
	txStatus int
}

func NewDatabase[K cmp.Ordered, V any]() *Database[K, V] {
	return &Database[K, V]{
		data:  make(map[K]V),
		locks: make(map[K]*sync.Mutex),
	}
}

// **Helper: Get per-key lock**
func (db *Database[K, V]) getLock(key K) *sync.Mutex {
	db.lockMu.Lock()
	defer db.lockMu.Unlock()
	if _, exists := db.locks[key]; !exists {
		db.locks[key] = &sync.Mutex{}
	}
	return db.locks[key]
}

// **Get: Read-only access (concurrent-safe)**
func (db *Database[K, V]) Get(key K) (V, bool) {
	db.dataMu.RLock()
	defer db.dataMu.RUnlock()
	val, exists := db.data[key]
	return val, exists
}

// **Set: Write operation (outside transactions), waits if the key is locked**
func (db *Database[K, V]) Set(key K, value V) error {
	// Acquire per-key lock to prevent conflicts with transactions
	lock := db.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	db.dataMu.Lock()
	defer db.dataMu.Unlock()
	db.data[key] = value

	return nil
}

// **StartTransaction: Locks specified keys for modification**
func (db *Database[K, V]) StartTransaction(keys []K) *Transaction[K, V] {
	tx := &Transaction[K, V]{
		db:   db,
		temp: make(map[K]V),
	}

	// Sort & remove duplicates to preven deadlock
	slices.Sort(keys)
	keys = slices.Compact(keys)

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

// **Get: Read inside a transaction (modifies temp storage)**
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
	if tx.txStatus == TxRolledBack {
		return errors.New("transaction already rolled back")
	}
	if tx.txStatus == TxCommitted {
		return errors.New("transaction already committed")
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

// **Rollback: Discard transaction changes**
func (tx *Transaction[K, V]) Rollback() {
	if tx.txStatus != TxPending {
		return
	}
	tx.txStatus = TxRolledBack

	for _, lock := range tx.locks {
		lock.Unlock()
	}
}
