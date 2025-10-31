// internal/wallet/wallet_test.go
package wallet

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"
)

// TestWalletService_CreateUser tests user creation functionality
func TestWalletService_CreateUser(t *testing.T) {
	ws := NewWalletService()

	tests := []struct {
		name     string
		userID   string
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "create new user",
			userID:   "user1",
			username: "John Doe",
			email:    "john@example.com",
			wantErr:  false,
		},
		{
			name:     "create duplicate user",
			userID:   "user1",
			username: "John Doe",
			email:    "john@example.com",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ws.CreateUser(tt.userID, tt.username, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWalletService_Deposit tests deposit functionality with various scenarios
func TestWalletService_Deposit(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")

	tests := []struct {
		name        string
		userID      string
		amount      float64
		description string
		wantErr     bool
	}{
		{
			name:        "valid deposit",
			userID:      "user1",
			amount:      100.50,
			description: "initial deposit",
			wantErr:     false,
		},
		{
			name:        "invalid negative amount",
			userID:      "user1",
			amount:      -50.0,
			description: "invalid deposit",
			wantErr:     true,
		},
		{
			name:        "invalid zero amount",
			userID:      "user1",
			amount:      0.0,
			description: "zero deposit",
			wantErr:     true,
		},
		{
			name:        "non-existent user",
			userID:      "nonexistent",
			amount:      100.0,
			description: "test deposit",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ws.Deposit(tt.userID, tt.amount, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deposit() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify balance for successful deposits
			if !tt.wantErr && err == nil {
				balance, err := ws.GetBalance(tt.userID)
				if err != nil {
					t.Errorf("GetBalance() error = %v", err)
				}
				if balance != tt.amount {
					t.Errorf("Expected balance %.2f, got %.2f", tt.amount, balance)
				}
			}
		})
	}
}

// TestWalletService_Withdraw tests withdrawal functionality including edge cases
func TestWalletService_Withdraw(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		amount      float64
		description string
		wantErr     bool
	}{
		{
			name:        "valid withdrawal",
			userID:      "user1",
			amount:      50.0,
			description: "withdrawal",
			wantErr:     false,
		},
		{
			name:        "insufficient balance",
			userID:      "user1",
			amount:      300.0,
			description: "overdraw",
			wantErr:     true,
		},
		{
			name:        "invalid negative amount",
			userID:      "user1",
			amount:      -10.0,
			description: "invalid withdrawal",
			wantErr:     true,
		},
		{
			name:        "non-existent user",
			userID:      "nonexistent",
			amount:      50.0,
			description: "test withdrawal",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh wallet service for each test
			ws := NewWalletService()
			ws.CreateUser("user1", "John Doe", "john@example.com")

			// Setup: deposit some money for withdrawal tests
			if tt.name == "valid withdrawal" || tt.name == "insufficient balance" {
				ws.Deposit("user1", 200.0, "setup deposit")
			}

			err := ws.Withdraw(tt.userID, tt.amount, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("Withdraw() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestWalletService_Transfer tests transfer functionality between users
func TestWalletService_Transfer(t *testing.T) {
	tests := []struct {
		name        string
		fromUserID  string
		toUserID    string
		amount      float64
		description string
		wantErr     bool
	}{
		{
			name:        "valid transfer",
			fromUserID:  "user1",
			toUserID:    "user2",
			amount:      100.0,
			description: "transfer to jane",
			wantErr:     false,
		},
		{
			name:        "insufficient balance",
			fromUserID:  "user1",
			toUserID:    "user2",
			amount:      500.0,
			description: "large transfer",
			wantErr:     true,
		},
		{
			name:        "transfer to same user",
			fromUserID:  "user1",
			toUserID:    "user1",
			amount:      50.0,
			description: "self transfer",
			wantErr:     true,
		},
		{
			name:        "invalid amount",
			fromUserID:  "user1",
			toUserID:    "user2",
			amount:      -50.0,
			description: "invalid transfer",
			wantErr:     true,
		},
		{
			name:        "non-existent from user",
			fromUserID:  "nonexistent",
			toUserID:    "user2",
			amount:      50.0,
			description: "test transfer",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh wallet service for each test
			ws := NewWalletService()
			ws.CreateUser("user1", "John Doe", "john@example.com")
			ws.CreateUser("user2", "Jane Smith", "jane@example.com")

			// Setup: deposit money for transfer tests
			if tt.name == "valid transfer" || tt.name == "insufficient balance" {
				ws.Deposit("user1", 300.0, "setup deposit")
			}

			// Get initial balances for verification
			initialFromBalance, _ := ws.GetBalance(tt.fromUserID)
			initialToBalance, _ := ws.GetBalance(tt.toUserID)

			err := ws.Transfer(tt.fromUserID, tt.toUserID, tt.amount, tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify balance changes for successful transfers
			if !tt.wantErr && err == nil {
				finalFromBalance, _ := ws.GetBalance(tt.fromUserID)
				finalToBalance, _ := ws.GetBalance(tt.toUserID)

				if finalFromBalance != initialFromBalance-tt.amount {
					t.Errorf("From user balance incorrect: expected %.2f, got %.2f",
						initialFromBalance-tt.amount, finalFromBalance)
				}
				if finalToBalance != initialToBalance+tt.amount {
					t.Errorf("To user balance incorrect: expected %.2f, got %.2f",
						initialToBalance+tt.amount, finalToBalance)
				}
			}
		})
	}
}

// TestWalletService_ConcurrentOperations tests thread safety under concurrent load
func TestWalletService_ConcurrentOperations(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")
	ws.CreateUser("user2", "Jane Smith", "jane@example.com")
	ws.Deposit("user1", 1000.0, "initial deposit")

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent deposits to user1
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			ws.Deposit("user1", 10.0, "concurrent deposit")
		}()
	}

	// Concurrent transfers from user1 to user2
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			ws.Transfer("user1", "user2", 1.0, "concurrent transfer")
		}()
	}

	wg.Wait()

	// Verify final balances
	balance1, _ := ws.GetBalance("user1")
	balance2, _ := ws.GetBalance("user2")

	expectedBalance1 := 1000.0 + float64(iterations)*10.0 - float64(iterations)*1.0
	expectedBalance2 := float64(iterations) * 1.0

	if balance1 != expectedBalance1 {
		t.Errorf("Expected user1 balance %.2f, got %.2f", expectedBalance1, balance1)
	}
	if balance2 != expectedBalance2 {
		t.Errorf("Expected user2 balance %.2f, got %.2f", expectedBalance2, balance2)
	}
}

// TestWalletService_TransactionHistory tests transaction history retrieval
func TestWalletService_TransactionHistory(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")
	ws.CreateUser("user2", "Jane Smith", "jane@example.com")

	// Perform multiple transactions
	ws.Deposit("user1", 100.0, "deposit 1")
	ws.Deposit("user1", 200.0, "deposit 2")
	ws.Transfer("user1", "user2", 50.0, "transfer to jane")

	// Test user1 transaction history
	user1Transactions, err := ws.GetTransactionHistory("user1")
	if err != nil {
		t.Errorf("GetTransactionHistory() error = %v", err)
	}
	if len(user1Transactions) != 3 {
		t.Errorf("Expected 3 transactions for user1, got %d", len(user1Transactions))
	}

	// Test user2 transaction history
	user2Transactions, err := ws.GetTransactionHistory("user2")
	if err != nil {
		t.Errorf("GetTransactionHistory() error = %v", err)
	}
	if len(user2Transactions) != 1 {
		t.Errorf("Expected 1 transaction for user2, got %d", len(user2Transactions))
	}

	// Test non-existent user
	_, err = ws.GetTransactionHistory("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected user not found error, got %v", err)
	}
}

// TestWalletService_PrecisionHandling tests decimal precision in calculations
func TestWalletService_PrecisionHandling(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")

	// Test small decimal amounts using decimal interface
	err := ws.DepositDecimal("user1", decimal.NewFromFloat(0.1), "small deposit 1")
	if err != nil {
		t.Errorf("Deposit() error = %v", err)
	}

	err = ws.DepositDecimal("user1", decimal.NewFromFloat(0.2), "small deposit 2")
	if err != nil {
		t.Errorf("Deposit() error = %v", err)
	}

	balance, err := ws.GetBalanceDecimal("user1")
	if err != nil {
		t.Errorf("GetBalance() error = %v", err)
	}

	expected := decimal.NewFromFloat(0.3)
	if !balance.Equal(expected) {
		t.Errorf("Expected balance %s, got %s", expected.String(), balance.String())
	}
}

// TestWalletService_DecimalPrecision tests more decimal precision scenarios
func TestWalletService_DecimalPrecision(t *testing.T) {
	tests := []struct {
		name     string
		deposits []decimal.Decimal
		expected decimal.Decimal
	}{
		{
			name: "small decimals",
			deposits: []decimal.Decimal{
				decimal.NewFromFloat(0.1),
				decimal.NewFromFloat(0.2),
			},
			expected: decimal.NewFromFloat(0.3),
		},
		{
			name: "more decimals",
			deposits: []decimal.Decimal{
				decimal.NewFromFloat(0.01),
				decimal.NewFromFloat(0.02),
				decimal.NewFromFloat(0.03),
			},
			expected: decimal.NewFromFloat(0.06),
		},
		{
			name: "exact decimal amounts",
			deposits: []decimal.Decimal{
				decimal.NewFromFloat(0.333),
				decimal.NewFromFloat(0.667),
			},
			expected: decimal.NewFromFloat(1.0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh wallet service for each test to avoid state pollution
			ws := NewWalletService()
			ws.CreateUser("user1", "John Doe", "john@example.com")

			// Perform deposits
			for i, amount := range tt.deposits {
				err := ws.DepositDecimal("user1", amount, "deposit")
				if err != nil {
					t.Errorf("Deposit %d failed: %v", i, err)
				}
			}

			balance, err := ws.GetBalanceDecimal("user1")
			if err != nil {
				t.Errorf("GetBalance() error = %v", err)
			}

			if !balance.Equal(tt.expected) {
				t.Errorf("Expected balance %s, got %s", tt.expected.String(), balance.String())
			}
		})
	}
}

// TestWalletService_TransferPrecision tests precision in transfer operations
func TestWalletService_TransferPrecision(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")
	ws.CreateUser("user2", "Jane Smith", "jane@example.com")

	// Deposit with decimal amounts
	err := ws.DepositDecimal("user1", decimal.NewFromFloat(100.75), "initial deposit")
	if err != nil {
		t.Errorf("Deposit() error = %v", err)
	}

	// Transfer decimal amount
	err = ws.Transfer("user1", "user2", 50.25, "decimal transfer")
	if err != nil {
		t.Errorf("Transfer() error = %v", err)
	}

	balance1, err := ws.GetBalanceDecimal("user1")
	if err != nil {
		t.Errorf("GetBalance() error = %v", err)
	}

	balance2, err := ws.GetBalanceDecimal("user2")
	if err != nil {
		t.Errorf("GetBalance() error = %v", err)
	}

	// Check balances
	expected1 := decimal.NewFromFloat(50.5)
	expected2 := decimal.NewFromFloat(50.25)

	if !balance1.Equal(expected1) {
		t.Errorf("Expected user1 balance %s, got %s", expected1.String(), balance1.String())
	}
	if !balance2.Equal(expected2) {
		t.Errorf("Expected user2 balance %s, got %s", expected2.String(), balance2.String())
	}
}

// TestWalletService_EdgeCases tests various edge cases
func TestWalletService_EdgeCases(t *testing.T) {
	ws := NewWalletService()

	// Test operations on non-existent user
	_, err := ws.GetBalance("nonexistent")
	if err != ErrUserNotFound {
		t.Error("Expected user not found error")
	}

	err = ws.Withdraw("nonexistent", 100.0, "Test")
	if err != ErrUserNotFound {
		t.Error("Expected user not found error")
	}

	err = ws.Transfer("nonexistent", "other", 100.0, "Test")
	if err != ErrUserNotFound {
		t.Error("Expected user not found error")
	}
}

// TestWalletService_DecimalEdgeCases tests decimal-specific edge cases
func TestWalletService_DecimalEdgeCases(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")

	// Test very small amounts
	err := ws.DepositDecimal("user1", decimal.NewFromFloat(0.0001), "tiny deposit")
	if err != nil {
		t.Errorf("Deposit failed for tiny amount: %v", err)
	}

	// Test very large amounts
	err = ws.DepositDecimal("user1", decimal.NewFromFloat(1000000.99), "large deposit")
	if err != nil {
		t.Errorf("Deposit failed for large amount: %v", err)
	}

	balance, err := ws.GetBalanceDecimal("user1")
	if err != nil {
		t.Errorf("GetBalance() error = %v", err)
	}

	// Fixed: 0.0001 + 1000000.99 = 1000000.9901
	expected := decimal.NewFromFloat(1000000.9901)
	if !balance.Equal(expected) {
		t.Errorf("Expected balance %s, got %s", expected.String(), balance.String())
	}
}

// TestWalletService_WithdrawDecimal tests withdrawal with decimal amounts
func TestWalletService_WithdrawDecimal(t *testing.T) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")

	// Setup: deposit some money
	ws.DepositDecimal("user1", decimal.NewFromFloat(200.0), "initial deposit")

	// Test valid withdrawal with decimal
	err := ws.Withdraw("user1", 50.25, "decimal withdrawal")
	if err != nil {
		t.Errorf("Withdraw failed: %v", err)
	}

	balance, err := ws.GetBalanceDecimal("user1")
	if err != nil {
		t.Errorf("GetBalance() error = %v", err)
	}

	expected := decimal.NewFromFloat(149.75)
	if !balance.Equal(expected) {
		t.Errorf("Expected balance %s, got %s", expected.String(), balance.String())
	}
}

// BenchmarkWalletService_ConcurrentTransfers benchmarks transfer performance
func BenchmarkWalletService_ConcurrentTransfers(b *testing.B) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")
	ws.CreateUser("user2", "Jane Smith", "jane@example.com")
	ws.Deposit("user1", float64(b.N)*2, "initial deposit")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ws.Transfer("user1", "user2", 1.0, "benchmark transfer")
		}
	})
}

// BenchmarkWalletService_DecimalOperations benchmarks decimal operation performance
func BenchmarkWalletService_DecimalOperations(b *testing.B) {
	ws := NewWalletService()
	ws.CreateUser("user1", "John Doe", "john@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.DepositDecimal("user1", decimal.NewFromFloat(1.0), "benchmark deposit")
	}
}
