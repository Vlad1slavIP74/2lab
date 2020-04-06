package gomodule

import (
	"bytes"
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"strings"
	"testing"
)

func TestSimpleBinFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
			go_binary {
			  name: "test-out",
			  pkg: ".",
        testPkg: ".",
        srcs: ["test-src.go"],
	      vendorFirst: true
			}
		`),
		"test-src.go": nil,
	})

	ctx.RegisterModuleType("go_binary", SimpleBinFactory)

	cfg := bood.NewConfig()

	_, errs := ctx.ParseBlueprintsFiles(".", cfg)
	if len(errs) != 0 {
		t.Fatalf("Syntax errors in the test blueprint file: %s", errs)
	}

	_, errs = ctx.PrepareBuildActions(cfg)
	if len(errs) != 0 {
		t.Errorf("Unexpected errors while preparing build actions: %s", errs)
	}
	buffer := new(bytes.Buffer)
	if err := ctx.WriteBuildFile(buffer); err != nil {
		t.Errorf("Error writing ninja file: %s", err)
	} else {
		text := buffer.String()
		fmt.Println(text)
		t.Logf("Generated ninja build file:\n%s", text)
		if !strings.Contains(text, "out/bin/test-out: ") {
			t.Errorf("Generated ninja file does not have build of the test module")
		}
		if !strings.Contains(text, " test-src.go") {
			t.Errorf("Generated ninja file does not have source dependency")
		}
		if !strings.Contains(text, "build vendor: g.gomodule.vendor | go.mod") {
			t.Errorf("Generated ninja file does not have vendor build rule")
		}
	}
}
