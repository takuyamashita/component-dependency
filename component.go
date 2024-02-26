package main

import (
	"fmt"
	"io"
	"sort"
)

type Component struct {
	Script   string `xml:"script"`
	Path     string
	Froms    []string
	Children []*Component
	Parents  []*Component
}

func (c Component) isDependentOf(target *Component) bool {
	for _, child := range c.Children {
		if child == target {
			return true
		}
	}

	return false
}

type Components map[string]*Component

func (c Components) Sorted() []*Component {
	paths := make([]string, 0, len(c))
	for path := range c {
		paths = append(paths, path)
	}

	sort.Strings(paths)
	new := make([]*Component, 0, len(c))
	for _, path := range paths {
		new = append(new, c[path])
	}

	return new
}

type color string

const (
	Red      color  = "\033[31m"
	Green    color  = "\033[32m"
	colorEnd string = "\033[0m"
)

type cfg struct {
	ShowParents bool
	Flat        bool
	Color       color
	Cwd         string
}

func WriteComponets(
	w io.Writer,

	// Current component
	cur *Component,

	// How many children|parents base component has
	groupCount int,

	// Current component index on groupCount
	index int,
	depth int,

	//
	// | ← This is parasol
	//
	// This slice means that line of current component should render parasols
	//
	// ├── BBB/Dog.vue
	// │   └── AAA/BBB/Bard.vue
	// │       ├── AAA/Alice.vue
	// │       │   └── CCC/Bike.vue
	// ↑   ↑   ↑   ↑
	// T   F   T   T
	parasols []bool,
	cfg cfg,
) {

	colorStart := cfg.Color
	colorClose := colorEnd
	if cfg.Color == "" || depth != 0 {
		colorStart = ""
		colorClose = ""
	}

	indent := ""
	for _, needParasol := range parasols {
		if needParasol {
			indent += "│   "
			continue
		}
		indent += "    "
	}

	branch := "├── "
	if groupCount == index+1 {
		parasols = append(parasols, false)
		branch = "└── "
	} else {
		parasols = append(parasols, true)
	}

	if cfg.Flat {
		branch = ""
		indent = ""
	}

	w.Write([]byte(fmt.Sprintf(
		"%s%s%s%s%s\n",
		indent,
		branch,
		colorStart,
		cur.Path[len(cfg.Cwd)+1:],
		colorClose,
	)))

	if cfg.ShowParents {

		for i, dependent := range cur.Parents {
			WriteComponets(w, dependent, len(cur.Parents), i, depth+1, parasols, cfg)
		}
		return
	}

	for i, dependent := range cur.Children {
		WriteComponets(w, dependent, len(cur.Children), i, depth+1, parasols, cfg)
	}
}
