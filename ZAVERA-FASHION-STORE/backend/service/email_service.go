package service

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"
	"zavera/models"
	"zavera/repository"
)

// EmailService handles transactional email sending
// Following Tokopedia-style policy: only send emails for events with legal/financial impact
type EmailService interface {
	// SendOrderCreated sends email when order is created (invoice legally exists)
	SendOrderCreated(order *models.Order, items []models.OrderItem, shippingAddress string, courier, service string) error
	
	// SendPaymentSuccess sends email when payment is confirmed (money received)
	SendPaymentSuccess(order *models.Order, paymentMethod string) error
	
	// SendOrderShipped sends email when order is shipped (goods handed to courier)
	SendOrderShipped(order *models.Order, shipment *models.Shipment, shippingAddress string) error
	
	// SendOrderDelivered sends email when order is delivered (goods received)
	SendOrderDelivered(order *models.Order, shipment *models.Shipment) error
	
	// SendOrderCancelled sends email when order is cancelled (contract voided)
	SendOrderCancelled(order *models.Order, items []models.OrderItem, shippingAddress string, reason string) error
	
	// SendOrderRefunded sends email when order is refunded (money returned)
	SendOrderRefunded(order *models.Order, items []models.OrderItem, refundAmount float64, refundReason string) error
	
	// GetEmailLogs returns email logs for an order
	GetEmailLogs(orderID int) ([]models.EmailLog, error)
	
	// HasSentEmail checks if an email type has already been sent for an order
	HasSentEmail(orderID int, templateKey string) (bool, error)
}

type emailService struct {
	emailRepo repository.EmailRepository
	smtpHost  string
	smtpPort  string
	smtpUser  string
	smtpPass  string
	fromEmail string
	fromName  string
	baseURL   string
}

