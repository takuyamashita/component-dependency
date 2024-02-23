package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
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

func WriteComponets(w io.Writer, cur *Component, depth int, cfg cfg) {

	colorStart := cfg.Color
	colorClose := colorEnd
	if cfg.Color == "" || depth != 0 {
		colorStart = ""
		colorClose = ""
	}

	indent := strings.Repeat("  ", depth)
	if cfg.Flat {
		indent = ""
	}

	w.Write([]byte(fmt.Sprintf("%s%s%s%s\n", colorStart, indent, cur.Path[len(cfg.Cwd)+1:], colorClose)))

	if cfg.ShowParents {

		for _, dependent := range cur.Parents {
			WriteComponets(w, dependent, depth+1, cfg)
		}
		return
	}

	for _, dependent := range cur.Children {
		WriteComponets(w, dependent, depth+1, cfg)
	}
}
