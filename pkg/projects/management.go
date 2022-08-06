package projects

var (
	defaultProjectFile    = "project.yaml"
	defaultWorkspacesFile = "workspaces.yaml"
)

type ProjectManager interface {
	GetWorkspaces() (*Workspace, error)
	SaveWorkspaces(*Workspace) error
	GetProject(id string) (*Project, error)
	ListProjects() ([]*Project, error)
	DeleteProject(id string) error
	CreateProject(projectDescription ProjectDescription) (*Project, error)
	UpdateProject(projectID string, projectDescription ProjectDescription,
		wsSummariser WorkspaceSummariser) (*Project, error)
	UpdateModel(projectID string, msg *Message) (*Message, error)
	GetModel(projectID string) (*Message, error)
	GetProjectLocation(projID string) string
	//ZeroTrust base directory
	GetBaseDir() string
}
