package main

import (
	"bytes"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

var (
	re       = regexp.MustCompile(`from\s+['"](.+?)['"]`)
	ReadFile = os.ReadFile
)

func Parse(cwd string, walk func(root string, fn filepath.WalkFunc) error) (Components, error) {

	var components = make(Components)

	err := walk(cwd, func(path string, info os.FileInfo, err error) error {

		if isIgnore(path, info) {
			return nil
		}

		cmp := Component{
			Froms:    []string{},
			Path:     path,
			Children: []*Component{},
			Parents:  []*Component{},
		}

		if err := setScriptContent(path, &cmp); err != nil {
			return err
		}

		for _, match := range re.FindAllStringSubmatch(string(cmp.Script), -1) {

			dependentFilePath := ""
			filename := match[1]
			if filepath.Ext(filename) == "" {
				filename = filename + ".vue"
			}

			switch filename[0] {
			case '.':
				dependentFilePath = filepath.Join(filepath.Dir(path), filename)

			case '~', '@', '/':
				dependentFilePath = filepath.Join(cwd, filename[1:])

			default:
				dependentFilePath = filepath.Join(filepath.Dir(path), filename)
			}

			cmp.Froms = append(
				cmp.Froms,
				dependentFilePath,
			)
		}

		sort.Strings(cmp.Froms)

		components[path] = &cmp

		return nil
	})

	if err != nil {
		return nil, err
	}

	setDependencies(components)

	return components, nil

}

func isIgnore(path string, info os.FileInfo) bool {
	return info.IsDir() || filepath.Ext(path) != ".vue"
}

func setScriptContent(path string, cmp *Component) error {

	content, err := ReadFile(path)
	if err != nil {
		return err
	}

	content = bytes.Join(
		[][]byte{[]byte("<c>"), content, []byte("</c>")},
		[]byte(""),
	)

	var c = &struct {
		Script string `xml:"script"`
	}{}
	if err := xml.Unmarshal(content, c); err != nil {
		return err
	}

	cmp.Script = c.Script

	return nil
}

func setDependencies(components Components) {

	for _, component := range components.Sorted() {

		dependents := make([]*Component, 0, len(component.Froms))

		for _, path := range component.Froms {

			if dependent, ok := components[path]; ok {
				dependent.Parents = append(dependent.Parents, component)
				dependents = append(dependents, dependent)
			}
		}

		component.Children = dependents
	}
}
