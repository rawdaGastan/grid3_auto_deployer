// Balance models for database models
package models

// Balance struct for user balance
type Balance struct {
	ID                    int    `json:"id" gorm:"primaryKey"`
	UserID                string `json:"user_id"`
	BalanceInUSD          uint64 `json:"balance_in_usd"`
	Leftover              uint64 `json:"leftover"`
	SmallVMs              int    `json:"small_vms" validate:"nonzero"`
	SmallVMsWithPublicIP  int    `json:"small_vms_with_public_ip" validate:"nonzero"`
	MediumVMs             int    `json:"medium_vms" validate:"nonzero"`
	MediumVMsWithPublicIP int    `json:"medium_vms_with_public_ip" validate:"nonzero"`
	LargeVMs              int    `json:"large_vms" validate:"nonzero"`
	LargeVMsWithPublicIP  int    `json:"large_vms_with_public_ip" validate:"nonzero"`
}

// CreateBalance creates new balance
func (d *DB) CreateBalance(b *Balance) error {
	return d.db.Create(&b).Error
}

// GetBalance return balance by its id
func (d *DB) GetBalance(id int) (Balance, error) {
	var pkg Balance
	query := d.db.First(&pkg, id)
	return pkg, query.Error
}

// GetBalanceByUserID return balance by its user ID
func (d *DB) GetBalanceByUserID(userID string) (Balance, error) {
	var pkg Balance
	query := d.db.First(&pkg, "user_id = ?", userID)
	return pkg, query.Error
}

// UpdateBalanceQuota updates quota
func (d *DB) UpdateBalanceQuota(userID string, b Balance) error {
	return d.db.Model(&Balance{}).Where("user_id = ?", userID).Updates(b).Error
}

// UpdateBalance updates balance
func (d *DB) UpdateBalance(b Balance) error {
	result := d.db.Model(&Balance{}).Where("id = ?", b.ID).Updates(b)
	return result.Error
}
