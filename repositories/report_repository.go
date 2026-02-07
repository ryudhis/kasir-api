package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"kasir-api/models"
	"time"
)

var (
	ErrInvalidDateFormat = errors.New("format tanggal tidak valid")
	ErrInvalidDateRange  = errors.New("rentang tanggal tidak valid")
	ErrNoTransactions    = errors.New("tidak ada transaksi")
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func parseDateOrDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("%w: tanggal kosong", ErrInvalidDateFormat)
	}

	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("%w: format tanggal tidak valid (%q)", ErrInvalidDateFormat, s)
}

func formatDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}

func (repo *ReportRepository) GetDailyReport() (*models.Report, error) {
	query := `
		WITH daily_tx AS (
			SELECT id
			FROM transactions
			WHERE created_at >= CURRENT_DATE
			  AND created_at < CURRENT_DATE + INTERVAL '1 day'
		),
		top AS (
			SELECT
				td.product_name AS product_name,
				SUM(td.quantity) AS qty
			FROM transaction_details td
			JOIN daily_tx dt ON dt.id = td.transaction_id
			GROUP BY td.product_name
			ORDER BY qty DESC, td.product_name ASC
			LIMIT 1
		)
		SELECT
			COALESCE(SUM(t.total_amount), 0) AS total_revenue,
			COUNT(t.id) AS total_transactions,
			COALESCE((SELECT product_name FROM top), '') AS top_product_name,
			COALESCE((SELECT qty FROM top), 0) AS top_product_qty
		FROM transactions t
		WHERE t.created_at >= CURRENT_DATE
		  AND t.created_at < CURRENT_DATE + INTERVAL '1 day';
	`

	report := &models.Report{}
	var topName string
	var topQty int

	err := repo.db.QueryRow(query).Scan(
		&report.TotalRevenue,
		&report.TotalTransactions,
		&topName,
		&topQty,
	)
	if err != nil {
		return nil, err
	}

	if report.TotalTransactions == 0 {
		return nil, fmt.Errorf("%w: tidak ada transaksi hari ini", ErrNoTransactions)
	}

	report.TopProduct = models.TopProduct{
		ProductName:  topName,
		QuantitySold: topQty,
	}
	return report, nil
}

func (repo *ReportRepository) GetReport(startDate, endDate string) (*models.Report, error) {
	start, err := parseDateOrDateTime(startDate)
	if err != nil {
		return nil, err
	}

	end, err := parseDateOrDateTime(endDate)
	if err != nil {
		return nil, err
	}

	if !end.After(start) {
		return nil, fmt.Errorf(
			"%w: end_date (%s) harus setelah start_date (%s)",
			ErrInvalidDateRange,
			formatDateOnly(end),
			formatDateOnly(start),
		)
	}

	query := `
		WITH daily_tx AS (
			SELECT id
			FROM transactions
			WHERE created_at >= $1
			  AND created_at < $2
		),
		top AS (
			SELECT
				td.product_name AS product_name,
				SUM(td.quantity) AS qty
			FROM transaction_details td
			JOIN daily_tx dt ON dt.id = td.transaction_id
			GROUP BY td.product_name
			ORDER BY qty DESC, td.product_name ASC
			LIMIT 1
		)
		SELECT
			COALESCE(SUM(t.total_amount), 0) AS total_revenue,
			COUNT(t.id) AS total_transactions,
			COALESCE((SELECT product_name FROM top), '') AS top_product_name,
			COALESCE((SELECT qty FROM top), 0) AS top_product_qty
		FROM transactions t
		WHERE t.created_at >= $1
		  AND t.created_at < $2;
	`

	report := &models.Report{}
	var topName string
	var topQty int

	err = repo.db.QueryRow(query, start, end).Scan(
		&report.TotalRevenue,
		&report.TotalTransactions,
		&topName,
		&topQty,
	)
	if err != nil {
		return nil, err
	}

	if report.TotalTransactions == 0 {
		return nil, fmt.Errorf(
			"%w: tidak ada transaksi pada %s sampai %s",
			ErrNoTransactions,
			formatDateOnly(start),
			formatDateOnly(end),
		)
	}

	report.TopProduct = models.TopProduct{
		ProductName:  topName,
		QuantitySold: topQty,
	}
	return report, nil
}
