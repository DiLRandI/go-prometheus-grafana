package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/DiLRandI/go-prometheus-grafana/app/model"
)

func generateRandomAddress() model.Address {
	streets := []string{"Main St", "Elm St", "Oak St", "Maple Ave", "Cedar Rd"}
	cities := []string{"Cityville", "Townsville", "Megatown", "Villagetown", "Suburbia"}
	states := []string{"CA", "NY", "TX", "FL", "IL"}
	return model.Address{
		Street:     streets[rand.Intn(len(streets))],
		City:       cities[rand.Intn(len(cities))],
		State:      states[rand.Intn(len(states))],
		PostalCode: fmt.Sprintf("%05d", rand.Intn(100000)),
	}
}

func generateRandomEmployee(id int) model.Employee {
	firstNames := []string{"John", "Jane", "Michael", "Emily", "David", "Sarah"}
	lastNames := []string{"Doe", "Smith", "Johnson", "Williams", "Brown", "Jones"}
	positions := []string{"Software Engineer", "HR Manager", "Sales Representative", "Product Manager", "Designer"}
	salaries := rand.Float64()*50000 + 50000

	return model.Employee{
		ID:        id,
		FirstName: firstNames[rand.Intn(len(firstNames))],
		LastName:  lastNames[rand.Intn(len(lastNames))],
		Position:  positions[rand.Intn(len(positions))],
		Salary:    salaries,
		Address:   generateRandomAddress(),
	}
}

func generateCompany() model.Company {
	numDepartments := 20
	numEmployees := 50

	var departments []model.Department
	for i := 0; i < numDepartments; i++ {
		manager := generateRandomEmployee(1)
		var employees []model.Employee
		for j := 0; j < numEmployees; j++ {
			employees = append(employees, generateRandomEmployee(j+2))
		}

		departments = append(departments, model.Department{
			Name:      fmt.Sprintf("Department %d", i+1),
			Manager:   manager,
			Employees: employees,
		})
	}

	return model.Company{
		Name:        "TechCorp",
		Address:     generateRandomAddress(),
		Departments: departments,
	}
}

func main() {
	numCompanies := 100
	var companies []model.Company

	for i := 0; i < numCompanies; i++ {
		companies = append(companies, generateCompany())
	}

	f, err := os.Create("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(companies); err != nil {
		log.Fatal(err)
	}
}
