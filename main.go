package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Task struct {
	Name      string `json:"name"`
	IsChecked bool   `json:"isChecked"`
}

type ReceivedUser struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

const JsonFile = "users.json"

var simpleMutex sync.RWMutex

// ========== Main ==========

func main() {
	port := os.Getenv("PORT")

	mux := http.NewServeMux()
	mux.HandleFunc("/", Cors(IsWorkingTest))
	mux.HandleFunc("POST /tasks", Cors(WriteUser))
	mux.HandleFunc("GET /users/{user}", Cors(GetUser))

	fmt.Println("Server listening to " + port)
	HandleError(http.ListenAndServe(":"+port, mux))
}

// ========== Request Handlers ==========

// IsWorkingTest displays api status message
func IsWorkingTest(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "Server Working")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// GetUser handles GET rq
func GetUser(w http.ResponseWriter, r *http.Request) {
	// get user
	user := r.PathValue("user")

	// fetch data
	simpleMutex.RLock()
	tasks := GetUserJson(user)
	simpleMutex.RUnlock()

	// write back request
	w.Header().Set("Content-Type", "application/json")
	jsonTasks, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonTasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// WriteUser handles POST rq
func WriteUser(w http.ResponseWriter, r *http.Request) {
	// get user data
	var user ReceivedUser
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// write to file
	simpleMutex.Lock()
	WriteUserJson(user.Name, user.Tasks)
	simpleMutex.Unlock()

	// response
	w.WriteHeader(http.StatusNoContent)
}

// Cors middleware to allow cross-origin access
func Cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "https://todo-typescript-frontend.pages.dev" || origin == "https://todo.vladzimmerl.com" {
			w.Header().Add("Access-Control-Allow-Origin", origin)
		}
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		next(w, r)
	}
}

// ========== JSON Access ==========

// GetUserJson returns user from file
func GetUserJson(user string) []Task {
	// get data
	users := FileData()

	// return data if it exists
	if users[user] == nil {
		return make([]Task, 0)
	} else {
		return users[user]
	}
}

// WriteUserJson writes data to disk
func WriteUserJson(user string, tasks []Task) {
	users := FileData()

	// write to file
	users[user] = tasks
	out, err := json.Marshal(users)
	HandleError(err)
	HandleError(os.WriteFile(JsonFile, out, 0666))
}

// FileData returns stored data
func FileData() map[string][]Task {
	// open file
	file, err := os.OpenFile(JsonFile, os.O_RDWR|os.O_CREATE, 0666)
	HandleError(err)
	defer (func() { HandleError(file.Close()) })()

	// get users
	var users map[string][]Task
	decoder := json.NewDecoder(file)
	if decoder.More() {
		HandleError(decoder.Decode(&users))
	} else {
		fmt.Println("Warning: json file reads as empty")
	}

	// return
	return users
}

// HandleError logs internal errors
func HandleError(err error) {
	if err != nil {
		fmt.Println("ERROR:")
		log.Fatal(err)
	}
}
