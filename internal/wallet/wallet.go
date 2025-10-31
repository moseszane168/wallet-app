// internal/wallet/wallet.go
package wallet

import (
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// WalletService manages all wallet operations and user accounts
type WalletService struct {
	users        map[string]*User
	wallets      map[string]*Wallet
	transactions []*Transaction
	mu           sync.RWMutex
	userLocks    *userLockManager
}

// userLockManager manages locks for individual users to prevent deadlocks
type userLockManager struct {
	locks sync.Map
}

// getLock returns a mutex for the given user ID
func (ulm *userLockManager) getLock(userID string) *sync.Mutex {
	lock, _ := ulm.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// NewWalletService creates and initializes a new WalletService instance
func NewWalletService() *WalletService {
	return &WalletService{
		users:        make(map[string]*User),
		wallets:      make(map[string]*Wallet),
		transactions: make([]*Transaction, 0),
		userLocks:    &userLockManager{},
	}
}

// CreateUser creates a new user and initializes an empty wallet for them
func (ws *WalletService) CreateUser(userID, name, email string) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if _, exists := ws.users[userID]; exists {
		return ErrUserAlreadyExists
	}

	user := &User{
		ID:    userID,
		Name:  name,
		Email: email,
	}

	wallet := &Wallet{
		UserID:  userID,
		Balance: decimal.NewFromFloat(0.0),
	}

	ws.users[userID] = user
	ws.wallets[userID] = wallet

	return nil
}

// Deposit adds funds to a user's wallet
func (ws *WalletService) Deposit(userID string, amount float64, description string) error {
	decimalAmount := decimal.NewFromFloat(amount)
	if decimalAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	// Get user-specific lock to prevent concurrent operations
	userLock := ws.userLocks.getLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	ws.mu.RLock()
	wallet, exists := ws.wallets[userID]
	ws.mu.RUnlock()

	if !exists {
		return ErrUserNotFound
	}

	wallet.mu.Lock()
	wallet.Balance = wallet.Balance.Add(decimalAmount)
	wallet.mu.Unlock()

	// Record the transaction
	tx := &Transaction{
		ID:          generateTransactionID(),
		FromUserID:  userID,
		ToUserID:    userID,
		Amount:      decimalAmount,
		Type:        TransactionDeposit,
		Description: description,
		Timestamp:   time.Now().Unix(),
	}

	ws.recordTransaction(tx)

	return nil
}

// DepositDecimal adds funds to a user's wallet using decimal.Decimal
func (ws *WalletService) DepositDecimal(userID string, amount decimal.Decimal, description string) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	// Get user-specific lock to prevent concurrent operations
	userLock := ws.userLocks.getLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	ws.mu.RLock()
	wallet, exists := ws.wallets[userID]
	ws.mu.RUnlock()

	if !exists {
		return ErrUserNotFound
	}

	wallet.mu.Lock()
	wallet.Balance = wallet.Balance.Add(amount)
	wallet.mu.Unlock()

	// Record the transaction
	tx := &Transaction{
		ID:          generateTransactionID(),
		FromUserID:  userID,
		ToUserID:    userID,
		Amount:      amount,
		Type:        TransactionDeposit,
		Description: description,
		Timestamp:   time.Now().Unix(),
	}

	ws.recordTransaction(tx)

	return nil
}

// Withdraw removes funds from a user's wallet
func (ws *WalletService) Withdraw(userID string, amount float64, description string) error {
	decimalAmount := decimal.NewFromFloat(amount)
	if decimalAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	// Get user-specific lock
	userLock := ws.userLocks.getLock(userID)
	userLock.Lock()
	defer userLock.Unlock()

	ws.mu.RLock()
	wallet, exists := ws.wallets[userID]
	ws.mu.RUnlock()

	if !exists {
		return ErrUserNotFound
	}

	wallet.mu.Lock()
	defer wallet.mu.Unlock()

	if wallet.Balance.LessThan(decimalAmount) {
		return ErrInsufficientBalance
	}

	wallet.Balance = wallet.Balance.Sub(decimalAmount)

	// Record the transaction
	tx := &Transaction{
		ID:          generateTransactionID(),
		FromUserID:  userID,
		ToUserID:    userID,
		Amount:      decimalAmount,
		Type:        TransactionWithdraw,
		Description: description,
		Timestamp:   time.Now().Unix(),
	}

	ws.recordTransaction(tx)

	return nil
}

