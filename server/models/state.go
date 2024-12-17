package models

type state string

const (
	StateCreated    = "CREATED" // TODO:
	StateFailed     = "FAILED"
	StateInProgress = "INPROGRESS"
)

type State struct {
	ID    int    `json:"id" gorm:"primaryKey"`
	Value string `json:"value"`
}

// CreateState creates state table values
func (d *DB) CreateState() error {
	if err := d.db.Create(&State{Value: StateCreated}).Error; err != nil {
		return err
	}

	if err := d.db.Create(&State{Value: StateFailed}).Error; err != nil {
		return err
	}

	if err := d.db.Create(&State{Value: StateInProgress}).Error; err != nil {
		return err
	}

	return nil
}