// NewEmailService creates a new email service
func NewEmailService(emailRepo repository.EmailRepository) EmailService {
	return &emailService{
		emailRepo: emailRepo,
		smtpHost:  getEnvOrDefault("SMTP_HOST", "smtp.gmail.com"),
		smtpPort:  getEnvOrDefault("SMTP_PORT", "587"),
		smtpUser:  getEnvOrDefault("SMTP_USERNAME", ""),  // Match .env
		smtpPass:  getEnvOrDefault("SMTP_PASSWORD", ""),  // Match .env
		fromEmail: getEnvOrDefault("SMTP_FROM", "noreply@zavera.com"),  // Match .env
		fromName:  getEnvOrDefault("SMTP_FROM_NAME", "ZAVERA"),
		baseURL:   getEnvOrDefault("FRONTEND_URL", "http://localhost:3000"),  // Match .env
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// OrderCreatedData holds data for order created email
type OrderCreatedData struct {
	CustomerName    string
	OrderCode       string
	CreatedAt       string
	Items           []OrderItemData
	Subtotal        string
	ShippingCost    string
	TotalAmount     string
	Courier         string
	Service         string
	ShippingAddress string
	PaymentURL      string
}

// OrderItemData holds item data for email
type OrderItemData struct {
	ProductName string
	Quantity    int
	Subtotal    string
}

// PaymentSuccessData holds data for payment success email
type PaymentSuccessData struct {
	CustomerName  string
	OrderCode     string
	TotalAmount   string
	PaymentMethod string
	PaidAt        string
}

// OrderShippedData holds data for order shipped email
type OrderShippedData struct {
	CustomerName    string
	OrderCode       string
	Courier         string
	Service         string
	Resi            string
	ETD             string
	TrackingURL     string
	ShippingAddress string
}

// OrderDeliveredData holds data for order delivered email
type OrderDeliveredData struct {
	CustomerName string
	OrderCode    string
	DeliveredAt  string
	Courier      string
	Service      string
	ReviewURL    string
	ShopURL      string
}

// OrderCancelledData holds data for order cancelled email
type OrderCancelledData struct {
	CustomerName       string
	OrderCode          string
	CreatedAt          string
	CancelledAt        string
	CancellationReason string
	Items              []OrderItemData
	Subtotal           string
	ShippingCost       string
	TotalAmount        string
	ShippingAddress    string
	RefundInfo         string
	ShopURL            string
}

// OrderRefundedData holds data for order refunded email
type OrderRefundedData struct {
	CustomerName  string
	OrderCode     string
	RefundCode    string
	RefundedAt    string
	RefundReason  string
	RefundAmount  string
	RefundMethod  string
	Items         []OrderItemData
	Subtotal      string
	ShippingCost  string
	ShopURL       string
}

// SendOrderCreated sends email when order is created
func (s *emailService) SendOrderCreated(order *models.Order, items []models.OrderItem, shippingAddress string, courier, service string) error {
	// Check for duplicate - don't send if already sent
	sent, err := s.HasSentEmail(order.ID, "ORDER_CREATED")
	if err == nil && sent {
		log.Printf("⚠️ ORDER_CREATED email already sent for order %s, skipping", order.OrderCode)
		return nil
	}

	// Prepare template data
	var itemsData []OrderItemData
	for _, item := range items {
		itemsData = append(itemsData, OrderItemData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Subtotal:    formatCurrency(item.Subtotal),
		})
	}

	data := OrderCreatedData{
		CustomerName:    order.CustomerName,
		OrderCode:       order.OrderCode,
		CreatedAt:       order.CreatedAt.Format("02 Jan 2006 15:04"),
		Items:           itemsData,
		Subtotal:        formatCurrency(order.Subtotal),
		ShippingCost:    formatCurrency(order.ShippingCost),
		TotalAmount:     formatCurrency(order.TotalAmount),
		Courier:         courier,
		Service:         service,
		ShippingAddress: shippingAddress,
		PaymentURL:      fmt.Sprintf("%s/checkout/payment/%s", s.baseURL, order.OrderCode),
	}

	subject := fmt.Sprintf("🛍️ Pesanan ZAVERA #%s telah dibuat", order.OrderCode)
	
	// Get template from database or use default
	htmlBody, err := s.renderTemplate("ORDER_CREATED", data)
	if err != nil {
		log.Printf("Warning: failed to render ORDER_CREATED template: %v", err)
		htmlBody = s.getDefaultOrderCreatedHTML(data)
	}

	return s.sendEmail(order.CustomerEmail, subject, htmlBody, order.ID, "ORDER_CREATED")
}

// SendPaymentSuccess sends email when payment is confirmed
func (s *emailService) SendPaymentSuccess(order *models.Order, paymentMethod string) error {
	// Check for duplicate - don't send if already sent
	sent, err := s.HasSentEmail(order.ID, "PAYMENT_SUCCESS")
	if err == nil && sent {
		log.Printf("⚠️ PAYMENT_SUCCESS email already sent for order %s, skipping", order.OrderCode)
		return nil
	}

	paidAt := time.Now()
	if order.PaidAt != nil {
		paidAt = *order.PaidAt
	}

	data := PaymentSuccessData{
		CustomerName:  order.CustomerName,
		OrderCode:     order.OrderCode,
		TotalAmount:   formatCurrency(order.TotalAmount),
		PaymentMethod: paymentMethod,
		PaidAt:        paidAt.Format("02 Jan 2006 15:04"),
	}

	subject := fmt.Sprintf("💳 Pembayaran diterima – Pesanan #%s", order.OrderCode)
	
	htmlBody, err := s.renderTemplate("PAYMENT_SUCCESS", data)
	if err != nil {
		log.Printf("Warning: failed to render PAYMENT_SUCCESS template: %v", err)
		htmlBody = s.getDefaultPaymentSuccessHTML(data)
	}

	return s.sendEmail(order.CustomerEmail, subject, htmlBody, order.ID, "PAYMENT_SUCCESS")
}

// SendOrderShipped sends email when order is shipped
func (s *emailService) SendOrderShipped(order *models.Order, shipment *models.Shipment, shippingAddress string) error {
	// Check for duplicate - don't send if already sent
	sent, err := s.HasSentEmail(order.ID, "ORDER_SHIPPED")
	if err == nil && sent {
		log.Printf("⚠️ ORDER_SHIPPED email already sent for order %s, skipping", order.OrderCode)
		return nil
	}

	trackingURL := GetTrackingURL(shipment.ProviderCode, order.Resi)

	data := OrderShippedData{
		CustomerName:    order.CustomerName,
		OrderCode:       order.OrderCode,
		Courier:         shipment.ProviderName,
		Service:         shipment.ServiceName,
		Resi:            order.Resi,
		ETD:             shipment.ETD,
		TrackingURL:     trackingURL,
		ShippingAddress: shippingAddress,
	}

	subject := fmt.Sprintf("📦 Pesanan #%s sedang dikirim", order.OrderCode)
	
	htmlBody, err := s.renderTemplate("ORDER_SHIPPED", data)
	if err != nil {
		log.Printf("Warning: failed to render ORDER_SHIPPED template: %v", err)
		htmlBody = s.getDefaultOrderShippedHTML(data)
	}

	return s.sendEmail(order.CustomerEmail, subject, htmlBody, order.ID, "ORDER_SHIPPED")
}

// SendOrderDelivered sends email when order is delivered
func (s *emailService) SendOrderDelivered(order *models.Order, shipment *models.Shipment) error {
	// Check for duplicate - don't send if already sent
	sent, err := s.HasSentEmail(order.ID, "ORDER_DELIVERED")
	if err == nil && sent {
		log.Printf("⚠️ ORDER_DELIVERED email already sent for order %s, skipping", order.OrderCode)
		return nil
	}

	deliveredAt := time.Now()
	if order.DeliveredAt != nil {
		deliveredAt = *order.DeliveredAt
	}

	data := OrderDeliveredData{
		CustomerName: order.CustomerName,
		OrderCode:    order.OrderCode,
		DeliveredAt:  deliveredAt.Format("02 Jan 2006 15:04"),
		Courier:      shipment.ProviderName,
		Service:      shipment.ServiceName,
		ReviewURL:    fmt.Sprintf("%s/orders/%s/review", s.baseURL, order.OrderCode),
		ShopURL:      s.baseURL,
	}

	subject := fmt.Sprintf("🎉 Pesanan #%s sudah sampai", order.OrderCode)
	
	htmlBody, err := s.renderTemplate("ORDER_DELIVERED", data)
	if err != nil {
		log.Printf("Warning: failed to render ORDER_DELIVERED template: %v", err)
		htmlBody = s.getDefaultOrderDeliveredHTML(data)
	}

	return s.sendEmail(order.CustomerEmail, subject, htmlBody, order.ID, "ORDER_DELIVERED")
}

// GetEmailLogs returns email logs for an order
func (s *emailService) GetEmailLogs(orderID int) ([]models.EmailLog, error) {
	return s.emailRepo.GetEmailLogsByOrder(orderID)
}

// HasSentEmail checks if an email type has already been sent for an order
func (s *emailService) HasSentEmail(orderID int, templateKey string) (bool, error) {
	return s.emailRepo.HasSentEmail(orderID, templateKey)
}

// SendOrderCancelled sends email when order is cancelled
func (s *emailService) SendOrderCancelled(order *models.Order, items []models.OrderItem, shippingAddress string, reason string) error {
	// Check for duplicate - don't send if already sent
	sent, err := s.HasSentEmail(order.ID, "ORDER_CANCELLED")
	if err == nil && sent {
		log.Printf("⚠️ ORDER_CANCELLED email already sent for order %s, skipping", order.OrderCode)
		return nil
	}

	// Prepare template data
	var itemsData []OrderItemData
	for _, item := range items {
		itemsData = append(itemsData, OrderItemData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Subtotal:    formatCurrency(item.Subtotal),
		})
	}

	cancelledAt := time.Now()
	if order.CancelledAt != nil {
		cancelledAt = *order.CancelledAt
	}

	// Determine refund info based on payment status
	refundInfo := ""
	if order.PaidAt != nil {
		refundInfo = "Dana akan dikembalikan ke metode pembayaran asal dalam 3-14 hari kerja."
	}

	data := OrderCancelledData{
		CustomerName:       order.CustomerName,
		OrderCode:          order.OrderCode,
		CreatedAt:          order.CreatedAt.Format("02 Jan 2006 15:04"),
		CancelledAt:        cancelledAt.Format("02 Jan 2006 15:04"),
		CancellationReason: reason,
		Items:              itemsData,
		Subtotal:           formatCurrency(order.Subtotal),
		ShippingCost:       formatCurrency(order.ShippingCost),
		TotalAmount:        formatCurrency(order.TotalAmount),
		ShippingAddress:    shippingAddress,
		RefundInfo:         refundInfo,
		ShopURL:            s.baseURL,
	}

	subject := fmt.Sprintf("❌ Pesanan #%s telah dibatalkan", order.OrderCode)
	
	htmlBody, err := s.renderTemplate("ORDER_CANCELLED", data)
	if err != nil {
		log.Printf("Warning: failed to render ORDER_CANCELLED template: %v", err)
		htmlBody = s.getDefaultOrderCancelledHTML(data)
	}

	return s.sendEmail(order.CustomerEmail, subject, htmlBody, order.ID, "ORDER_CANCELLED")
}

