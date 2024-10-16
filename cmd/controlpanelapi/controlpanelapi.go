package controlpanelapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	guuid "github.com/google/uuid"
	"github.com/gorilla/mux"

	auth "tartarus.moon.mine/internal/auth"
	k8s "tartarus.moon.mine/internal/kubernetes"
)

// Sessions is a struct that holds the sessions
type Sessions struct {
	sync.Mutex
	Sessions map[string]string `json:"sessions"`
}

// Create a new sessions struct
var sessions = Sessions{Sessions: make(map[string]string)}

// Create global variable that is the cluster.NewClient
var cluster, _ = k8s.NewClient()

// Start the api server
func Controlpanel() {
	// Create a new router
	r := mux.NewRouter()

	// serve a static pathprefix of / to /public/
	static := r.PathPrefix("/api/").Subrouter()
	static.HandleFunc("/login", login).Methods("POST")
	static.HandleFunc("/logout", logout).Methods("POST")
	static.HandleFunc("/rovers", getRovers).Methods("GET")
	static.HandleFunc("/rovers", createRover).Methods("POST")
	static.HandleFunc("/rovers/{id}", getRover).Methods("GET")
	static.HandleFunc("/rovers/{id}", deleteRover).Methods("DELETE")
	static.HandleFunc("/rovers/{id}/command", commandRover).Methods("POST")
	static.HandleFunc("/health", health).Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public/"))))

	// Log message about starting the server under Control Panel Mode
	log.Println("Starting the server in Control Panel Mode")

	// Log all requests
	static.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
	})

	// either 80 or os.Getenv(SERVER_PORT)
	port, present := os.LookupEnv("SERVER_PORT")
	if !present {
		port = "80"
	}
	
	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// login endpoint, if valid it should give the user back a cookie with a session id
func login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Println("Logging in user: ", username, password)
	err := auth.Login(username, password)
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	sessionID := guuid.New().String()
	sessions.Lock()
	sessions.Sessions[sessionID] = username
	sessions.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: sessionID,
	})

	// redirect to public/controlpanel.html
	http.Redirect(w, r, "/controlpanel.html", http.StatusFound)
}

// logout endpoint, if valid it should remove the session id from the sessions map
func logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	log.Println("Logging out user: ", sessions.Sessions[session.Value])
	sessions.Lock()
	delete(sessions.Sessions, session.Value)
	sessions.Unlock()
}

// getRovers endpoint, if valid it should return all the rovers
func getRovers(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if sessions.Sessions[session.Value] == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	rovers, _ := cluster.GetPods("rovers")

	json.NewEncoder(w).Encode(rovers)

	log.Println("Getting all rovers")
}

// createRover endpoint, if valid it should create a new rover
func createRover(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if sessions.Sessions[session.Value] == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	log.Println("Creating a new rover")

	podname, cluster_err := cluster.CreatePod("rovers", nil)
	if cluster_err != nil {
		log.Println("Error creating rover: ", err)
	}

	// set return header to include podname
	w.Header().Set("X-RoverID", podname)
	w.Header().Set("Location", "/controlpanel.html")
	log.Println("Created rover: ", podname)

	// redirect to controlpanel.html
	http.Redirect(w, r, "/controlpanel.html", http.StatusFound)
}

// getRover endpoint, if valid it should return a specific rover
func getRover(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if sessions.Sessions[session.Value] == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	params := mux.Vars(r)

	cluster.GetPod("rovers", params["id"])
}

// Command Rover, it should pull the IP address of a provided rover and access the /command endpoint on it
func commandRover(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if sessions.Sessions[session.Value] == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	params := mux.Vars(r)

	target, err := cluster.GetPod("rovers", params["id"])
	if err != nil {
		http.Redirect(w, r, "/controlpanel.html", http.StatusFound)
		return
	}

	// read the body of the request and capture the command to send to the rover
	body := new(strings.Builder)
	io.Copy(body, r.Body)

	//command := strings.NewReader(body.String())

	port, present := os.LookupEnv("CLIENT_PORT")
	if !present {
		port = "80"
	}

	// http post to target IP address:8080/command with the body contents of the command
	//_, err = http.Post("http://"+target.Status.PodIP+":8080/command", "application/json", command)
	commandresp, err := http.Post("http://"+target.Status.PodIP+":"+port+"/command", "application/json", strings.NewReader(body.String()))
	if err != nil {
		log.Println("Error commanding rover: ", err)
		http.Redirect(w, r, "/controlpanel.html", http.StatusFound)
		return
	}

	log.Println("Commanding rover: ", target.Status.PodIP)
	// log the command that's being sent in the request
	log.Println("Command: ", body.String())
	// log the results from the "Command-Output" header
	log.Println("Command-Output: ", commandresp.Header.Get("Command-Output"))

	// Add the results to the header
	w.Header().Set("X-Command-Output", commandresp.Header.Get("Command-Output"))
	w.Header().Set("Location", "/controlpanel.html")

	// respond with the podname
	w.Write([]byte(commandresp.Header.Get("Command-Output")))
}

// deleteRover endpoint, if valid it should delete a specific rover
func deleteRover(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if sessions.Sessions[session.Value] == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	params := mux.Vars(r)

	log.Println("Deleting rover: ", params["id"])

	cluster.DeletePod("rovers", params["id"])
}

// health endpoint, if valid it should return a 200 status
func health(w http.ResponseWriter, r *http.Request) {
	_, err := cluster.GetPods("rovers")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)

	log.Println("Checking health")

	w.Write([]byte("OK"))
}
