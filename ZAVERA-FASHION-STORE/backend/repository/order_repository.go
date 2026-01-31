package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
	"zavera/models"
)

type OrderRepository interface {
	Create(order *models.Order, items []models.OrderItem) error
	FindByID(id int) (*models.Order, error)
	FindByOrderCode(orderCode string) (*models.Order, error)
	FindByOrderCodeForUpdate(orderCode string) (*models.Order, *sql.Tx, error)
	FindExpiredPendingOrders(maxAge time.Duration) ([]*models.Order, error)
	GetOrderItems(orderID int) ([]models.OrderItem, error)
	UpdateStatus(orderID int, status models.OrderStatus) error
	UpdateStatusTx(tx *sql.Tx, orderID int, status models.OrderStatus) error
	MarkAsPaid(orderID int) error
	MarkAsPaidTx(tx *sql.Tx, orderID int) error
	MarkAsPacking(orderID int) error
	MarkAsShipped(orderID int) error
	MarkAsShippedWithResi(orderID int, resi string) error
	MarkAsDelivered(orderID int) error
	MarkAsCompleted(orderID int) error
	MarkAsCancelled(orderID int) error
	MarkAsExpired(orderID int) error
	MarkAsRefunded(orderID int) error
	UpdateResi(orderID int, resi string) error
	RestoreStock(orderID int) error
	RestoreStockTx(tx *sql.Tx, orderID int) error
	RecordStatusChange(orderID int, fromStatus, toStatus models.OrderStatus, changedBy, reason string) error
	IsResiExists(resi string) (bool, error)
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order, items []models.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate unique order code
	order.OrderCode = r.generateOrderCode()
	order.StockReserved = true // Stock will be reserved

	// Step 1: Validate and reserve stock within transaction
	for _, item := range items {
		log.Printf("🔍 Checking stock for item: product_id=%d, variant_id=%v, quantity=%d", 
			item.ProductID, item.VariantID, item.Quantity)
		
		// Check if this is a variant product
		if item.VariantID != nil && *item.VariantID > 0 {
			log.Printf("✅ Variant product detected: variant_id=%d", *item.VariantID)
			// Variant product - check and reserve variant stock
			var currentStock int
			stockQuery := `
				SELECT stock_quantity FROM product_variants WHERE id = $1 FOR UPDATE
			`
			err := tx.QueryRow(stockQuery, *item.VariantID).Scan(&currentStock)
			if err != nil {
				return fmt.Errorf("failed to check stock for variant %d: %w", *item.VariantID, err)
			}
			
			log.Printf("📦 Variant stock: variant_id=%d, stock=%d, requested=%d", 
				*item.VariantID, currentStock, item.Quantity)

			if currentStock < item.Quantity {
				return fmt.Errorf("insufficient stock for product %s: available %d, requested %d",
					item.ProductName, currentStock, item.Quantity)
			}

			// Reserve variant stock (deduct)
			reserveQuery := `
				UPDATE product_variants SET stock_quantity = stock_quantity - $1 WHERE id = $2
			`
			_, err = tx.Exec(reserveQuery, item.Quantity, *item.VariantID)
			if err != nil {
				return fmt.Errorf("failed to reserve stock for variant %d: %w", *item.VariantID, err)
			}
		} else {
			log.Printf("⚠️ Simple product (no variant_id): product_id=%d", item.ProductID)
			// Simple product - check and reserve product stock
			var currentStock int
			stockQuery := `
				SELECT stock FROM products WHERE id = $1 FOR UPDATE
			`
			err := tx.QueryRow(stockQuery, item.ProductID).Scan(&currentStock)
			if err != nil {
				return fmt.Errorf("failed to check stock for product %d: %w", item.ProductID, err)
			}

			if currentStock < item.Quantity {
				return fmt.Errorf("insufficient stock for product %s: available %d, requested %d",
					item.ProductName, currentStock, item.Quantity)
			}

			// Reserve stock (deduct)
			reserveQuery := `
				UPDATE products SET stock = stock - $1 WHERE id = $2
			`
			_, err = tx.Exec(reserveQuery, item.Quantity, item.ProductID)
			if err != nil {
				return fmt.Errorf("failed to reserve stock for product %d: %w", item.ProductID, err)
			}
		}
	}

	// Step 2: Insert order
	metadataJSON, _ := json.Marshal(order.Metadata)
	orderQuery := `
		INSERT INTO orders (
			order_code, user_id, customer_name, customer_email, customer_phone,
			subtotal, shipping_cost, tax, discount, total_amount, status, 
			stock_reserved, notes, metadata, origin_city, destination_city
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRow(
		orderQuery,
		order.OrderCode, order.UserID, order.CustomerName, order.CustomerEmail, order.CustomerPhone,
		order.Subtotal, order.ShippingCost, order.Tax, order.Discount, order.TotalAmount,
		order.Status, order.StockReserved, order.Notes, metadataJSON,
		order.OriginCity, order.DestinationCity,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return err
	}

	// Step 3: Insert order items and record stock movements
	itemQuery := `
		INSERT INTO order_items (
			order_id, product_id, variant_id, product_name, quantity, price_per_unit, subtotal, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	stockMovementQuery := `
		INSERT INTO stock_movements (product_id, order_id, movement_type, quantity, balance_after, notes)
		SELECT $1, $2, 'RESERVE', $3, stock, 'Stock reserved at checkout'
		FROM products WHERE id = $1
	`

	for _, item := range items {
		itemMetadataJSON, _ := json.Marshal(item.Metadata)
		_, err := tx.Exec(
			itemQuery,
			order.ID, item.ProductID, item.VariantID, item.ProductName, item.Quantity,
			item.PricePerUnit, item.Subtotal, itemMetadataJSON,
		)
		if err != nil {
			return err
		}

		// Record stock movement for audit
		_, err = tx.Exec(stockMovementQuery, item.ProductID, order.ID, item.Quantity)
		if err != nil {
			// Non-critical, log but don't fail
			fmt.Printf("Warning: failed to record stock movement for product %d: %v\n", item.ProductID, err)
		}
	}

	// Step 4: Record initial status
	historyQuery := `
		INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, reason)
		VALUES ($1, NULL, $2, 'system', 'Order created')
	`
	tx.Exec(historyQuery, order.ID, order.Status) // Ignore error, non-critical

	return tx.Commit()
}

func (r *orderRepository) FindByID(id int) (*models.Order, error) {
	query := `
		SELECT id, order_code, user_id, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status,
		       COALESCE(stock_reserved, true) as stock_reserved, 
		       COALESCE(resi, '') as resi,
		       COALESCE(origin_city, 'Semarang') as origin_city,
		       COALESCE(destination_city, '') as destination_city,
		       notes, metadata, created_at, updated_at, 
		       paid_at, shipped_at, delivered_at, completed_at, cancelled_at
		FROM orders
		WHERE id = $1
	`

	var order models.Order
	var metadataJSON []byte
	err := r.db.QueryRow(query, id).Scan(
		&order.ID, &order.OrderCode, &order.UserID, &order.CustomerName,
		&order.CustomerEmail, &order.CustomerPhone, &order.Subtotal, &order.ShippingCost,
		&order.Tax, &order.Discount, &order.TotalAmount, &order.Status, &order.StockReserved,
		&order.Resi, &order.OriginCity, &order.DestinationCity,
		&order.Notes, &metadataJSON, &order.CreatedAt, &order.UpdatedAt,
		&order.PaidAt, &order.ShippedAt, &order.DeliveredAt, &order.CompletedAt, &order.CancelledAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &order.Metadata)
	}

	// Load items
	items, _ := r.findItemsByOrderID(order.ID)
	order.Items = items

	return &order, nil
}

