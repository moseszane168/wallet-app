// examples/main.go
package main

import (
	"fmt"
	"log"
	"sort"

	"github.com/shopspring/decimal"
	"wallet-app/internal/wallet"
)

// main demonstrates the usage of the wallet service with decimal precision
func main() {
	// Initialize wallet service
	ws := wallet.NewWalletService()

	// Create users
	users := []struct {
		id    string
		name  string
		email string
	}{
		{"user1", "Alice Johnson", "alice@example.com"},
		{"user2", "Bob Smith", "bob@example.com"},
		{"user3", "Charlie Brown", "charlie@example.com"},
	}

	fmt.Println("=== Wallet Application Demo ===")
	fmt.Println("Creating users...")

	for _, user := range users {
		err := ws.CreateUser(user.id, user.name, user.email)
		if err != nil {
			log.Printf("Failed to create user %s: %v", user.id, err)
		} else {
			fmt.Printf("✓ Created user: %s (%s)\n", user.name, user.id)
		}
	}

	// Demonstrate decimal precision with float64 interface
	fmt.Println("\n--- Deposit Operations (Float64 Interface) ---")
	err := ws.Deposit("user1", 1000.0, "Initial deposit")
	if err != nil {
		log.Printf("Deposit failed: %v", err)
	} else {
		fmt.Println("✓ Deposited $1000.00 to Alice's wallet")
	}

	err = ws.Deposit("user2", 500.0, "Initial deposit")
	if err != nil {
		log.Printf("Deposit failed: %v", err)
	} else {
		fmt.Println("✓ Deposited $500.00 to Bob's wallet")
	}

	// Demonstrate precise decimal operations
	fmt.Println("\n--- Precise Decimal Operations ---")

	// Using decimal.Decimal for exact amounts
	smallAmount1 := decimal.NewFromFloat(0.1)
	smallAmount2 := decimal.NewFromFloat(0.2)

	err = ws.DepositDecimal("user1", smallAmount1, "Small precise deposit 1")
	if err != nil {
		log.Printf("Precise deposit failed: %v", err)
	} else {
		fmt.Printf("✓ Deposited %s to Alice's wallet (precise decimal)\n", formatDecimal(smallAmount1))
	}

	err = ws.DepositDecimal("user1", smallAmount2, "Small precise deposit 2")
	if err != nil {
		log.Printf("Precise deposit failed: %v", err)
	} else {
		fmt.Printf("✓ Deposited %s to Alice's wallet (precise decimal)\n", formatDecimal(smallAmount2))
	}

	// Show the exact decimal balance
	preciseBalance, err := ws.GetBalanceDecimal("user1")
	if err != nil {
		log.Printf("Failed to get precise balance: %v", err)
	} else {
		fmt.Printf("✓ Alice's precise balance after 0.1 + 0.2: %s\n", formatDecimal(preciseBalance))
		fmt.Printf("  Note: Float64 would show: 0.1 + 0.2 = 0.30000000000000004\n")
		fmt.Printf("  Decimal correctly shows: 1000 + 0.1 + 0.2 = %s\n", formatDecimal(preciseBalance))
	}

	// Perform transfer operations with decimal precision
	fmt.Println("\n--- Transfer Operations with Decimal Precision ---")

	// Transfer using float64 interface
	err = ws.Transfer("user1", "user2", 250.50, "Lunch payment")
	if err != nil {
		log.Printf("Transfer failed: %v", err)
	} else {
		fmt.Println("✓ Transferred $250.50 from Alice to Bob")
	}

	// Transfer using exact decimal amount
	exactTransferAmount := decimal.NewFromFloat(123.45)
	err = ws.DepositDecimal("user3", exactTransferAmount, "Exact decimal deposit")
	if err != nil {
		log.Printf("Deposit failed: %v", err)
	} else {
		fmt.Printf("✓ Deposited exact amount %s to Charlie's wallet\n", formatDecimal(exactTransferAmount))
	}

	// Perform withdrawal operation with mixed interfaces
	fmt.Println("\n--- Withdrawal Operations ---")
	err = ws.Withdraw("user1", 100.0, "Cash withdrawal")
	if err != nil {
		log.Printf("Withdrawal failed: %v", err)
	} else {
		fmt.Println("✓ Withdrew $100.00 from Alice's wallet")
	}

	// Demonstrate complex decimal operations
	fmt.Println("\n--- Complex Decimal Operations ---")

	// Add multiple small decimal amounts
	smallAmounts := []decimal.Decimal{
		decimal.NewFromFloat(0.001),
		decimal.NewFromFloat(0.002),
		decimal.NewFromFloat(0.003),
		decimal.NewFromFloat(0.004),
	}

	for i, amount := range smallAmounts {
		err = ws.DepositDecimal("user3", amount, fmt.Sprintf("Small deposit %d", i+1))
		if err != nil {
			log.Printf("Small deposit failed: %v", err)
		} else {
			fmt.Printf("✓ Added %s to Charlie's wallet\n", formatDecimal(amount))
		}
	}

	// Display final balances with both float64 and decimal precision
	fmt.Println("\n--- Final Balances ---")
	fmt.Println("User           | Float64 Balance | Precise Decimal Balance")
	fmt.Println("---------------+-----------------+-------------------------")

	allUsers := ws.GetAllUsers()

	// Sort users by name for consistent display
	sort.Slice(allUsers, func(i, j int) bool {
		return allUsers[i].Name < allUsers[j].Name
	})

	for _, user := range allUsers {
		floatBalance, err := ws.GetBalance(user.ID)
		if err != nil {
			log.Printf("Failed to get float balance for %s: %v", user.Name, err)
			continue
		}

		decimalBalance, err := ws.GetBalanceDecimal(user.ID)
		if err != nil {
			log.Printf("Failed to get decimal balance for %s: %v", user.Name, err)
			continue
		}

		fmt.Printf("%-13s | $%13.2f | %s\n", user.Name, floatBalance, formatDecimal(decimalBalance))
	}

	// Display transaction history for user1 with decimal amounts
	fmt.Println("\n--- Alice's Transaction History (with Decimal Amounts) ---")
	transactions, err := ws.GetTransactionHistory("user1")
	if err != nil {
		log.Printf("Failed to get transaction history: %v", err)
	} else {
		fmt.Printf("Found %d transactions:\n", len(transactions))
		for i, tx := range transactions {
			amountFloat, _ := tx.Amount.Float64()
			fmt.Printf("%2d. %-10s: $%10.2f (precise: %12s) - %s\n",
				i+1, tx.Type, amountFloat, formatDecimal(tx.Amount), tx.Description)
		}
	}

	// Demonstrate error handling with decimal amounts
	fmt.Println("\n--- Error Handling Examples ---")

	// Try to transfer more than balance
	err = ws.Transfer("user3", "user1", 10000.0, "Large transfer")
	if err != nil {
		fmt.Printf("✓ Expected error (insufficient balance): %v\n", err)
	}

	// Try to withdraw negative amount
	err = ws.Withdraw("user1", -50.0, "Invalid withdrawal")
	if err != nil {
		fmt.Printf("✓ Expected error (invalid amount): %v\n", err)
	}

	// Try to deposit zero amount using decimal
	err = ws.DepositDecimal("user1", decimal.NewFromFloat(0.0), "Zero deposit")
	if err != nil {
		fmt.Printf("✓ Expected error (invalid amount): %v\n", err)
	}

	// Demonstrate the advantage of decimal in accumulation
	fmt.Println("\n--- Decimal Accumulation Demo ---")
	fmt.Println("Demonstrating why decimal is better for financial calculations:")

	// Reset a test user for this demo
	ws.CreateUser("test_user", "Test User", "test@example.com")

	// Add 0.1 ten times using decimal
	decimalAmount := decimal.NewFromFloat(0.1)
	for i := 0; i < 10; i++ {
		ws.DepositDecimal("test_user", decimalAmount, "Repeated deposit")
	}

	testBalance, _ := ws.GetBalanceDecimal("test_user")
	fmt.Printf("Adding 0.1 ten times using decimal: %s\n", formatDecimal(testBalance))
	fmt.Printf("Expected result: 1.0\n")
	fmt.Printf("Correct: %v\n", testBalance.Equal(decimal.NewFromFloat(1.0)))

	// Compare with what float64 would do
	var floatTotal float64
	for i := 0; i < 10; i++ {
		floatTotal += 0.1
	}
	fmt.Printf("Adding 0.1 ten times using float64: %.17f\n", floatTotal)
	fmt.Printf("Expected result: 1.0\n")
	fmt.Printf("Correct: %v\n", floatTotal == 1.0)

	// Demonstrate advanced decimal operations
	advancedDemo(ws)

	fmt.Println("\n=== Demo Completed ===")
}

