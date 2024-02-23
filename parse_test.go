package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type file struct {
	name  string
	isDir bool
}

func (f file) Name() string {

	return f.name
} // base name of the file
func (f file) Size() int64 {
	return 0
} // length in bytes for regular files; system-dependent for others
func (f file) Mode() os.FileMode {

	return 0
} // file mode bits
func (f file) ModTime() time.Time {
	return time.Time{}
} // modification time
func (f file) IsDir() bool {
	return f.isDir
} // abbreviation for Mode().IsDir()
func (f file) Sys() any {
	return nil
} // underlying data source (can return nil)

func newFile(name string, isDir bool) file {
	return file{name: name, isDir: isDir}
}

/*
├── AAA
│   ├── Alice.vue
│   ├── BBB
│   │   ├── Bard.vue
│   │   ├── Cat.vue
│   │   └── Dog.vue
│   ├── Bob.vue
│   └── Charile.vue
├── BBB
│   ├── Bard.vue
│   ├── Cat.vue
│   └── Dog.vue
├── CCC
│   ├── Bike.vue
│   ├── Car.vue
│   └── Train.vue
*/

var (
	dir = []struct {
		path string
		info os.FileInfo
	}{
		{"AAA", newFile("AAA", true)},
		{"BBB", newFile("BBB", true)},
		{"CCC", newFile("CCC", true)},
		{"AAA/Alice.vue", newFile("Alice.vue", false)},
		{"AAA/Bob.vue", newFile("Bob.vue", false)},
		{"AAA/Charile.vue", newFile("Charile.vue", false)},
		{"AAA/BBB", newFile("BBB", true)},
		{"AAA/BBB/Bard.vue", newFile("Bard.vue", false)},
		{"AAA/BBB/Cat.vue", newFile("Cat.vue", false)},
		{"AAA/BBB/Dog.vue", newFile("Dog.vue", false)},
		{"BBB/Bard.vue", newFile("Bard.vue", false)},
		{"BBB/Cat.vue", newFile("Cat.vue", false)},
		{"BBB/Dog.vue", newFile("Dog.vue", false)},
		{"CCC/Bike.vue", newFile("Bike.vue", false)},
		{"CCC/Car.vue", newFile("Car.vue", false)},
		{"CCC/Train.vue", newFile("Train.vue", false)},
	}

	fileContent = map[string]string{
		"AAA/Alice.vue":    "<script>import Bard from './BBB/Bard'</script>",
		"AAA/Bob.vue":      "<script>import Cat from './BBB/Cat'</script>",
		"AAA/Charile.vue":  "<script>import Dog from './BBB/Dog'</script>",
		"AAA/BBB/Bard.vue": "<script>import Alice from '../Alice'</script>",
		"AAA/BBB/Cat.vue":  "<script>import Bob from '../Bob'</script>",
		"AAA/BBB/Dog.vue":  "<script>import Charile from '../Charile'</script>",
		"BBB/Bard.vue":     "<script>import Alice from '../AAA/Alice'</script>",
		"BBB/Cat.vue":      "<script>import Bob from '../AAA/Bob'</script>",
		"BBB/Dog.vue":      "<script>import Charile from '../AAA/Charile'</script>",
		"CCC/Bike.vue":     "<script>import Car from './Car'</script>",
		"CCC/Car.vue":      "<script>import Train from './Train'</script>",
		"CCC/Train.vue":    "<script>import Bike from './Bike'</script>",
	}
)

