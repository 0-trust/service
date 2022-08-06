package projects

import (
	otm "github.com/adedayo/open-threat-model/pkg"
)

type WorkspaceSummariser func(pm ProjectManager, workspacesToUpdate []string) (*Workspace, error)

type Workspace struct {
	Details map[string]*WorkspaceDetail `json:"details" yaml:"details"`
}

type WorkspaceDetail struct {
	Projects []*Project `json:"projects" yaml:"projects"`
}

func (ws *Workspace) RemoveProject(ps *Project, pm ProjectManager) error {
	workspace := ps.Workspace
	if ws.Details == nil {
		ws.Details = make(map[string]*WorkspaceDetail)
	}
	if w, exist := ws.Details[workspace]; exist {
		newProjects := []*Project{}
		for _, p := range w.Projects {
			if p.ID != ps.ID {
				newProjects = append(newProjects, p)
			}
		}
		w.Projects = newProjects
		ws.Details[workspace] = w
	}

	return pm.SaveWorkspaces(ws)
}

type ProjectModel struct {
	ThreatModel otm.OpenThreatModel `json:"threatModel" yaml:"threatModel"`
	VisualModel string              `json:"visualModel"`
}

// type ProjectDescription struct {
// 	Name      string `json:"name" yaml:"name"`
// 	Workspace string `json:"workspace" yaml:"workspace"`
// }

type ProjectDescription struct {
	Name         string            `json:"name" yaml:"name"`
	Workspace    string            `json:"workspace"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	Owner        string            `yaml:"owner" json:"owner"`
	OwnerContact string            `yaml:"ownerContact" json:"ownerContact"`
	Attributes   map[string]string `yaml:"attributes,omitempty" json:"attributes,omitempty"`
}

// func (pd ProjectDescription) MarshalJSON() ([]byte, error) {
// 	tm, err := yaml.Marshal(pd.ThreatModel)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	return json.Marshal(WireProjectDescription{
// 		Name:         pd.Name,
// 		Workspace:    pd.Workspace,
// 		ThreatModel:  string(tm),
// 		Description:  pd.ThreatModel.Project.Description,
// 		Owner:        pd.ThreatModel.Project.Owner,
// 		OwnerContact: pd.ThreatModel.Project.OwnerContact,
// 		Attributes:   pd.ThreatModel.Project.Attributes,
// 	})
// }

// func (pd *ProjectDescription) UnmarshalJSON(data []byte) error {
// 	var wpd WireProjectDescription
// 	err := json.Unmarshal(data, &wpd)
// 	if err != nil {
// 		return err
// 	}
// 	pd.Name = wpd.Name
// 	pd.Workspace = wpd.Workspace
// 	model, err := otm.Parse(strings.NewReader(wpd.ThreatModel))
// 	if err != nil {
// 		return err
// 	}
// 	pd.ThreatModel = model
// 	return nil
// }

type Project struct {
	ID string `json:"id" yaml:"id"`
	ProjectDescription
}

type Message struct {
	Type        string `json:"type"` //indicates type of instruction
	ProjectID   string `json:"projectID"`
	Workspace   string `json:"workspace"`
	ThreatModel string `json:"threatModel"`
	VisualModel string `json:"visualModel"`
	HasError    bool   `json:"hasError"`
	Error       string `json:"error"`
}

type Model struct {
	ThreatModel     string `json:"threatModel"`
	VisualModel     string `json:"visualModel"`
	VisualIsUpdated bool   `json:"visualIsUpdated"`
	ThreatIsUpdated bool   `json:"threatIsUpdated"`
}
