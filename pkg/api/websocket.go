package api

import (
	"context"
	"log"
	"strings"
	"sync"

	otm_transform "github.com/0-trust/service/pkg/otm"
	"github.com/0-trust/service/pkg/projects"
	otm "github.com/adedayo/open-threat-model/pkg"
	"github.com/gorilla/websocket"
)

var (
	longLivedSockets = make(map[string]map[string]*websocket.Conn) //projectID -> remoteAddres -> listening socket, they get removed when remote closes
	longSocLock      sync.RWMutex
)

func addLongLivedSocket(ctx context.Context, msg projects.Message, ws *websocket.Conn) {
	longSocLock.Lock()
	defer longSocLock.Unlock()

	conns := make(map[string]*websocket.Conn)
	if cc, exists := longLivedSockets[msg.ProjectID]; exists {
		conns = cc
	}
	remoteAdd := ws.RemoteAddr().String()
	conns[remoteAdd] = ws
	longLivedSockets[msg.ProjectID] = conns

	go cleanClose(ws)

	processMessage(msg, ws)

	go readLoop(ctx, ws)
}

//websocket read loop
func readLoop(_ context.Context, ws *websocket.Conn) {
	for {
		var msg projects.Message
		if err := ws.ReadJSON(&msg); err == nil {
			processMessage(msg, ws)
		} else {
			ws.Close()
			break
		}
	}
}

func processMessage(msg projects.Message, ws *websocket.Conn) {
	log.Printf("Got projects.message %v", msg)
	switch msg.Type {
	case "update_model":
		updateModel(msg, ws)
	case "process_model":
		processModel(msg, ws)
	case "get_model":
		getModel(msg, ws)
	default:
		log.Printf("Unhandles message type: %s", msg.Type)
	}
}

func getModel(msg projects.Message, ws *websocket.Conn) {
	m, _ := pm.GetModel(msg.ProjectID)
	m.Type = "update_ui"
	ws.WriteJSON(m)
}

func processModel(msg projects.Message, ws *websocket.Conn) {
	if model, err := otm.Parse(strings.NewReader(msg.ThreatModel)); err == nil {
		if g, err := otm_transform.OtmToGraphviz(model); err == nil {
			ws.WriteJSON(projects.Message{
				Type:        "graphviz",
				ProjectID:   msg.ProjectID,
				Workspace:   msg.Workspace,
				VisualModel: g,
			})
		} else {
			ws.WriteJSON(projects.Message{
				Type:      "graphviz",
				ProjectID: msg.ProjectID,
				Workspace: msg.Workspace,
				HasError:  true,
				Error:     err.Error(),
			})
		}
	} else {
		ws.WriteJSON(projects.Message{
			Type:      "graphviz",
			ProjectID: msg.ProjectID,
			Workspace: msg.Workspace,
			HasError:  true,
			Error:     err.Error(),
		})
	}
}

func updateModel(msg projects.Message, ws *websocket.Conn) {

	m, err := pm.UpdateModel(msg.ProjectID, &msg)
	if err != nil {
		m.Error = err.Error()
		m.HasError = true
	}

	ws.WriteJSON(m)
}

func cleanClose(ws *websocket.Conn) {
	ws.SetCloseHandler(socketCloseHandler(ws))
}

func socketCloseHandler(ws *websocket.Conn) func(code int, text string) error {
	return func(c int, t string) error {
		// log.Printf("Closing socket. Code: %d, Text: %s", c, t)
		longSocLock.Lock()
		defer longSocLock.Unlock()
		for projID, socks := range longLivedSockets {
			delete(socks, ws.RemoteAddr().String())
			longLivedSockets[projID] = socks
		}
		return nil
	}
}

type webSocketDiagnosticConsumer struct {
	id      string
	started bool
	buff    []string
}

func (c *webSocketDiagnosticConsumer) start(id string) {
	c.id = id
	c.started = true

}

//Stop streaming diagnostics - noop
func (c *webSocketDiagnosticConsumer) ReceiveDiagnostic() {

}

func GetListeningSocketsByProjectID(id string) []*websocket.Conn {
	projID := strings.Split(id, ":")[0]
	longSocLock.Lock()
	defer longSocLock.Unlock()

	out := []*websocket.Conn{}
	if conns, exist := longLivedSockets[projID]; exist {
		for _, c := range conns {
			out = append(out, c)
		}
	}
	return out
}