func TestParse(t *testing.T) {

	ReadFile = func(name string) ([]byte, error) {
		return []byte(fileContent[name]), nil
	}

	components, err := Parse("/", func(root string, fn filepath.WalkFunc) error {
		for _, d := range dir {
			if err := fn(d.path, d.info, nil); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		t.Error(err)
	}

	if len(components) != 12 {
		t.Errorf("unexpected length: %d", len(components))
	}
}

func TestSetScriptContent(t *testing.T) {

	type arg struct {
		content []byte
	}

	// import defaultExport from "module-name";
	// import * as name from "module-name";
	// import { export1 } from "module-name";
	// import { export1 as alias1 } from "module-name";
	// import { default as alias } from "module-name";
	// import { export1, export2 } from "module-name";
	// import { export1, export2 as alias2, /* … */ } from "module-name";
	// import { "string name" as alias } from "module-name";
	// import defaultExport, { export1, /* … */ } from "module-name";
	// import defaultExport, * as name from "module-name";

	tests := []struct {
		name string
		arg  arg
		want string
	}{
		{
			name: "Default Export",
			arg: arg{
				content: []byte(`<script>import defaultExport from "module-name";</script>`),
			},
			want: `import defaultExport from "module-name";`,
		},
		{
			name: "Namespace Import",
			arg: arg{
				content: []byte(`<script>import * as name from "module-name";</script>`),
			},
			want: `import * as name from "module-name";`,
		},
		{
			name: "Named Import",
			arg: arg{
				content: []byte(`<script>import { export1 } from "module-name";</script>`),
			},
			want: `import { export1 } from "module-name";`,
		},
		{
			name: "Named Import with Alias",
			arg: arg{
				content: []byte(`<script>import { export1 as alias1 } from "module-name";</script>`),
			},
			want: `import { export1 as alias1 } from "module-name";`,
		},
		{
			name: "Default Import with Alias",
			arg: arg{
				content: []byte(`<script>import { default as alias } from "module-name";</script>`),
			},
			want: `import { default as alias } from "module-name";`,
		},
		{
			name: "Multiple Named Import",
			arg: arg{
				content: []byte(`<script>import { export1, export2 } from "module-name";</script>`),
			},
			want: `import { export1, export2 } from "module-name";`,
		},
		{
			name: "Multiple Named Import with Alias",
			arg: arg{
				content: []byte(`<script>import { export1, export2 as alias2 } from "module-name";</script>`),
			},
			want: `import { export1, export2 as alias2 } from "module-name";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cmp := &Component{}

			if err := setScriptContent(tt.arg.content, cmp); err != nil {
				t.Error(err)
			}

			if tt.want != cmp.Script {
				t.Errorf("unexpected script: %s", cmp.Script)
			}
		})
	}
}

func TestSetDependencies(t *testing.T) {

	components := Components{
		"/AAA/Alice.vue": &Component{
			Froms: []string{
				"/AAA/BBB/Bard.vue",
				"/BBB/Cat.vue",
			},
			Path: "/AAA/Alice.vue",
		},
		"/AAA/Bob.vue": &Component{
			Froms: []string{},
			Path:  "/AAA/Bob.vue",
		},
		"/AAA/Charile.vue": &Component{
			Froms: []string{},
			Path:  "/AAA/Charile.vue",
		},

		"/AAA/BBB/Bard.vue": &Component{
			Froms: []string{
				"/BBB/Bard.vue",
				"/BBB/Cat.vue",
				"/BBB/Dog.vue",
			},
			Path: "/AAA/BBB/Bard.vue",
		},
		"/AAA/BBB/Cat.vue": &Component{
			Froms: []string{},
			Path:  "/AAA/BBB/Cat.vue",
		},
		"/AAA/BBB/Dog.vue": &Component{
			Froms: []string{},
			Path:  "/AAA/BBB/Dog.vue",
		},

		"/BBB/Bard.vue": &Component{
			Froms: []string{},
			Path:  "/BBB/Bard.vue",
		},
		"/BBB/Cat.vue": &Component{
			Froms: []string{},
			Path:  "/BBB/Cat.vue",
		},
		"/BBB/Dog.vue": &Component{
			Froms: []string{},
			Path:  "/BBB/Dog.vue",
		},
		"/CCC/Bike.vue": &Component{
			Froms: []string{
				"/AAA/Alice.vue",
				"/AAA/BBB/Bard.vue",
			},
			Path: "/CCC/Bike.vue",
		},
		"/CCC/Car.vue": &Component{
			Froms: []string{},
			Path:  "/CCC/Car.vue",
		},
		"/CCC/Train.vue": &Component{
			Froms: []string{},
			Path:  "/CCC/Train.vue",
		},
	}

	setDependencies(components)

	type want struct {
		children []string
		parents  []string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "/AAA/Alice.vue",
			want: want{
				children: []string{
					"/AAA/BBB/Bard.vue",
					"/BBB/Cat.vue",
				},
				parents: []string{"/CCC/Bike.vue"},
			},
		},
		{
			name: "/AAA/Bob.vue",
			want: want{
				children: []string{},
				parents:  []string{},
			},
		},
		{
			name: "/AAA/Charile.vue",
			want: want{
				children: []string{},
				parents:  []string{},
			},
		},
		{
			name: "/AAA/BBB/Bard.vue",
			want: want{
				children: []string{"/BBB/Bard.vue", "/BBB/Cat.vue", "/BBB/Dog.vue"},
				parents:  []string{"/AAA/Alice.vue", "/CCC/Bike.vue"},
			},
		},
		{
			name: "/AAA/BBB/Cat.vue",
			want: want{

				children: []string{},
				parents:  []string{},
			},
		},
		{
			name: "/AAA/BBB/Dog.vue",
			want: want{

				children: []string{},
				parents:  []string{},
			},
		},
		{
			name: "/BBB/Bard.vue",
			want: want{
				children: []string{},
				parents: []string{
					"/AAA/BBB/Bard.vue",
				},
			},
		},
		{
			name: "/BBB/Cat.vue",
			want: want{
				children: []string{},
				parents: []string{
					"/AAA/Alice.vue",
					"/AAA/BBB/Bard.vue",
				},
			},
		},
		{
			name: "/BBB/Dog.vue",
			want: want{
				children: []string{},
				parents: []string{
					"/AAA/BBB/Bard.vue",
				},
			},
		},
		{
			name: "/CCC/Bike.vue",
			want: want{
				children: []string{
					"/AAA/Alice.vue",
					"/AAA/BBB/Bard.vue",
				},
				parents: []string{},
			},
		},
		{
			name: "/CCC/Car.vue",
			want: want{
				children: []string{},
				parents:  []string{},
			},
		},
		{
			name: "/CCC/Train.vue",
			want: want{
				children: []string{},
				parents:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cmp := components[tt.name]

			if len(cmp.Children) != len(tt.want.children) {
				t.Fatalf("unexpected length: %d", len(cmp.Children))
			}

			for i, child := range cmp.Children {
				if child.Path != tt.want.children[i] {
					t.Errorf("unexpected child: %s", child.Path)
				}
			}

			if len(cmp.Parents) != len(tt.want.parents) {
				t.Fatalf("unexpected length: %d", len(cmp.Parents))
			}

			for i, parent := range cmp.Parents {

				if parent.Path != tt.want.parents[i] {
					t.Errorf("unexpected parent: %s", parent.Path)
				}
			}

		})
	}

}