// SendOrderRefunded sends email when order is refunded
func (s *emailService) SendOrderRefunded(order *models.Order, items []models.OrderItem, refundAmount float64, refundReason string) error {
	// Check for duplicate - don't send if already sent
	sent, err := s.HasSentEmail(order.ID, "ORDER_REFUNDED")
	if err == nil && sent {
		log.Printf("⚠️ ORDER_REFUNDED email already sent for order %s, skipping", order.OrderCode)
		return nil
	}

	// Prepare template data
	var itemsData []OrderItemData
	for _, item := range items {
		itemsData = append(itemsData, OrderItemData{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Subtotal:    formatCurrency(item.Subtotal),
		})
	}

	data := OrderRefundedData{
		CustomerName:  order.CustomerName,
		OrderCode:     order.OrderCode,
		RefundCode:    fmt.Sprintf("REF-%s", order.OrderCode),
		RefundedAt:    time.Now().Format("02 Jan 2006 15:04"),
		RefundReason:  refundReason,
		RefundAmount:  formatCurrency(refundAmount),
		RefundMethod:  "Metode pembayaran asal",
		Items:         itemsData,
		Subtotal:      formatCurrency(order.Subtotal),
		ShippingCost:  formatCurrency(order.ShippingCost),
		ShopURL:       s.baseURL,
	}

	subject := fmt.Sprintf("💰 Refund untuk Pesanan #%s telah diproses", order.OrderCode)
	
	htmlBody, err := s.renderTemplate("ORDER_REFUNDED", data)
	if err != nil {
		log.Printf("Warning: failed to render ORDER_REFUNDED template: %v", err)
		htmlBody = s.getDefaultOrderRefundedHTML(data)
	}

	return s.sendEmail(order.CustomerEmail, subject, htmlBody, order.ID, "ORDER_REFUNDED")
}

