package model

type Address struct {
	Street     string
	City       string
	State      string
	PostalCode string
}

type Employee struct {
	ID        int
	FirstName string
	LastName  string
	Position  string
	Salary    float64
	Address   Address
}

type Department struct {
	Name      string
	Manager   Employee
	Employees []Employee
}

type Company struct {
	Name        string
	Address     Address
	Departments []Department
}
