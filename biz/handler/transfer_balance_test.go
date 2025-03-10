package handler

import (
	"bytes"
	"delman/biz/model"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestHandler_TransferBalance(t *testing.T) {
	db := model.InitDBUserBalance() // Initializes Mark: 100, Jane: 30, Adam: 0
	handler := InitUserBalanceHandler(db)

	var wg sync.WaitGroup
	markToJaneTransfers := 100 // Mark has enough to send 100 times
	janeToAdamTransfers := 30  // Jane only has $30 to send

	// Define transfer function to run in goroutines
	transfer := func(sender, receiver string, amount int64) {
		defer wg.Done()
		reqBody, _ := json.Marshal(TransferBalanceRequest{
			Sender: sender, Receiver: receiver, Amount: amount,
		})

		req, _ := http.NewRequest("POST", "/balance/transfer", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.TransferBalance(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Unexpected error: %v", resp.Status)
		}
	}

	// Run 100 concurrent transfers from Mark to Jane (Mark starts with $100)
	for i := 0; i < markToJaneTransfers; i++ {
		wg.Add(1)
		go transfer("Mark", "Jane", 1)
	}

	// Run 30 concurrent transfers from Jane to Adam (Jane starts with $30)
	for i := 0; i < janeToAdamTransfers; i++ {
		wg.Add(1)
		go transfer("Jane", "Adam", 1)
	}

	wg.Wait() // Wait for all goroutines to finish

	// Verify final balances are correct
	markBalance, _ := handler.db.Get("mark")
	janeBalance, _ := handler.db.Get("jane")
	adamBalance, _ := handler.db.Get("adam")

	fmt.Println(markBalance, janeBalance, adamBalance)
	// Check if any balance is negative (which shouldn't happen)
	if markBalance.Balance < 0 || janeBalance.Balance < 0 || adamBalance.Balance < 0 {
		t.Errorf("Negative balance detected! Mark: %d, Jane: %d, Adam: %d",
			markBalance.Balance, janeBalance.Balance, adamBalance.Balance)
	}

	// Check total balance consistency (should always be 130)
	totalBalance := markBalance.Balance + janeBalance.Balance + adamBalance.Balance
	if totalBalance != 130 {
		t.Errorf("Inconsistent total balance! Expected 130, got %d", totalBalance)
	}
}

func TestHandler_ConcurrentStressTest(t *testing.T) {
	db := model.InitDBUserBalance() // Initializes Mark: 100, Jane: 30, Adam: 0
	handler := InitUserBalanceHandler(db)

	var wg sync.WaitGroup
	totalTransfers := 1000 // 🚀 Let's make it extreme!
	users := []string{"Mark", "Jane", "Adam"}

	// Helper function for concurrent transfers
	transfer := func(sender, receiver string, amount int64) {
		defer wg.Done()
		reqBody, _ := json.Marshal(TransferBalanceRequest{
			Sender: sender, Receiver: receiver, Amount: amount,
		})

		req, _ := http.NewRequest("POST", "/balance/transfer", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.TransferBalance(rec, req)

		resp := rec.Result()
		defer resp.Body.Close()
	}

	// 🚀 Scenario 1: Many-to-Many Transfers (1000 Random Transactions)
	for i := 0; i < totalTransfers; i++ {
		wg.Add(1)
		go func() {
			sender := users[rand.Intn(len(users))]
			receiver := users[rand.Intn(len(users))]
			amount := int64(rand.Intn(20) + 1) // Random amount between 1 and 20

			// Ensure sender and receiver are not the same
			if sender != receiver {
				transfer(sender, receiver, amount)
			} else {
				wg.Done()
			}
		}()
	}

	// 🚀 Scenario 2: Rapid-Fire Transfers (Every 1ms)
	stopCh := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopCh:
				return
			default:
				wg.Add(1)
				go transfer("Mark", "Jane", 1)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Let it run for a while
	time.Sleep(5 * time.Second)
	stopCh <- true

	wg.Wait() // Wait for all transactions to complete

	// 📌 Scenario 3: Final Checks
	markBalance, _ := handler.db.Get("mark")
	janeBalance, _ := handler.db.Get("jane")
	adamBalance, _ := handler.db.Get("adam")

	// ✅ Ensure no negative balances
	if markBalance.Balance < 0 || janeBalance.Balance < 0 || adamBalance.Balance < 0 {
		t.Errorf("Negative balance detected! Mark: %d, Jane: %d, Adam: %d",
			markBalance.Balance, janeBalance.Balance, adamBalance.Balance)
	}

	// ✅ Ensure total balance consistency (should always be 130)
	totalBalance := markBalance.Balance + janeBalance.Balance + adamBalance.Balance
	if totalBalance != 130 {
		t.Errorf("Inconsistent total balance! Expected 130, got %d", totalBalance)
	}
}