func (r *orderRepository) FindByOrderCode(orderCode string) (*models.Order, error) {
	query := `
		SELECT id, order_code, user_id, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status,
		       COALESCE(stock_reserved, true) as stock_reserved,
		       COALESCE(resi, '') as resi,
		       COALESCE(origin_city, 'Semarang') as origin_city,
		       COALESCE(destination_city, '') as destination_city,
		       notes, metadata, created_at, updated_at, 
		       paid_at, shipped_at, delivered_at, completed_at, cancelled_at
		FROM orders
		WHERE order_code = $1
	`

	var order models.Order
	var metadataJSON []byte
	err := r.db.QueryRow(query, orderCode).Scan(
		&order.ID, &order.OrderCode, &order.UserID, &order.CustomerName,
		&order.CustomerEmail, &order.CustomerPhone, &order.Subtotal, &order.ShippingCost,
		&order.Tax, &order.Discount, &order.TotalAmount, &order.Status, &order.StockReserved,
		&order.Resi, &order.OriginCity, &order.DestinationCity,
		&order.Notes, &metadataJSON, &order.CreatedAt, &order.UpdatedAt,
		&order.PaidAt, &order.ShippedAt, &order.DeliveredAt, &order.CompletedAt, &order.CancelledAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &order.Metadata)
	}

	// Load items
	items, _ := r.findItemsByOrderID(order.ID)
	order.Items = items

	return &order, nil
}

// GetOrderItems returns order items for an order
func (r *orderRepository) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	return r.findItemsByOrderID(orderID)
}

