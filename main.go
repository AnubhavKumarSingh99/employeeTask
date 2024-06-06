package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	employeeStore = make(map[int]Employee)
	storeMutex    sync.Mutex
	idCounter     = 1
)

type Employee struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Position string  `json:"position"`
	Salary   float64 `json:"salary"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	fmt.Println("starting at :8080-------")
	http.HandleFunc("/createemployee", CreateEmployeeHandler)
	http.HandleFunc("/getEmployeeById", GetEmployeeByIDHandler)
	http.HandleFunc("/deleteEmployee", DeleteEmployeeHandler)
	http.HandleFunc("/updateemployee", UpdateEmployeeHandler)
	http.HandleFunc("/listEmployee", ListEmployeeHandler)
	http.ListenAndServe(":8080", nil)
}

func CreateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	contentType := strings.ToLower(r.Header.Get("content-type"))
	if r.Method != http.MethodPost {
		sendError(w, "method not allow", http.StatusMethodNotAllowed)
	} else {
		var empDetail *Employee
		Json, err := parseRequestData(contentType, r)
		if contentType == "application/json" {
			err = Json.Decode(&empDetail)
		} else {
			err = Json.Decode(&empDetail)
		}
		if err != nil {
			fmt.Println("err", err.Error())
			return
		}
		emp := CreateEmployee(empDetail)
		json.NewEncoder(w).Encode(emp)
	}
}

func CreateEmployee(empDetail *Employee) Employee {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	employee := Employee{
		Id:       idCounter,
		Name:     empDetail.Name,
		Position: empDetail.Position,
		Salary:   empDetail.Salary,
	}
	employeeStore[idCounter] = employee
	idCounter++
	return employee
}

func GetEmployeeByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "method not allow", http.StatusMethodNotAllowed)
	} else {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		emp, err := GetEmployeeByID(id)
		if err != nil {
			sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(emp)
	}
}

func GetEmployeeByID(id int) (Employee, error) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	if employee, exists := employeeStore[id]; exists {
		return employee, nil
	}
	return Employee{}, errors.New("employee not found")
}

func DeleteEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendError(w, "method not allow", http.StatusMethodNotAllowed)
	} else {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		err := DeleteEmployee(id)
		if err != nil {
			sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode("employee delete successfully")
	}
}

func DeleteEmployee(id int) error {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	if _, exists := employeeStore[id]; exists {
		delete(employeeStore, id)
		return nil
	}
	return errors.New("employee not found")
}

func UpdateEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	contentType := strings.ToLower(r.Header.Get("content-type"))
	if r.Method != http.MethodPost {
		sendError(w, "method not allow", http.StatusMethodNotAllowed)
	} else {
		var empDetail *Employee
		Json, err := parseRequestData(contentType, r)
		if contentType == "application/json" {
			err = Json.Decode(&empDetail)
		} else {
			err = Json.Decode(&empDetail)
		}
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		emp, err := UpdateEmployee(empDetail)
		if err != nil {
			sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(emp)
	}
}

func UpdateEmployee(emp *Employee) (Employee, error) {
	storeMutex.Lock()
	defer storeMutex.Unlock()

	if employee, exists := employeeStore[emp.Id]; exists {
		employee.Name = emp.Name
		employee.Position = emp.Position
		employee.Salary = emp.Salary
		employeeStore[emp.Id] = employee
		return employee, nil
	}
	return Employee{}, errors.New("employee not found")
}

func ListEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "method not allow", http.StatusMethodNotAllowed)
	} else {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		start := (page - 1) * limit
		end := start + limit

		storeMutex.Lock()
		defer storeMutex.Unlock()

		employees := make([]Employee, 0, len(employeeStore))
		for _, emp := range employeeStore {
			employees = append(employees, emp)
		}

		if start >= len(employees) {
			json.NewEncoder(w).Encode([]Employee{})
			return
		}

		if end > len(employees) {
			end = len(employees)
		}

		json.NewEncoder(w).Encode(employees[start:end])
	}
}

func parseRequestData(contentType string, r *http.Request) (*json.Decoder, error) {
	var Json *json.Decoder
	switch contentType {
	case "application/json":
		Json = json.NewDecoder(r.Body)
	default:
		Json = json.NewDecoder(r.Body)
	}

	return Json, nil
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errResp := ErrorResponse{Message: message}
	json.NewEncoder(w).Encode(errResp)
}
