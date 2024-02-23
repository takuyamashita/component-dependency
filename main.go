package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func main() {

	if err := run(buildOpt()); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(opt Option) error {

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	components, err := Parse(cwd, filepath.Walk)
	if err != nil {
		return err
	}

	outputChildren := new(bytes.Buffer)
	outputParents := new(bytes.Buffer)

	childrenCfg := cfg{ShowParents: !opt.ShowChildren, Flat: opt.Flat, Cwd: cwd}
	parentsCfg := cfg{ShowParents: opt.ShowParents, Flat: opt.Flat, Cwd: cwd}

	if opt.Color {
		childrenCfg.Color = Green
		parentsCfg.Color = Red
	}

	for _, component := range components.Sorted() {

		if opt.IsShowAllTarget() {
			WriteComponets(outputChildren, component, 0, childrenCfg)
			WriteComponets(outputParents, component, 0, parentsCfg)
			continue
		}

		for _, target := range opt.Targets {
			if component.Path == filepath.Join(cwd, target) {
				WriteComponets(outputChildren, component, 0, childrenCfg)
				WriteComponets(outputParents, component, 0, parentsCfg)
			}
		}
	}

	if opt.ShowChildren {
		outputChildren.WriteTo(os.Stdout)
	}

	if opt.ShowParents {
		outputParents.WriteTo(os.Stdout)
	}

	return nil
}
