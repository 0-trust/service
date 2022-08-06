package projects

import (
	"log"
)

var (
	nothing = struct{}{}
)

func SimpleWorkspaceSummariser(pm ProjectManager, workspacesToUpdate []string) (*Workspace, error) {
	wspaces, err := pm.GetWorkspaces()
	if err != nil {
		log.Printf("SimpleWorkspaceSummariser: %v", err)
		return nil, err
	}
	if len(workspacesToUpdate) > 0 {
		//TODO
	}
	return wspaces, nil
}
