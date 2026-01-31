package repository

import (
	"database/sql"
	"zavera/models"
)

// StockRepository handles stock movement operations
type StockRepository interface {
	// RecordMovement records a stock movement and returns the new balance
	RecordMovement(productID int, orderID *int, movementType models.StockMovementType, quantity int, notes string) (*models.StockMovement, error)
	
	// RecordMovementTx records a stock movement within a transaction
	RecordMovementTx(tx *sql.Tx, productID int, orderID *int, movementType models.StockMovementType, quantity int, notes string) (*models.StockMovement, error)
	
	// GetMovementsByProduct returns all movements for a product
	GetMovementsByProduct(productID int) ([]models.StockMovement, error)
	
	// GetMovementsByOrder returns all movements for an order
	GetMovementsByOrder(orderID int) ([]models.StockMovement, error)
	
	// GetCurrentStock returns the current stock for a product
	GetCurrentStock(productID int) (int, error)
}

type stockRepository struct {
	db *sql.DB
}

// NewStockRepository creates a new stock repository
func NewStockRepository(db *sql.DB) StockRepository {
	return &stockRepository{db: db}
}

// RecordMovement records a stock movement
func (r *stockRepository) RecordMovement(productID int, orderID *int, movementType models.StockMovementType, quantity int, notes string) (*models.StockMovement, error) {
	// Get current stock balance
	var currentStock int
	err := r.db.QueryRow("SELECT stock FROM products WHERE id = $1", productID).Scan(&currentStock)
	if err != nil {
		return nil, err
	}

	// Calculate balance after movement
	balanceAfter := currentStock
	switch movementType {
	case models.StockMovementReserve, models.StockMovementDeduct:
		balanceAfter = currentStock - quantity
	case models.StockMovementRelease:
		balanceAfter = currentStock + quantity
	}

	// Insert movement record
	query := `
		INSERT INTO stock_movements (product_id, order_id, movement_type, quantity, balance_after, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	movement := &models.StockMovement{
		ProductID:    productID,
		OrderID:      orderID,
		MovementType: movementType,
		Quantity:     quantity,
		BalanceAfter: balanceAfter,
		Notes:        notes,
	}

	err = r.db.QueryRow(query, productID, orderID, movementType, quantity, balanceAfter, notes).
		Scan(&movement.ID, &movement.CreatedAt)
	if err != nil {
		return nil, err
	}

	return movement, nil
}

// RecordMovementTx records a stock movement within a transaction
func (r *stockRepository) RecordMovementTx(tx *sql.Tx, productID int, orderID *int, movementType models.StockMovementType, quantity int, notes string) (*models.StockMovement, error) {
	// Get current stock balance (after any updates in this transaction)
	var currentStock int
	err := tx.QueryRow("SELECT stock FROM products WHERE id = $1", productID).Scan(&currentStock)
	if err != nil {
		return nil, err
	}

	// Balance after is the current stock (which already reflects the change)
	balanceAfter := currentStock

	// Insert movement record
	query := `
		INSERT INTO stock_movements (product_id, order_id, movement_type, quantity, balance_after, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	movement := &models.StockMovement{
		ProductID:    productID,
		OrderID:      orderID,
		MovementType: movementType,
		Quantity:     quantity,
		BalanceAfter: balanceAfter,
		Notes:        notes,
	}

	err = tx.QueryRow(query, productID, orderID, movementType, quantity, balanceAfter, notes).
		Scan(&movement.ID, &movement.CreatedAt)
	if err != nil {
		return nil, err
	}

	return movement, nil
}

// GetMovementsByProduct returns all movements for a product
func (r *stockRepository) GetMovementsByProduct(productID int) ([]models.StockMovement, error) {
	query := `
		SELECT id, product_id, order_id, movement_type, quantity, balance_after, notes, created_at
		FROM stock_movements
		WHERE product_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []models.StockMovement
	for rows.Next() {
		var m models.StockMovement
		err := rows.Scan(&m.ID, &m.ProductID, &m.OrderID, &m.MovementType, &m.Quantity, &m.BalanceAfter, &m.Notes, &m.CreatedAt)
		if err != nil {
			continue
		}
		movements = append(movements, m)
	}

	return movements, nil
}

// GetMovementsByOrder returns all movements for an order
func (r *stockRepository) GetMovementsByOrder(orderID int) ([]models.StockMovement, error) {
	query := `
		SELECT id, product_id, order_id, movement_type, quantity, balance_after, notes, created_at
		FROM stock_movements
		WHERE order_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []models.StockMovement
	for rows.Next() {
		var m models.StockMovement
		err := rows.Scan(&m.ID, &m.ProductID, &m.OrderID, &m.MovementType, &m.Quantity, &m.BalanceAfter, &m.Notes, &m.CreatedAt)
		if err != nil {
			continue
		}
		movements = append(movements, m)
	}

	return movements, nil
}

// GetCurrentStock returns the current stock for a product
func (r *stockRepository) GetCurrentStock(productID int) (int, error) {
	var stock int
	err := r.db.QueryRow("SELECT stock FROM products WHERE id = $1", productID).Scan(&stock)
	return stock, err
}
