// internal/wallet/types.go
package wallet

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
)

// Error definitions for wallet operations
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrSameUserTransfer    = errors.New("cannot transfer to same user")
	ErrUserAlreadyExists   = errors.New("user already exists")
)

// User represents a wallet user with basic information
type User struct {
	ID    string
	Name  string
	Email string
}

// Wallet represents a user's wallet with balance and locking mechanism
type Wallet struct {
	UserID  string
	Balance decimal.Decimal
	mu      sync.RWMutex
}

// TransactionType defines the type of transaction
type TransactionType string

const (
	TransactionDeposit  TransactionType = "deposit"
	TransactionWithdraw TransactionType = "withdraw"
	TransactionTransfer TransactionType = "transfer"
)

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID          string
	FromUserID  string
	ToUserID    string
	Amount      decimal.Decimal
	Type        TransactionType
	Description string
	Timestamp   int64
}
