package projects

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/0-trust/service/pkg/util"
	"github.com/dgraph-io/badger/v3"
)

var (
	projectFile          = "project-summary.yaml"
	defaultCodeDirPrefix = "code"
)

func NewDBProjectManager(ztBaseDir string) (ProjectManager, error) {

	pm := dbProjectManager{
		baseDir:          ztBaseDir,
		projectsLocation: path.Join(ztBaseDir, "projects_db"),
		projectTable:     "proj_",
		workspaceTable:   "works_",
		modelTable:       "model_",
	}

	//attempt to create the project location if it doesn't exist
	os.MkdirAll(pm.projectsLocation, 0755)

	//attempt to manage memory by setting WithNum...
	opts := badger.DefaultOptions(pm.projectsLocation) //.WithNumMemtables(1).WithNumLevelZeroTables(1).WithNumLevelZeroTablesStall(5)

	//clean up lock on the DB if previous crash
	lockFile := path.Join(opts.Dir, "LOCK")
	_ = os.Remove(lockFile)

	db, err := badger.Open(opts)
	if err != nil {
		return pm, err
	}
	pm.db = db

	return pm, nil
}

type dbProjectManager struct {
	baseDir, projectsLocation    string
	db                           *badger.DB
	projectTable, workspaceTable string
	modelTable                   string
}

// GetModel implements ProjectManager
func (pm dbProjectManager) GetModel(projectID string) (*Message, error) {
	var msg Message
	err := pm.db.View(func(txn *badger.Txn) error {
		item, e := txn.Get(toKey(pm.modelTable, projectID))
		if e == nil {
			return item.Value(func(val []byte) error {
				return json.Unmarshal(val, &msg)
			})
		}
		return e
	})

	if err != nil {
		msg.HasError = true
		msg.Error = err.Error()
	}

	log.Printf("Returning %v", msg)

	return &msg, err
}

// UpdateModel implements ProjectManager
func (pm dbProjectManager) UpdateModel(projectID string, msg *Message) (*Message, error) {

	model := Model{
		ThreatModel:     msg.ThreatModel,
		VisualModel:     stripMXGraph(msg.VisualModel),
		VisualIsUpdated: strings.TrimSpace(msg.VisualModel) != "",
		ThreatIsUpdated: strings.TrimSpace(msg.ThreatModel) != "",
	}

	data, err := json.Marshal(model)

	if err != nil {
		return msg, err
	}

	err = pm.db.Update(func(txn *badger.Txn) error {
		return txn.Set(toKey(pm.modelTable, projectID), data)
	})

	return msg, nil
}

func stripMXGraph(s string) string {
	return strings.Replace(strings.Replace(s, "<mxGraphModel>", "", 1), "</mxGraphModel>", "", 1)
}

// DeleteProject implements ProjectManager
func (pm dbProjectManager) DeleteProject(id string) error {
	proj, err := pm.GetProject(id)
	if err != nil {
		return err
	}

	//delete project
	pm.deleteProject(id)
	//remove it from workspaces
	if ws, err := pm.GetWorkspaces(); err == nil {
		ws.RemoveProject(proj, pm)
	}

	return nil
}

func (pm dbProjectManager) Close() error {
	if pm.db != nil {
		return pm.db.Close()
	}
	return errors.New("Attempting to close uninitialised DB")
}

// CreateProject implements ProjectManager
func (pm dbProjectManager) CreateProject(projectDescription ProjectDescription) (*Project, error) {

	project := &Project{
		ID:                 util.NewRandomUUID().String(),
		ProjectDescription: projectDescription,
	}

	data, err := json.Marshal(project)

	if err != nil {
		return project, err
	}

	err = pm.db.Update(func(txn *badger.Txn) error {
		return txn.Set(pm.toProjectKey(project.ID), data)
	})

	if err == nil {
		pm.UpdateModel(project.ID, &Message{})
	}
	return project, err
}

func (pm dbProjectManager) toProjectKey(projID string) []byte {
	return toKey(pm.projectTable, projID)
}