// FindByOrderCodeForUpdate finds order with row lock for atomic updates
func (r *orderRepository) FindByOrderCodeForUpdate(orderCode string) (*models.Order, *sql.Tx, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, nil, err
	}

	query := `
		SELECT id, order_code, user_id, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status,
		       COALESCE(stock_reserved, true) as stock_reserved, notes,
		       metadata, created_at, updated_at, paid_at, shipped_at, completed_at, cancelled_at
		FROM orders
		WHERE order_code = $1
		FOR UPDATE
	`

	var order models.Order
	var metadataJSON []byte
	err = tx.QueryRow(query, orderCode).Scan(
		&order.ID, &order.OrderCode, &order.UserID, &order.CustomerName,
		&order.CustomerEmail, &order.CustomerPhone, &order.Subtotal, &order.ShippingCost,
		&order.Tax, &order.Discount, &order.TotalAmount, &order.Status, &order.StockReserved,
		&order.Notes, &metadataJSON, &order.CreatedAt, &order.UpdatedAt,
		&order.PaidAt, &order.ShippedAt, &order.CompletedAt, &order.CancelledAt,
	)

	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}

	// Parse metadata
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &order.Metadata)
	}

	return &order, tx, nil
}

func (r *orderRepository) UpdateStatus(orderID int, status models.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, status, orderID)
	return err
}

func (r *orderRepository) UpdateStatusTx(tx *sql.Tx, orderID int, status models.OrderStatus) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := tx.Exec(query, status, orderID)
	return err
}

func (r *orderRepository) MarkAsPaid(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, paid_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusPaid, orderID)
	return err
}

func (r *orderRepository) MarkAsPaidTx(tx *sql.Tx, orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, paid_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := tx.Exec(query, models.OrderStatusPaid, orderID)
	return err
}

func (r *orderRepository) MarkAsShipped(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, shipped_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusShipped, orderID)
	return err
}

func (r *orderRepository) MarkAsCompleted(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, completed_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusCompleted, orderID)
	return err
}

func (r *orderRepository) MarkAsCancelled(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, cancelled_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusCancelled, orderID)
	return err
}

func (r *orderRepository) MarkAsExpired(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusExpired, orderID)
	return err
}

// MarkAsPacking marks an order as being packed
func (r *orderRepository) MarkAsPacking(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusPacking, orderID)
	return err
}

// MarkAsShippedWithResi marks an order as shipped with a resi number
func (r *orderRepository) MarkAsShippedWithResi(orderID int, resi string) error {
	query := `
		UPDATE orders
		SET status = $1, resi = $2, shipped_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, models.OrderStatusShipped, resi, orderID)
	return err
}

// MarkAsDelivered marks an order as delivered
func (r *orderRepository) MarkAsDelivered(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, delivered_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusDelivered, orderID)
	return err
}

// MarkAsRefunded marks an order as refunded
func (r *orderRepository) MarkAsRefunded(orderID int) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(query, models.OrderStatusRefunded, orderID)
	return err
}

// UpdateResi updates the resi for an order (only if not locked)
func (r *orderRepository) UpdateResi(orderID int, resi string) error {
	// Check if order is in a state where resi can be updated
	var status models.OrderStatus
	err := r.db.QueryRow("SELECT status FROM orders WHERE id = $1", orderID).Scan(&status)
	if err != nil {
		return err
	}

	// Resi is locked after shipping
	switch status {
	case models.OrderStatusShipped, models.OrderStatusDelivered, models.OrderStatusCompleted, models.OrderStatusRefunded:
		return fmt.Errorf("resi cannot be modified after order is shipped")
	}

	query := `
		UPDATE orders
		SET resi = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err = r.db.Exec(query, resi, orderID)
	return err
}

