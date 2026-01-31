package service

import (
	"fmt"
	"log"
	"time"
)

// NotificationSeverity represents the severity level of a notification
type NotificationSeverity string

const (
	SeverityInfo     NotificationSeverity = "info"
	SeverityWarning  NotificationSeverity = "warning"
	SeverityCritical NotificationSeverity = "critical"
)

// AdminNotification represents a notification sent to admin
type AdminNotification struct {
	ID        string               `json:"id"`
	Type      string               `json:"type"`
	Title     string               `json:"title"`
	Message   string               `json:"message"`
	Severity  NotificationSeverity `json:"severity"`
	Data      interface{}          `json:"data,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
	Read      bool                 `json:"read"`
}

// Notification types
const (
	NotifOrderCreated    = "order_created"
	NotifPaymentReceived = "payment_received"
	NotifPaymentExpired  = "payment_expired"
	NotifShipmentUpdate  = "shipment_update"
	NotifStockLow        = "stock_low"
	NotifRefundRequest   = "refund_request"
	NotifDisputeCreated  = "dispute_created"
	NotifUserRegistered  = "user_registered"
	NotifUserLogin       = "user_login"
)

// Global notification channel
var NotificationChannel = make(chan AdminNotification, 256)

// BroadcastNotification sends notification to the channel
func BroadcastNotification(notif AdminNotification) {
	if notif.ID == "" {
		notif.ID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix()%1000)
	}
	if notif.Timestamp.IsZero() {
		notif.Timestamp = time.Now()
	}
	
	log.Printf("📢 Broadcasting notification: type=%s, title=%s", notif.Type, notif.Title)
	
	// Send directly to SSE broker
	broker := GetSSEBroker()
	if broker != nil {
		select {
		case broker.broadcast <- notif:
			log.Printf("✅ Notification sent to SSE broker successfully")
		default:
			log.Printf("⚠️ SSE broker broadcast channel full")
		}
	} else {
		log.Printf("⚠️ SSE broker not initialized")
	}
}

// NotifyOrderCreated sends notification when new order is created
func NotifyOrderCreated(orderCode string, customerName string, totalAmount float64) {
	BroadcastNotification(AdminNotification{
		Type:     NotifOrderCreated,
		Title:    "🛍️ New Order",
		Message:  fmt.Sprintf("%s placed an order #%s (Rp %s)", customerName, orderCode, formatRupiah(totalAmount)),
		Severity: SeverityInfo,
		Data: map[string]interface{}{
			"order_code":   orderCode,
			"customer":     customerName,
			"total_amount": totalAmount,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyPaymentReceived sends notification when payment is confirmed
func NotifyPaymentReceived(orderCode string, paymentMethod string, amount float64) {
	BroadcastNotification(AdminNotification{
		Type:     NotifPaymentReceived,
		Title:    "💰 Payment Received",
		Message:  fmt.Sprintf("Order #%s paid via %s (Rp %s)", orderCode, paymentMethod, formatRupiah(amount)),
		Severity: SeverityInfo,
		Data: map[string]interface{}{
			"order_code":     orderCode,
			"payment_method": paymentMethod,
			"amount":         amount,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyPaymentExpired sends notification when payment expires
func NotifyPaymentExpired(orderCode string) {
	BroadcastNotification(AdminNotification{
		Type:     NotifPaymentExpired,
		Title:    "⏰ Payment Expired",
		Message:  fmt.Sprintf("Order #%s payment has expired", orderCode),
		Severity: SeverityWarning,
		Data: map[string]interface{}{
			"order_code": orderCode,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyShipmentUpdate sends notification for shipment status changes
func NotifyShipmentUpdate(orderCode string, status string, courierName string) {
	BroadcastNotification(AdminNotification{
		Type:     NotifShipmentUpdate,
		Title:    "📦 Shipment Update",
		Message:  fmt.Sprintf("Order #%s - %s (%s)", orderCode, status, courierName),
		Severity: SeverityInfo,
		Data: map[string]interface{}{
			"order_code":   orderCode,
			"status":       status,
			"courier_name": courierName,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyStockLow sends notification when product stock is low
func NotifyStockLow(productName string, currentStock int) {
	BroadcastNotification(AdminNotification{
		Type:     NotifStockLow,
		Title:    "⚠️ Low Stock Alert",
		Message:  fmt.Sprintf("%s has only %d items left", productName, currentStock),
		Severity: SeverityWarning,
		Data: map[string]interface{}{
			"product_name":  productName,
			"current_stock": currentStock,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyRefundRequest sends notification when refund is requested
func NotifyRefundRequest(orderCode string, amount float64, reason string) {
	BroadcastNotification(AdminNotification{
		Type:     NotifRefundRequest,
		Title:    "💸 Refund Request",
		Message:  fmt.Sprintf("Refund requested for order #%s (Rp %s)", orderCode, formatRupiah(amount)),
		Severity: SeverityWarning,
		Data: map[string]interface{}{
			"order_code": orderCode,
			"amount":     amount,
			"reason":     reason,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyDisputeCreated sends notification when dispute is created
func NotifyDisputeCreated(orderCode string, disputeCode string, reason string) {
	BroadcastNotification(AdminNotification{
		Type:     NotifDisputeCreated,
		Title:    "⚠️ New Dispute",
		Message:  fmt.Sprintf("Dispute #%s created for order #%s", disputeCode, orderCode),
		Severity: SeverityCritical,
		Data: map[string]interface{}{
			"order_code":   orderCode,
			"dispute_code": disputeCode,
			"reason":       reason,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// formatRupiah formats number to Rupiah currency
func formatRupiah(amount float64) string {
	return fmt.Sprintf("%.0f", amount)
}

// NotifyUserRegistered sends notification when new user registers
func NotifyUserRegistered(email string, name string) {
	BroadcastNotification(AdminNotification{
		Type:     NotifUserRegistered,
		Title:    "👤 New User Registration",
		Message:  fmt.Sprintf("%s (%s) just registered", name, email),
		Severity: SeverityInfo,
		Data: map[string]interface{}{
			"email": email,
			"name":  name,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}

// NotifyUserLogin sends notification when user logs in
func NotifyUserLogin(email string, name string) {
	BroadcastNotification(AdminNotification{
		Type:     NotifUserLogin,
		Title:    "🔐 User Login",
		Message:  fmt.Sprintf("%s logged in", name),
		Severity: SeverityInfo,
		Data: map[string]interface{}{
			"email": email,
			"name":  name,
		},
		Timestamp: time.Now(),
		Read:      false,
	})
}
