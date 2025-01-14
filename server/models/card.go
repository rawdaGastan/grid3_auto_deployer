package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Card struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	UserID          string `json:"user_id"  binding:"required"`
	PaymentMethodID string `json:"payment_method_id" gorm:"unique" binding:"required"`
	CustomerID      string `json:"customer_id" binding:"required"`
	Fingerprint     string `json:"fingerprint" gorm:"unique" binding:"required"`
	CardType        string `json:"card_type" binding:"required"`
	ExpMonth        int64  `json:"exp_month"`
	ExpYear         int64  `json:"exp_year"`
	Last4           string `json:"last_4"`
	Brand           string `json:"brand"`
}

// AddCard adds a new card
func (d *DB) AddCard(c *Card) error {
	result := d.db.Create(&c)
	return result.Error
}

// GetCard gets a user card using ID
func (d *DB) GetCard(id int) (Card, error) {
	var res Card
	return res, d.db.First(&res, &id).Error
}

// GetCardByPaymentMethod gets a user card using stripe payment method ID
func (d *DB) GetCardByPaymentMethod(paymentMethodID string) (Card, error) {
	var res Card
	return res, d.db.First(&res, "payment_method_id = ?", paymentMethodID).Error
}

// IsCardUnique gets checks if the entered card is not a duplicate
func (d *DB) IsCardUnique(fingerprint string) (bool, error) {
	var res []Card
	err := d.db.Find(&res, "fingerprint = ?", fingerprint).Error
	if err == gorm.ErrRecordNotFound || len(res) == 0 {
		return true, nil
	}

	return false, err
}

// GetUserCards gets user cards
func (d *DB) GetUserCards(userID string) ([]Card, error) {
	var res []Card
	return res, d.db.Find(&res, "user_id = ?", userID).Error
}

// DeleteCard deletes card by its id
func (d *DB) DeleteCard(id int) error {
	var card Card
	return d.db.Delete(&card, id).Error
}

// DeleteAllCards deletes all cards of user
func (d *DB) DeleteAllCards(userID string) error {
	var cards []Card
	return d.db.Clauses(clause.Returning{}).Where("user_id = ?", userID).Delete(&cards).Error
}
