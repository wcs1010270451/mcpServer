package handlers

import "McpServer/internal/models"

// DatabaseAdapter 数据库适配器，将 models.Employee 转换为 handlers.Employee
type DatabaseAdapter struct {
	db interface {
		GetEmployeeByName(name string) (*models.Employee, error)
		GetAllEmployees() ([]models.Employee, error)
	}
}

// NewDatabaseAdapter 创建数据库适配器
func NewDatabaseAdapter(db interface {
	GetEmployeeByName(name string) (*models.Employee, error)
	GetAllEmployees() ([]models.Employee, error)
}) *DatabaseAdapter {
	return &DatabaseAdapter{db: db}
}

// GetEmployeeByName 查询员工信息
func (da *DatabaseAdapter) GetEmployeeByName(name string) (*Employee, error) {
	employee, err := da.db.GetEmployeeByName(name)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, nil
	}

	// 转换为 handlers.Employee
	return &Employee{
		ID:      employee.ID,
		Name:    employee.Name,
		Address: employee.Address,
		Phone:   employee.Phone,
		Enabled: employee.Enabled,
	}, nil
}

// GetAllEmployees 获取所有员工
func (da *DatabaseAdapter) GetAllEmployees() ([]Employee, error) {
	employees, err := da.db.GetAllEmployees()
	if err != nil {
		return nil, err
	}

	// 转换为 handlers.Employee 切片
	result := make([]Employee, len(employees))
	for i, emp := range employees {
		result[i] = Employee{
			ID:      emp.ID,
			Name:    emp.Name,
			Address: emp.Address,
			Phone:   emp.Phone,
			Enabled: emp.Enabled,
		}
	}

	return result, nil
}