// Transfer moves funds from one user to another
func (ws *WalletService) Transfer(fromUserID, toUserID string, amount float64, description string) error {
	decimalAmount := decimal.NewFromFloat(amount)
	if decimalAmount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	if fromUserID == toUserID {
		return ErrSameUserTransfer
	}

	// Verify both users exist
	ws.mu.RLock()
	fromWallet, fromExists := ws.wallets[fromUserID]
	toWallet, toExists := ws.wallets[toUserID]
	ws.mu.RUnlock()

	if !fromExists || !toExists {
		return ErrUserNotFound
	}

	// To prevent deadlocks, always acquire locks in consistent order
	firstLock, secondLock := ws.getOrderedLocks(fromUserID, toUserID)

	firstLock.Lock()
	secondLock.Lock()
	defer firstLock.Unlock()
	defer secondLock.Unlock()

	// Check sufficient balance
	fromWallet.mu.Lock()
	if fromWallet.Balance.LessThan(decimalAmount) {
		fromWallet.mu.Unlock()
		return ErrInsufficientBalance
	}
	fromWallet.Balance = fromWallet.Balance.Sub(decimalAmount)
	fromWallet.mu.Unlock()

	// Update recipient balance
	toWallet.mu.Lock()
	toWallet.Balance = toWallet.Balance.Add(decimalAmount)
	toWallet.mu.Unlock()

	// Record the transaction
	tx := &Transaction{
		ID:          generateTransactionID(),
		FromUserID:  fromUserID,
		ToUserID:    toUserID,
		Amount:      decimalAmount,
		Type:        TransactionTransfer,
		Description: description,
		Timestamp:   time.Now().Unix(),
	}

	ws.recordTransaction(tx)

	return nil
}

// GetBalance returns the current balance of a user's wallet as float64
func (ws *WalletService) GetBalance(userID string) (float64, error) {
	balance, err := ws.GetBalanceDecimal(userID)
	if err != nil {
		return 0, err
	}
	balanceFloat, _ := balance.Float64()
	return balanceFloat, nil
}

// GetBalanceDecimal returns the current balance of a user's wallet as decimal.Decimal
func (ws *WalletService) GetBalanceDecimal(userID string) (decimal.Decimal, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	wallet, exists := ws.wallets[userID]
	if !exists {
		return decimal.Zero, ErrUserNotFound
	}

	wallet.mu.RLock()
	defer wallet.mu.RUnlock()

	return wallet.Balance, nil
}

// GetTransactionHistory returns all transactions for a specific user
func (ws *WalletService) GetTransactionHistory(userID string) ([]*Transaction, error) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if _, exists := ws.users[userID]; !exists {
		return nil, ErrUserNotFound
	}

	var userTransactions []*Transaction
	for _, tx := range ws.transactions {
		if tx.FromUserID == userID || tx.ToUserID == userID {
			userTransactions = append(userTransactions, tx)
		}
	}

	return userTransactions, nil
}

// GetAllUsers returns a list of all users in the system
func (ws *WalletService) GetAllUsers() []*User {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	users := make([]*User, 0, len(ws.users))
	for _, user := range ws.users {
		users = append(users, user)
	}

	return users
}

// getOrderedLocks returns locks for two users in consistent order to prevent deadlocks
func (ws *WalletService) getOrderedLocks(userID1, userID2 string) (*sync.Mutex, *sync.Mutex) {
	lock1 := ws.userLocks.getLock(userID1)
	lock2 := ws.userLocks.getLock(userID2)

	// Always acquire locks in alphabetical order of user IDs to prevent deadlocks
	if userID1 < userID2 {
		return lock1, lock2
	}
	return lock2, lock1
}

// recordTransaction safely adds a transaction to the history
func (ws *WalletService) recordTransaction(tx *Transaction) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.transactions = append(ws.transactions, tx)
}

// generateTransactionID creates a unique transaction ID
func generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}