// IsResiExists checks if a resi already exists
func (r *orderRepository) IsResiExists(resi string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM orders WHERE resi = $1)`
	err := r.db.QueryRow(query, resi).Scan(&exists)
	return exists, err
}

// RestoreStock restores stock for all items in an order (used when order is cancelled/expired/failed)
func (r *orderRepository) RestoreStock(orderID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = r.restoreStockInternal(tx, orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RestoreStockTx restores stock within an existing transaction
func (r *orderRepository) RestoreStockTx(tx *sql.Tx, orderID int) error {
	return r.restoreStockInternal(tx, orderID)
}

func (r *orderRepository) restoreStockInternal(tx *sql.Tx, orderID int) error {
	// Check if stock was already restored (no lock needed, already locked by parent)
	var stockReserved bool
	checkQuery := `SELECT COALESCE(stock_reserved, true) FROM orders WHERE id = $1`
	err := tx.QueryRow(checkQuery, orderID).Scan(&stockReserved)
	if err != nil {
		return err
	}

	// Idempotency: if stock already restored, skip
	if !stockReserved {
		return nil
	}

	// Get order items
	itemsQuery := `
		SELECT product_id, quantity FROM order_items WHERE order_id = $1
	`
	rows, err := tx.Query(itemsQuery, orderID)
	if err != nil {
		return err
	}

	// Collect items first, then close rows before doing updates
	type item struct {
		productID int
		quantity  int
	}
	var items []item
	
	for rows.Next() {
		var productID, quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			rows.Close()
			return err
		}
		items = append(items, item{productID, quantity})
	}
	rows.Close() // Close rows before doing updates in same transaction

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return err
	}

	// Restore stock for each item and record movement
	stockMovementQuery := `
		INSERT INTO stock_movements (product_id, order_id, movement_type, quantity, balance_after, notes)
		SELECT $1, $2, 'RELEASE', $3, stock, 'Stock released on cancel/expire'
		FROM products WHERE id = $1
	`

	for _, itm := range items {
		restoreQuery := `
			UPDATE products SET stock = stock + $1 WHERE id = $2
		`
		_, err = tx.Exec(restoreQuery, itm.quantity, itm.productID)
		if err != nil {
			return err
		}

		// Record stock movement for audit
		_, err = tx.Exec(stockMovementQuery, itm.productID, orderID, itm.quantity)
		if err != nil {
			// Non-critical, log but don't fail
			fmt.Printf("Warning: failed to record stock release movement for product %d: %v\n", itm.productID, err)
		}
	}

	// Mark stock as restored
	markQuery := `UPDATE orders SET stock_reserved = false WHERE id = $1`
	_, err = tx.Exec(markQuery, orderID)
	return err
}

// RecordStatusChange records a status change in the audit history
func (r *orderRepository) RecordStatusChange(orderID int, fromStatus, toStatus models.OrderStatus, changedBy, reason string) error {
	query := `
		INSERT INTO order_status_history (order_id, from_status, to_status, changed_by, reason)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, orderID, fromStatus, toStatus, changedBy, reason)
	return err
}

func (r *orderRepository) findItemsByOrderID(orderID int) ([]models.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.product_name, oi.quantity,
		       oi.price_per_unit, oi.subtotal, oi.metadata, oi.created_at,
		       COALESCE(
		           (SELECT image_url FROM product_images WHERE product_id = oi.product_id ORDER BY is_primary DESC, display_order ASC LIMIT 1),
		           ''
		       ) as product_image
		FROM order_items oi
		WHERE oi.order_id = $1
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		var metadataJSON []byte

		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.PricePerUnit, &item.Subtotal,
			&metadataJSON, &item.CreatedAt, &item.ProductImage,
		)
		if err != nil {
			continue
		}

		// Parse metadata
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &item.Metadata)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *orderRepository) generateOrderCode() string {
	// Generate a unique, payment-gateway-safe order code
	// Format: ZVR-YYYYMMDD-XXXXXXXX (brand prefix + date + random hex)
	now := time.Now()
	dateStr := now.Format("20060102")

	// Generate 4 random bytes = 8 hex characters
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := strings.ToUpper(hex.EncodeToString(randomBytes))

	return fmt.Sprintf("ZVR-%s-%s", dateStr, randomHex)
}

// FindExpiredPendingOrders finds orders that are PENDING and older than maxAge
// Used by the order expiry job to auto-cancel unpaid orders (Tokopedia-style 24h limit)
func (r *orderRepository) FindExpiredPendingOrders(maxAge time.Duration) ([]*models.Order, error) {
	cutoffTime := time.Now().Add(-maxAge)
	
	query := `
		SELECT id, order_code, user_id, customer_name, customer_email, customer_phone,
		       subtotal, shipping_cost, tax, discount, total_amount, status, resi,
		       notes, metadata, stock_reserved, created_at, updated_at
		FROM orders
		WHERE status = 'PENDING' AND created_at < $1
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.Query(query, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		var metadataJSON []byte
		var resi sql.NullString
		var notes sql.NullString
		
		err := rows.Scan(
			&o.ID, &o.OrderCode, &o.UserID, &o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
			&o.Subtotal, &o.ShippingCost, &o.Tax, &o.Discount, &o.TotalAmount, &o.Status, &resi,
			&notes, &metadataJSON, &o.StockReserved, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		if resi.Valid {
			o.Resi = resi.String
		}
		if notes.Valid {
			o.Notes = notes.String
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &o.Metadata)
		}
		
		orders = append(orders, &o)
	}
	
	return orders, nil
}