// renderTemplate renders an email template with data
func (s *emailService) renderTemplate(templateKey string, data interface{}) (string, error) {
	// Get template from database
	emailTemplate, err := s.emailRepo.GetTemplateByKey(templateKey)
	if err != nil {
		return "", err
	}

	// Parse and execute template
	tmpl, err := template.New(templateKey).Parse(emailTemplate.HTMLTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// sendEmail sends an email and logs it
func (s *emailService) sendEmail(to, subject, htmlBody string, orderID int, templateKey string) error {
	// Create email log entry
	emailLog := &models.EmailLog{
		OrderID:        &orderID,
		TemplateKey:    templateKey,
		RecipientEmail: to,
		Subject:        subject,
		Status:         models.EmailLogStatusPending,
	}

	// Try to send email
	err := s.sendSMTP(to, subject, htmlBody)
	if err != nil {
		emailLog.Status = models.EmailLogStatusFailed
		emailLog.ErrorMessage = err.Error()
		s.emailRepo.CreateEmailLog(emailLog)
		return err
	}

	// Mark as sent
	now := time.Now()
	emailLog.Status = models.EmailLogStatusSent
	emailLog.SentAt = &now
	s.emailRepo.CreateEmailLog(emailLog)

	return nil
}

// sendSMTP sends email via SMTP
func (s *emailService) sendSMTP(to, subject, htmlBody string) error {
	// Skip if SMTP not configured
	if s.smtpUser == "" || s.smtpPass == "" {
		log.Printf("📧 [MOCK] Email to %s: %s", to, subject)
		return nil
	}

	// Build email message
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	// Send via SMTP
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("📧 Email sent to %s: %s", to, subject)
	return nil
}

// formatCurrency formats a number as Indonesian Rupiah
func formatCurrency(amount float64) string {
	// Simple formatting - in production use a proper library
	return fmt.Sprintf("%.0f", amount)
}

// Default HTML templates (fallback if database templates fail)
func (s *emailService) getDefaultOrderCreatedHTML(data OrderCreatedData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>ZAVERA</h1>
<h2>Pesanan Anda Telah Dibuat!</h2>
<p>Halo %s,</p>
<p>Terima kasih telah berbelanja di ZAVERA. Pesanan Anda telah berhasil dibuat.</p>
<p><strong>Nomor Pesanan:</strong> %s</p>
<p><strong>Total:</strong> Rp %s</p>
<p><strong>Kurir:</strong> %s - %s</p>
<p><a href="%s">Bayar Sekarang</a></p>
</body>
</html>`, data.CustomerName, data.OrderCode, data.TotalAmount, data.Courier, data.Service, data.PaymentURL)
}

func (s *emailService) getDefaultPaymentSuccessHTML(data PaymentSuccessData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>ZAVERA</h1>
<h2>✓ Pembayaran Berhasil!</h2>
<p>Halo %s,</p>
<p>Pembayaran untuk pesanan Anda telah kami terima.</p>
<p><strong>Nomor Pesanan:</strong> %s</p>
<p><strong>Total:</strong> Rp %s</p>
<p><strong>Metode:</strong> %s</p>
</body>
</html>`, data.CustomerName, data.OrderCode, data.TotalAmount, data.PaymentMethod)
}

func (s *emailService) getDefaultOrderShippedHTML(data OrderShippedData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>ZAVERA</h1>
<h2>📦 Pesanan Sedang Dikirim!</h2>
<p>Halo %s,</p>
<p>Pesanan Anda telah dikirim.</p>
<p><strong>Nomor Pesanan:</strong> %s</p>
<p><strong>Kurir:</strong> %s - %s</p>
<p><strong>Nomor Resi:</strong> %s</p>
<p><strong>Estimasi:</strong> %s</p>
<p><a href="%s">Lacak Pengiriman</a></p>
</body>
</html>`, data.CustomerName, data.OrderCode, data.Courier, data.Service, data.Resi, data.ETD, data.TrackingURL)
}

func (s *emailService) getDefaultOrderDeliveredHTML(data OrderDeliveredData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>ZAVERA</h1>
<h2>🎉 Pesanan Telah Sampai!</h2>
<p>Halo %s,</p>
<p>Pesanan Anda telah berhasil diterima.</p>
<p><strong>Nomor Pesanan:</strong> %s</p>
<p><strong>Tanggal Diterima:</strong> %s</p>
<p><a href="%s">Beri Ulasan</a> | <a href="%s">Belanja Lagi</a></p>
</body>
</html>`, data.CustomerName, data.OrderCode, data.DeliveredAt, data.ReviewURL, data.ShopURL)
}

