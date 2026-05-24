package employee

import "github.com/pli/absensi-api/internal/auth"

// Re-export from auth package — employee data lives in auth module.
type Employee = auth.Employee
type CreateRequest = auth.CreateEmployeeRequest
