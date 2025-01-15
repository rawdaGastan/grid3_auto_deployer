// Package models for database models
package models

const (
	// VMsType deployment
	VMsType = "vms"
	// K8sType deployment
	K8sType = "k8s"
)

// Notification struct holds data of notifications
type Notification struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	UserID   string `json:"user_id"  binding:"required"`
	Msg      string `json:"msg" binding:"required"`
	Seen     bool   `json:"seen" binding:"required"`
	Notified bool   `json:"notified" binding:"required"`
	// to allow redirecting from notifications to the right pages
	// for example if the type is `vm` it will be redirected to the vm page
	Type string `json:"type" binding:"required"`
}

// ListNotifications returns a list of notifications for a user.
func (d *DB) ListNotifications(userID string) ([]Notification, error) {
	var res []Notification
	query := d.db.Where("user_id = ?", userID).Find(&res)
	return res, query.Error
}

// GetNewNotifications returns a list of new notifications for a user.
func (d *DB) GetNewNotifications(userID string) ([]Notification, error) {
	var res []Notification
	query := d.db.Where("user_id = ?", userID).Where("notified = ?", false).Find(&res)
	if query.Error != nil {
		return nil, query.Error
	}

	return res, d.db.Model(&Notification{}).Where("user_id = ?", userID).Updates(map[string]interface{}{"notified": true}).Error
}

// UpdateNotification updates seen field for notification
func (d *DB) UpdateNotification(id int, seen bool) error {
	return d.db.Model(&Notification{}).Where("id = ?", id).Updates(map[string]interface{}{"seen": seen}).Error
}

// UpdateUserNotification updates seen field for user notifications
func (d *DB) UpdateUserNotification(userID string, seen bool) error {
	return d.db.Model(&Notification{}).Where("user_id = ?", userID).Updates(map[string]interface{}{"seen": seen}).Error
}

// CreateNotification adds a new notification for a user
func (d *DB) CreateNotification(n *Notification) error {
	return d.db.Create(&n).Error
}