func (s *emailService) getDefaultOrderCancelledHTML(data OrderCancelledData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>ZAVERA</h1>
<h2>❌ Pesanan Dibatalkan</h2>
<p>Halo %s,</p>
<p>Pesanan Anda telah dibatalkan.</p>
<p><strong>Nomor Pesanan:</strong> %s</p>
<p><strong>Tanggal Dibatalkan:</strong> %s</p>
<p><strong>Alasan:</strong> %s</p>
<p><strong>Total:</strong> Rp %s</p>
<p>%s</p>
<p><a href="%s">Belanja Lagi</a></p>
</body>
</html>`, data.CustomerName, data.OrderCode, data.CancelledAt, data.CancellationReason, data.TotalAmount, data.RefundInfo, data.ShopURL)
}

func (s *emailService) getDefaultOrderRefundedHTML(data OrderRefundedData) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
<h1>ZAVERA</h1>
<h2>💰 Refund Berhasil Diproses</h2>
<p>Halo %s,</p>
<p>Pengembalian dana untuk pesanan Anda telah berhasil diproses.</p>
<p><strong>Nomor Pesanan:</strong> %s</p>
<p><strong>Nomor Refund:</strong> %s</p>
<p><strong>Jumlah Refund:</strong> Rp %s</p>
<p><strong>Alasan:</strong> %s</p>
<p>Dana akan masuk ke rekening/saldo Anda dalam 3-14 hari kerja.</p>
<p><a href="%s">Belanja Lagi</a></p>
</body>
</html>`, data.CustomerName, data.OrderCode, data.RefundCode, data.RefundAmount, data.RefundReason, data.ShopURL)
}
