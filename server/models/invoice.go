package models

import (
	"time"

	"gorm.io/gorm"
)

type Invoice struct {
	ID          int              `json:"id" gorm:"primaryKey"`
	UserID      string           `json:"user_id" binding:"required"`
	Total       float64          `json:"total"`
	Deployments []DeploymentItem `json:"deployments" gorm:"foreignKey:invoice_id"`
	// TODO:
	Tax            float64        `json:"tax"`
	Paid           bool           `json:"paid"`
	PaymentDetails PaymentDetails `json:"payment_details" gorm:"foreignKey:invoice_id"`
	LastReminderAt time.Time      `json:"last_reminder_at"`
	CreatedAt      time.Time      `json:"created_at"`
	PaidAt         time.Time      `json:"paid_at"`
	FileData       []byte         `json:"file_data" gorm:"type:blob"`
}

type DeploymentItem struct {
	ID                  int       `json:"id" gorm:"primaryKey"`
	InvoiceID           int       `json:"invoice_id"`
	DeploymentID        int       `json:"deployment_id"`
	DeploymentName      string    `json:"deployment_name"`
	DeploymentCreatedAt time.Time `json:"deployment_created_at"`
	DeploymentType      string    `json:"type"`
	DeploymentResources string    `json:"resources"`
	HasPublicIP         bool      `json:"has_public_ip"`
	PeriodInHours       float64   `json:"period"`
	Cost                float64   `json:"cost"`
}

type PaymentDetails struct {
	ID             int     `json:"id" gorm:"primaryKey"`
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
	return d.db.Model(&Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"last_reminder_at": time.Now()}).Error
}

// PayInvoice updates paid with true and paid at field with current time in the invoice
func (d *DB) PayInvoice(id int, payment PaymentDetails) error {
	var invoice Invoice
	if err := d.db.Model(&invoice).Association("PaymentDetails").Append(&payment); err != nil {
		return err
	}

	result := d.db.Model(&invoice).
		Where("id = ?", id).
		Update("paid", true).
		Update("paid_at", time.Now())

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
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
