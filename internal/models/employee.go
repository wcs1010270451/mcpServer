package models

// Employee 表示 employees 表的数据模型
type Employee struct {
	ID      int64  `json:"id" db:"id"`
	Name    string `json:"name" db:"name"`
	Address string `json:"address" db:"address"`
	Phone   string `json:"phone" db:"phone"`
	Enabled bool   `json:"enabled" db:"enabled"`
}
