package api

import (
	"strings"

	"github.com/0-trust/service/pkg/projects"
)

func updateTM(msg projects.Message) (*projects.Message, error) {

	m, e := validateAndUpdateTM(&msg)
	if e != nil {
		m.Error = e.Error()
		m.HasError = true
		return m, e
	}
	m, err := pm.UpdateModel(m.ProjectID, m)
	if err != nil {
		m.Error = err.Error()
		m.HasError = true
	}
	return m, err
}

func validateAndUpdateTM(msg *projects.Message) (*projects.Message, error) {

	var err error
	vm := msg.VisualModel
	vm = strings.ReplaceAll(strings.ReplaceAll(vm, "<mxGraphModel>", ""), "</mxGraphModel>", "")
	msg.VisualModel = vm
	// fmt.Printf("msg.VisualModel: %v\n", msg.VisualModel)

	return msg, err
}
