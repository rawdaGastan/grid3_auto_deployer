package models

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID          int              `json:"id" gorm:"primaryKey"`
	UserID      string           `json:"user_id"  binding:"required"`
	Total       float64          `json:"total"`
	Deployments []DeploymentItem `json:"deployments" gorm:"foreignKey:invoice_id"`
	// TODO:
	Tax            float64        `json:"tax"`
	Paid           bool           `json:"paid"`
	PaymentDetails PaymentDetails `json:"payment_details" gorm:"foreignKey:invoice_id"`
	LastReminderAt time.Time      `json:"last_remainder_at"`
	CreatedAt      time.Time      `json:"created_at"`
	PaidAt         time.Time      `json:"paid_at"`
}

type DeploymentItem struct {
	ID                  int     `json:"id" gorm:"primaryKey"`
	InvoiceID           int     `json:"invoice_id"`
	DeploymentID        int     `json:"deployment_id"`
	DeploymentType      string  `json:"type"`
	DeploymentResources string  `json:"resources"`
	HasPublicIP         bool    `json:"has_public_ip"`
	PeriodInHours       float64 `json:"period"`
	Cost                float64 `json:"cost"`
}

type PaymentDetails struct {
	InvoiceID      int     `json:"invoice_id"`
	Card           float64 `json:"card"`
	Balance        float64 `json:"balance"`
	VoucherBalance float64 `json:"voucher_balance"`
}

// CreateInvoice creates new invoice
func (d *DB) CreateInvoice(invoice *Invoice) error {
	return d.db.Create(&invoice).Error
}

// GetInvoice returns an invoice by ID
func (d *DB) GetInvoice(id int) (Invoice, error) {
	var invoice Invoice
	return invoice, d.db.First(&invoice, id).Error
}

// ListUserInvoices returns all invoices of user
func (d *DB) ListUserInvoices(userID string) ([]Invoice, error) {
	var invoices []Invoice
	return invoices, d.db.Where("user_id = ?", userID).Find(&invoices).Error
}

// ListInvoices returns all invoices (admin)
func (d *DB) ListInvoices() ([]Invoice, error) {
	var invoices []Invoice
	return invoices, d.db.Find(&invoices).Error
}

// ListUnpaidInvoices returns unpaid user invoices
func (d *DB) ListUnpaidInvoices(userID string) ([]Invoice, error) {
	var invoices []Invoice
	return invoices, d.db.Order("total desc").Where("user_id = ?", userID).Where("paid = ?", false).Find(&invoices).Error
}

func (d *DB) UpdateInvoiceLastRemainderDate(id int) error {
	return d.db.Model(&Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"last_remainder_at": time.Now()}).Error
}

// PayInvoice updates paid with true and paid at field with current time in the invoice
func (d *DB) PayInvoice(id int, payment PaymentDetails) error {
	var invoice Invoice
	result := d.db.Model(&invoice).
		Where("id = ?", id).
		Update("paid", true).
		Update("payment_details", payment).
		Update("paid_at", time.Now())

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// PayUserInvoices tries to pay invoices with a given balance
func (d *DB) PayUserInvoices(userID string, balance, voucherBalance float64) (float64, float64, error) {
	// get unpaid invoices
	var invoices []Invoice
	if err := d.db.
		Order("total desc").
		Where("user_id = ?", userID).
		Where("paid = ?", false).
		Find(&invoices).Error; err != nil && err != gorm.ErrRecordNotFound {
		return 0, 0, err
	}

	for _, invoice := range invoices {
		if balance == 0 && voucherBalance == 0 {
			break
		}

		// 1. check voucher balance
		if invoice.Total <= voucherBalance {
			if err := d.PayInvoice(invoice.ID, PaymentDetails{VoucherBalance: invoice.Total}); err != nil {
				return 0, 0, err
			}
			voucherBalance -= invoice.Total
			continue
		}

		// 2. check balance
		if invoice.Total <= balance {
			if err := d.PayInvoice(invoice.ID, PaymentDetails{Balance: invoice.Total}); err != nil {
				return 0, 0, err
			}
			balance -= invoice.Total
			continue
		}

		// 3. check both (total is more than both balance and voucher balance)
		if invoice.Total <= balance+voucherBalance {
			if err := d.PayInvoice(
				invoice.ID,
				PaymentDetails{VoucherBalance: voucherBalance, Balance: (invoice.Total - voucherBalance)},
			); err != nil {
				return 0, 0, err
			}

			// use voucher first
			balance -= (invoice.Total - voucherBalance)
			voucherBalance = 0
		}
	}

	return balance, voucherBalance, nil
}

// CalcUserDebt calculates the user debt according to invoices
func (d *DB) CalcUserDebt(userID string) (float64, error) {
	var debt float64
	result := d.db.Model(&Invoice{}).
		Select("sum(total)").
		Where("user_id = ?", userID).
		Where("paid = ?", false).
		Scan(&debt)

	return debt, result.Error
}
