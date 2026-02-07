package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (trx *models.Transaction, err error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("items tidak boleh kosong")
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0, len(items))

	updateProductStmt, err := tx.Prepare(`
		UPDATE products
		SET stock = stock - $1
		WHERE id = $2 AND stock >= $1
		RETURNING name, price
	`)
	if err != nil {
		return nil, err
	}
	defer updateProductStmt.Close()

	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantity tidak valid untuk product id %d", item.ProductID)
		}

		var productName string
		var productPrice int

		scanErr := updateProductStmt.QueryRow(item.Quantity, item.ProductID).Scan(&productName, &productPrice)
		if scanErr != nil {
			if scanErr == sql.ErrNoRows {
				var stock int
				e := tx.QueryRow(`SELECT stock FROM products WHERE id = $1`, item.ProductID).Scan(&stock)
				if e == sql.ErrNoRows {
					return nil, fmt.Errorf("product id %d not found", item.ProductID)
				}
				if e != nil {
					return nil, e
				}
				return nil, fmt.Errorf("stok tidak cukup untuk product id %d (stok: %d, request: %d)", item.ProductID, stock, item.Quantity)
			}
			return nil, scanErr
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	err = tx.QueryRow(`INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id`, totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	insertDetailStmt, err := tx.Prepare(`
		INSERT INTO transaction_details (transaction_id, product_id, product_name, quantity, subtotal)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return nil, err
	}
	defer insertDetailStmt.Close()

	for i := range details {
		details[i].TransactionID = transactionID

		_, err = insertDetailStmt.Exec(
			details[i].TransactionID,
			details[i].ProductID,
			details[i].ProductName,
			details[i].Quantity,
			details[i].Subtotal,
		)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}, nil
}
