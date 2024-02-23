package main

import "flag"

type Option struct {
	Targets      []string
	ShowParents  bool
	ShowChildren bool
	Flat         bool
	Color        bool
}

type Targets []string

func (t *Targets) String() string {
	return ""
}

func (t *Targets) Set(value string) error {
	*t = append(*t, value)
	return nil
}

func buildOpt() Option {

	var (
		showParents  = flag.Bool("p", false, "show parents")
		showChildren = flag.Bool("c", false, "show children")
		flat         = flag.Bool("f", false, "flat")
		color        = flag.Bool("color", false, "color")
	)

	flag.Parse()

	if !*showParents && !*showChildren {
		*showChildren = true
	}

	return Option{
		Targets:      flag.Args(),
		ShowParents:  *showParents,
		ShowChildren: *showChildren,
		Flat:         *flat,
		Color:        *color,
	}
}

func (opt Option) IsShowAllTarget() bool {
	return len(opt.Targets) == 0
}