// advancedDemo shows more advanced decimal usage patterns
func advancedDemo(ws *wallet.WalletService) {
	fmt.Println("\n--- Advanced Decimal Usage ---")

	ws.CreateUser("advanced_user", "Advanced User", "advanced@example.com")

	// Using decimal operations directly
	amount1 := decimal.NewFromFloat(100.75)
	amount2 := decimal.NewFromFloat(25.25)

	// Deposit the amounts
	ws.DepositDecimal("advanced_user", amount1, "Large deposit")
	ws.DepositDecimal("advanced_user", amount2, "Additional deposit")

	// Divide exactly
	splitAmount := amount1.Div(decimal.NewFromFloat(2))
	fmt.Printf("Splitting %s exactly: %s per person\n", formatDecimal(amount1), formatDecimal(splitAmount))

	// Calculate percentage
	tenPercent := amount1.Mul(decimal.NewFromFloat(0.1))
	fmt.Printf("10%% of %s: %s\n", formatDecimal(amount1), formatDecimal(tenPercent))

	// Demonstrate exact fractional amounts
	exactFraction := decimal.NewFromFloat(1.0).Div(decimal.NewFromFloat(3))
	fmt.Printf("1/3 as exact decimal: %s (repeating)\n", formatDecimal(exactFraction))

	ws.DepositDecimal("advanced_user", exactFraction, "Exact fraction deposit")
	balance, _ := ws.GetBalanceDecimal("advanced_user")
	fmt.Printf("Balance after complex operations: %s\n", formatDecimal(balance))

	// Show decimal precision in financial calculations
	fmt.Println("\n--- Financial Calculation Examples ---")

	// Calculate compound interest
	principal := decimal.NewFromFloat(1000.0)
	rate := decimal.NewFromFloat(0.05) // 5%
	years := decimal.NewFromFloat(3)

	compoundInterest := principal.Mul(
		decimal.NewFromFloat(1.0).Add(rate).Pow(years),
	).Sub(principal)

	fmt.Printf("Compound interest on $%s at %.1f%% for %s years: $%s\n",
		formatDecimal(principal), rate.Mul(decimal.NewFromFloat(100)).InexactFloat64(),
		formatDecimal(years), formatDecimal(compoundInterest))

	// Calculate tax
	income := decimal.NewFromFloat(50000.0)
	taxRate := decimal.NewFromFloat(0.22) // 22%
	taxAmount := income.Mul(taxRate)

	fmt.Printf("Tax on $%s at %.1f%% rate: $%s\n",
		formatDecimal(income), taxRate.Mul(decimal.NewFromFloat(100)).InexactFloat64(),
		formatDecimal(taxAmount))
}

// formatDecimal formats a decimal for consistent display
func formatDecimal(d decimal.Decimal) string {
	// If it's a whole number, show with .0 for consistency
	if d.Equal(d.Truncate(0)) {
		return d.StringFixed(1)
	}
	// Otherwise, show with appropriate precision
	return d.String()
}
