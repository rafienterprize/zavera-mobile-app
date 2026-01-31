package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run generate_signature.go <order_id> <status_code> <gross_amount>")
		fmt.Println("Example: go run generate_signature.go ZVR-20260106-ABC123 200 764000.00")
		return
	}

	orderID := os.Args[1]
	statusCode := os.Args[2]
	grossAmount := os.Args[3]
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	if serverKey == "" {
		serverKey = os.Getenv("MIDTRANS_SERVER_KEY")
	}	

	// Signature = SHA512(order_id + status_code + gross_amount + server_key)
	signatureInput := orderID + statusCode + grossAmount + serverKey
	hash := sha512.New()
	hash.Write([]byte(signatureInput))
	signature := hex.EncodeToString(hash.Sum(nil))

	fmt.Printf("Order ID: %s\n", orderID)
	fmt.Printf("Status Code: %s\n", statusCode)
	fmt.Printf("Gross Amount: %s\n", grossAmount)
	fmt.Printf("Server Key: %s\n", serverKey)
	fmt.Printf("\nSignature Key: %s\n", signature)
}
