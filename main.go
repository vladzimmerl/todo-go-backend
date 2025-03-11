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

type UserReceived struct {
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

const JsonFile = "users.json"

var simpleMutex sync.RWMutex

func main() {
	port := os.Getenv("PORT")

	mux := http.NewServeMux()
	mux.HandleFunc("/", CORS(workingTest))
	mux.HandleFunc("POST /tasks", CORS(writeUser))
	mux.HandleFunc("GET /users/{user}", CORS(getUser))

	fmt.Println("Server listening to 8080")
	handleError(http.ListenAndServe(":"+port, mux))
}

func workingTest(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, "Server Working")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	// get user
	user := r.PathValue("user")

	// fetch data
	simpleMutex.RLock()
	tasks := getUserJson(user)
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

func getUserJson(user string) []Task {
	// get data
	users := getFileData()

	// return data if it exists
	if users[user] == nil {
		return make([]Task, 0)
	} else {
		return users[user]
	}
}

func writeUser(w http.ResponseWriter, r *http.Request) {
	// get user data
	var user UserReceived
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// write to file
	simpleMutex.Lock()
	writeUserJson(user.Name, user.Tasks)
	simpleMutex.Unlock()

	// response
	w.WriteHeader(http.StatusNoContent)
}

func writeUserJson(user string, tasks []Task) {
	users := getFileData()

	// write to file
	users[user] = tasks
	out, err := json.Marshal(users)
	handleError(err)
	handleError(os.WriteFile(JsonFile, out, 0666))
}

func getFileData() map[string][]Task {
	// open file
	file, err := os.OpenFile(JsonFile, os.O_RDWR|os.O_CREATE, 0666)
	handleError(err)
	defer (func() { handleError(file.Close()) })()

	// get users
	var users map[string][]Task
	decoder := json.NewDecoder(file)
	if decoder.More() {
		handleError(decoder.Decode(&users))
	} else {
		fmt.Println("Warning: json file reads as empty")
	}

	// return
	return users
}

func handleError(err error) {
	if err != nil {
		fmt.Println("ERROR:")
		log.Fatal(err)
	}
}

func CORS(next http.HandlerFunc) http.HandlerFunc {
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
