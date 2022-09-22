package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/0-trust/service/pkg/projects"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	routes         = mux.NewRouter()
	apiVersion     = "0.0.0"
	pm             projects.ProjectManager
	allowedOrigins = []string{
		"localhost:18273",
		"http://localhost:4200",
	}
	corsOptions = []handlers.CORSOption{
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodHead, http.MethodPost}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Accept", "Accept-Language", "Origin"}),
		handlers.AllowCredentials(),
		handlers.AllowedOriginValidator(allowedOriginValidator),
	}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return allowedOriginValidator(r.Host)
		},
	}
)

func init() {
	addRoutes()
}

func allowedOriginValidator(origin string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
	}
	passCORS := strings.Split(strings.TrimPrefix(origin, "http://"), ":")[0] == "localhost" //allow localhost independent of port
	if !passCORS {
		fmt.Printf("Host %s fails CORS.", origin)
	}
	return passCORS
}

func addRoutes() {

	routes.HandleFunc("/api/version", version).Methods(http.MethodGet)
	routes.HandleFunc("/api/workspaces", getWorkspaces).Methods(http.MethodGet)
	routes.HandleFunc("/api/projects", getProjects).Methods(http.MethodGet)
	routes.HandleFunc("/api/project/{projectID}", getProject).Methods(http.MethodGet)
	routes.HandleFunc("/api/project/model/{projectID}", getModel).Methods(http.MethodGet)
	routes.HandleFunc("/api/project/delete", deleteProject).Methods(http.MethodPost)
	routes.HandleFunc("/api/project/create", createProject).Methods(http.MethodPost)
	routes.HandleFunc("/api/project/updatemodel", updateThreatModel).Methods(http.MethodPost)
	routes.HandleFunc("/api/message", getMessageWebSocket).Methods(http.MethodGet)

}

func version(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(apiVersion)
}

func getWorkspaces(w http.ResponseWriter, _ *http.Request) {
	wss, err := pm.GetWorkspaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(wss)
}

func updateThreatModel(w http.ResponseWriter, r *http.Request) {
	var model projects.Message
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	m, err := updateTM(model)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(m)
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var projDesc projects.ProjectDescription
	if err := json.NewDecoder(r.Body).Decode(&projDesc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// log.Printf("Got Proj Desc: %#v\n", projDesc)
	proj, err := pm.CreateProject(projDesc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(proj)
}

func getProjects(w http.ResponseWriter, _ *http.Request) {
	projects, err := pm.ListProjects()
	// log.Printf("List proj: %v, %v", projects, err)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(projects)
}

func getProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projID := vars["projectID"]
	project, err := pm.GetProject(projID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(project)
}

func getModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projID := vars["projectID"]
	m, err := pm.GetModel(projID)
	m.Type = "update_ui"

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(m)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	var id struct {
		ProjectID string
	}
	if err := json.NewDecoder(r.Body).Decode(&id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := pm.DeleteProject(id.ProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(id.ProjectID)
}

func getMessageWebSocket(w http.ResponseWriter, r *http.Request) {
	var msg projects.Message

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading websocket connection %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = ws.ReadJSON(&msg)
	if err != nil {
		log.Printf("Error deserialising initial message %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	addLongLivedSocket(r.Context(), msg, ws)

}

// ServeAPI serves the zero trust modeller service on the specified port
func ServeAPI(config Config) {
	hostPort := "localhost:%d"
	if !config.Local {
		// not localhost electron app
		hostPort = ":%d"
	}

	hostPort = fmt.Sprintf(hostPort, config.ApiPort)
	log.Printf("Serving API on %s", hostPort)
	corsOptions = append(corsOptions, handlers.AllowedOrigins(allowedOrigins))
	apiVersion = config.AppVersion
	pm, _ = projects.NewDBProjectManager(config.DataPath)
	log.Fatal(http.ListenAndServe(hostPort, handlers.CORS(corsOptions...)(routes)))
}