// GetBaseDir implements ProjectManager
func (pm dbProjectManager) GetBaseDir() string {
	return pm.baseDir
}

// GetProjectLocation implements ProjectManager
func (pm dbProjectManager) GetProjectLocation(projID string) string {
	return path.Join(pm.projectsLocation, projID)
}

// GetProject implements ProjectManager
func (pm dbProjectManager) GetProject(projectID string) (*Project, error) {
	var pSum Project
	err := pm.db.View(func(txn *badger.Txn) error {
		item, e := txn.Get(pm.toProjectKey(projectID))
		if e == nil {
			return item.Value(func(val []byte) error {
				return json.Unmarshal(val, &pSum)
			})
		}
		return e
	})
	return &pSum, err
}

func (pm dbProjectManager) deleteProject(projectID string) error {
	return pm.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(pm.toProjectKey(projectID))
	})
}

// GetWorkspaces implements ProjectManager
func (pm dbProjectManager) GetWorkspaces() (*Workspace, error) {
	wss := Workspace{
		Details: make(map[string]*WorkspaceDetail),
	}
	err := pm.db.View(func(txn *badger.Txn) error {
		item, rerr := txn.Get(toKey(pm.workspaceTable))
		if rerr != nil {
			return rerr
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &wss)
		})
	})

	if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
		//create a new workspace, if it didn't exist
		err = pm.SaveWorkspaces(&wss)
	}

	return &wss, err
}

// ListProjects implements ProjectManager
func (pm dbProjectManager) ListProjects() ([]*Project, error) {
	pSums := []*Project{}
	pm.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte(pm.projectTable)

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			item.Value(func(val []byte) error {
				var pSum Project
				err := json.Unmarshal(val, &pSum)
				if err == nil {
					pSums = append(pSums, &pSum)
				}
				return err
			})
		}
		return nil
	})
	sorted := make(ProjectSlice, 0, len(pSums))
	sorted = append(sorted, pSums...)
	sort.Sort(sorted)

	return sorted, nil
}

// SaveProject implements ProjectManager
func (pm dbProjectManager) SaveProject(proj *Project) error {
	return pm.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(proj)
		if err != nil {
			return err
		}
		return txn.Set(toKey(pm.projectTable, proj.ID), data)
	})
}

// SaveWorkspaces implements ProjectManager
func (pm dbProjectManager) SaveWorkspaces(ws *Workspace) error {
	return pm.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(ws)
		if err != nil {
			return err
		}
		return txn.Set(toKey(pm.workspaceTable), data)
	})
}

// UpdateProject implements ProjectManager
func (pm dbProjectManager) UpdateProject(projectID string, projectDescription ProjectDescription, wsSummariser WorkspaceSummariser) (*Project, error) {
	proj, err := pm.GetProject(projectID)

	if err != nil {
		return nil, err
	}
	if proj.ID == projectID {
		//found project, update
		proj.Name = projectDescription.Name
		wspaces := []string{proj.Workspace}
		wsChange := false
		if proj.Workspace != projectDescription.Workspace {
			//project workspace changing
			wsChange = true
			wspaces = append(wspaces, projectDescription.Workspace)
			proj.Workspace = projectDescription.Workspace
		}

		if wsChange && wsSummariser != nil {
			wss, err := wsSummariser(pm, wspaces)
			if err == nil {
				go pm.SaveWorkspaces(wss)
			} else {
				log.Printf("UpdateProject: %v", err)
			}
		}
		return proj, pm.SaveProject(proj)
	}
	//project not found, create one with a new ID
	return pm.CreateProject(projectDescription)

}

func toTableKey(prefix, projectID, scanID string) []byte {
	return []byte(fmt.Sprintf("%s%s%s", prefix, projectID, scanID))
}

func toKey(keys ...string) []byte {
	return []byte(strings.Join(keys, ""))
}

type ProjectSlice []*Project

func (t ProjectSlice) Len() int {

	return len(t)
}

func (t ProjectSlice) Less(i, j int) bool {
	return t[i].Name < (t[j].Name)
}

func (t ProjectSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
