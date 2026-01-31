package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type DebugHandler struct{}

func NewDebugHandler() *DebugHandler {
	return &DebugHandler{}
}

// TestMidtrans tests Midtrans connection
func (h *DebugHandler) TestMidtrans(c *gin.Context) {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	environment := os.Getenv("MIDTRANS_ENVIRONMENT")

	// Mask server key for security
	maskedKey := ""
	if len(serverKey) > 20 {
		maskedKey = serverKey[:15] + "..." + serverKey[len(serverKey)-5:]
	} else {
		maskedKey = "NOT SET or TOO SHORT"
	}

	// Initialize snap client
	var s snap.Client
	env := midtrans.Sandbox
	if environment == "production" {
		env = midtrans.Production
	}
	s.New(serverKey, env)

	// Try to create a test transaction
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  fmt.Sprintf("TEST-%d", 999999),
			GrossAmt: 10000,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: "Test",
			Email: "test@example.com",
			Phone: "08123456789",
		},
	}

	snapResp, midtransErr := s.CreateTransaction(req)

	result := gin.H{
		"server_key_masked": maskedKey,
		"environment":       environment,
		"midtrans_env":      env,
	}

	if midtransErr != nil {
		result["status"] = "error"
		result["error"] = fmt.Sprintf("%v", midtransErr)
		result["error_type"] = fmt.Sprintf("%T", midtransErr)
	} else if snapResp != nil && snapResp.Token != "" {
		result["status"] = "success"
		result["token_preview"] = snapResp.Token[:20] + "..."
		result["redirect_url"] = snapResp.RedirectURL
	} else {
		result["status"] = "error"
		result["error"] = "Empty response from Midtrans"
	}

	c.JSON(http.StatusOK, result)
}
