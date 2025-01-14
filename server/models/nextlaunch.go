package models

import "time"

// NextLaunch struct for next launch revealing
type NextLaunch struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Launched  bool      `json:"launched"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateNextLaunch updates the launched state of NextLaunch
func (d *DB) UpdateNextLaunch(on bool) error {
	return d.db.Model(&NextLaunch{}).Where("launched = ?", !on).Updates(map[string]interface{}{"launched": on, "updated_at": time.Now()}).Error
}

// GetNextLaunch queries on NextLaunch in db
func (d *DB) GetNextLaunch() (NextLaunch, error) {
	var res NextLaunch
	query := d.db.First(&res)
	return res, query.Error
}
