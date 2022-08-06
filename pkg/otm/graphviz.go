package otm_transform

import (
	"strings"
	"text/template"

	otm "github.com/adedayo/open-threat-model/pkg"
)

func OtmToGraphviz(model otm.OpenThreatModel) (g string, err error) {

	if valid, err := model.Validate(); !valid {
		return "", err
	}

	temp, err := template.New("graphviz").
		Funcs(template.FuncMap{
			"genContainers": genContainers,
			"sanitiseName":  sanitiseNameForGraphViz,
		}).Parse(tplate)
	if err != nil {
		return "", err
	}

	var buff strings.Builder
	err = temp.Execute(&buff, model)
	if err != nil {
		return
	}
	return buff.String(), nil
}

func sanitiseNameForGraphViz(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func genContainers(model otm.OpenThreatModel) map[string]*containment {
	containers := make(map[string]*containment)
	for _, comp := range model.Components {
		if comp.Parent != nil {
			id := comp.Parent.GetID()
			if cont, exists := containers[id]; exists {
				cont.LeafChildren[comp.ID] = comp.Name
				containers[id] = cont
			} else {
				t := "component"
				if comp.Parent.IsTrustZone() {
					t = "zone"
				}
				name := ""
				if n, found := model.GetNameByID(comp.Parent.GetID()); found {
					name = n
				}
				containers[id] = &containment{
					Type:           t,
					ID:             comp.Parent.GetID(),
					Name:           name,
					LeafChildren:   map[string]string{comp.ID: comp.Name},
					ParentChildren: []*containment{},
				}
			}
		}
	}

	for _, tz := range model.TrustZones {

		if _, exist := containers[tz.ID]; !exist {
			containers[tz.ID] = &containment{
				Type:           "zone",
				ID:             tz.ID,
				Name:           tz.Name,
				LeafChildren:   map[string]string{},
				ParentChildren: make([]*containment, 0),
			}
		}

		trustZone := containers[tz.ID]

		if tz.Parent != nil {
			id := tz.Parent.GetID()
			if cont, exists := containers[id]; exists {
				cont.ParentChildren = append(cont.ParentChildren, trustZone)
				containers[id] = cont
			} else {
				t := "zone"
				if tz.Parent.IsTrustZone() {
					t = "zone"
				}
				name := ""
				if n, found := model.GetNameByID(tz.Parent.GetID()); found {
					name = n
				}
				containers[id] = &containment{
					Type:           t,
					ID:             tz.Parent.GetID(),
					Name:           name,
					LeafChildren:   map[string]string{},
					ParentChildren: []*containment{trustZone},
				}
			}
		}
	}

	childContainers := []string{} //containers that are children, to be removed from the top level
	for _, cont := range containers {
		for _, id := range cont.LeafChildren {
			if cc, exists := containers[id]; exists {
				//if a child is itself a parent, add it to the ParentChildren and remove it from the leaf children map
				delete(cont.LeafChildren, id)
				childContainers = append(childContainers, id)
				cont.ParentChildren = append(cont.ParentChildren, cc)
			}
		}
	}

	for _, id := range childContainers {
		delete(containers, id)
	}

	return containers
}

type containment struct {
	Type, ID, Name string
	LeafChildren   map[string]string
	ParentChildren []*containment
}

var (
	tplate = `
	{{ define "subzone" }} 
	  	subgraph cluster_{{ sanitiseName .ID }} {
			label="{{ .Name }}"
			bgcolor=lightskyblue
			{{ range .ParentChildren }}
				{{ template "subzone" . }}
			{{ end }}
			{{range $node, $name := .LeafChildren }}
				{{ sanitiseName $node }}[label="{{ $name }}"]
			{{ end }}
	  	}
	{{end}}
	digraph G {
		/* rankdir=LR; */
		/* Containers/Zones */
		{{ range genContainers . }}
			{{ template "subzone" . }}
		{{ end }}
		/* Flows */
		{{ range .DataFlows }}
			{{ sanitiseName .Source }} -> {{ sanitiseName .Destination }}
			{{ if .Bidirectional }} {{ sanitiseName .Destination }} -> {{ sanitiseName .Source }} {{ end }}
		{{ end }}
		 /* Entities */
		 X [label="X", shape="square"]
 
		 /* Relationships */
		 G -> X[label=".63"]
		 a -> b
		 b -> e [label="ssh"]
		 e -> a
		 q -> d
		 j -> t
		 subgraph cluster_zone {
			 label="zone"
			 bgcolor=lightskyblue
 
		  b e[shape=circle]
		 }
 
		 /* Ranks */
		 { rank=same; X; };
	 }
	  
`
)
