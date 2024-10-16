// This is a restful API that interacts with k8s pods using internal/kubernetes/rover.go
// It has a create, delete, and command post function and the poster must be authenticated through internal/auth/userauth.go

package roverapi

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

// Start the api server
func Roverapi() {
	// Create a new router
	r := mux.NewRouter()

	r.HandleFunc("/command", commandPod).Methods("POST")

	// Log message about starting the server in Rover Mode
	log.Println("Starting the server in Rover Mode")

	// either 80 or os.Getenv(SERVER_PORT)
	port, present := os.LookupEnv("SERVER_PORT")
	if !present {
		port = "80"
	}
	
	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// CommandPod is a function that will run a command on a pod
func commandPod(w http.ResponseWriter, r *http.Request) {
	// Get the command from the request
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Run the command and capture the output to return to the user
	out, err := exec.Command("sh", "-c", req.Command).Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log the command that was executed
	log.Println("Command executed", "command", req.Command)
	
	// Log the output of the command as a header
	w.Header().Set("Command-Output", string(out))

	// Return
	w.Write(out)
}
