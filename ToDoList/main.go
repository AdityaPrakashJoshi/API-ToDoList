package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ToDo struct definition
type ToDo struct {
	TaskID            int    `json:"id"`
	TaskToBeDne       string `json:"task"`
	DescriptionOfTask string `json:"description,omitempty"`
	Priority          string `json:"priority"`
	Duedate           string `json:"dueDate"`
}

// Global slice to simulate a database or storage
var tasks []ToDo

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createTask(w, r)
	case http.MethodGet:
		readTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func createTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create One Task")
	w.Header().Set("Content-Type", "application/json")

	if r.Body == nil {
		http.Error(w, "Please send some data", http.StatusBadRequest)
		return
	}

	var task ToDo
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Bad Request: Invalid JSON data", http.StatusBadRequest)
		return
	}

	if task.TaskToBeDne == "" {
		http.Error(w, "Bad Request: Task title cannot be empty", http.StatusBadRequest)
		return
	}

	if task.Priority == "" {
		task.Priority = "Medium"
	}

	if task.Priority != "High" && task.Priority != "Medium" && task.Priority != "Low" {
		http.Error(w, "Bad Request: Invalid priority specified", http.StatusBadRequest)
		return
	}

	task.TaskID = len(tasks) + 1
	tasks = append(tasks, task)

	fmt.Printf("Received Task: %+v\n", task)

	jsonResponse, err := json.Marshal(task)
	if err != nil {
		http.Error(w, "Internal Server Error: Unable to serialize response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func readTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Read Task")
	w.Header().Set("Content-Type", "application/json")

	dateQuery := r.URL.Query().Get("date")
	if dateQuery == "" {
		http.Error(w, "Missing required query parameter 'date'", http.StatusBadRequest)
		return
	}

	parsedDate, err := time.Parse("2006-01-02", dateQuery)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid date format: %v. Use YYYY-MM-DD", err), http.StatusBadRequest)
		return
	}

	var tasksOnDate []ToDo
	for _, task := range tasks {
		taskDate, err := time.Parse("2006-01-02", task.Duedate)
		if err != nil {
			continue
		}

		if taskDate.Equal(parsedDate) {
			tasksOnDate = append(tasksOnDate, task)
		}
	}

	if len(tasksOnDate) == 0 {
		json.NewEncoder(w).Encode("No tasks found on the provided date.")
		return
	}

	if err := json.NewEncoder(w).Encode(tasksOnDate); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding tasks: %v", err), http.StatusInternalServerError)
		return
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete Task")

	// Extract task ID from URL parameters
	vars := mux.Vars(r)
	queryTaskID := vars["id"]

	// Convert task ID to integer
	num, err := strconv.Atoi(queryTaskID)
	if err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find and delete the task with the given ID
	for i, task := range tasks {
		if num == task.TaskID {
			// Remove task from tasks slice
			tasks = append(tasks[:i], tasks[i+1:]...)
			fmt.Fprintf(w, "Task with ID %d deleted successfully", num)
			return
		}
	}

	// If task with given ID not found
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Task with ID %d not found", num)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update Task")

	// Extract task ID from URL parameters
	vars := mux.Vars(r)
	queryTaskID := vars["id"]

	// Convert task ID to integer
	num, err := strconv.Atoi(queryTaskID)
	if err != nil {
		fmt.Println("Error:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find and update the task with the given ID
	found := false
	for i := range tasks {
		if tasks[i].TaskID == num {
			// Example update logic (you can modify this based on your needs)
			tasks[i].Priority = "Updated Priority"
			tasks[i].DescriptionOfTask = "Updated Description"
			// Respond with updated task details
			json.NewEncoder(w).Encode(tasks[i])
			found = true
			break
		}
	}

	// If task with given ID not found
	if !found {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Task with ID %d not found", num)
	}
}

func main() {
	// Initialize tasks (for demonstration purposes)
	tasks = []ToDo{
		{TaskID: 1, TaskToBeDne: "Buy groceries", DescriptionOfTask: "Milk, Bread, Eggs", Priority: "High", Duedate: "2024-07-01"},
		{TaskID: 2, TaskToBeDne: "Doctor's appointment", DescriptionOfTask: "Annual check-up", Priority: "Medium", Duedate: "2024-07-02"},
		{TaskID: 3, TaskToBeDne: "Workout", DescriptionOfTask: "Gym session", Priority: "Low", Duedate: "2024-07-01"},
	}

	// Initialize the router
	r := mux.NewRouter()

	// Define routes for task handling
	r.HandleFunc("/api/tasks", taskHandler).Methods(http.MethodPost, http.MethodGet)
	r.HandleFunc("/tasks/{id}", deleteTask).Methods(http.MethodDelete)
	r.HandleFunc("/tasks/{id}", updateTask).Methods(http.MethodPut)

	// Start the HTTP server
	fmt.Println("Server listening on port 9000")
	if err := http.ListenAndServe(":9000", r); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
