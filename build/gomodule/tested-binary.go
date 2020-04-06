package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/Vlad1slavIP74/2lab/build/gomodule")

	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd ${workDir} && go build -o ${output} ${pkg}",
		Description: "build go command ${pkg}",
	}, "workDir", "output", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd ${workDir} && go mod vendor",
		Description: "vendor dependencies of ${name}",
	}, "workDir", "name")

	// Ninja rule to execute go test.
	goTest = pctx.StaticRule("gotest", blueprint.RuleParams{
		Command:     "cd ${workDir} && go test -v ${testPkg} > ${testOutput}",
		Description: "test ${testPkg}",
	}, "workDir", "testOutput", "testPkg")
)

// goBinaryModuleType implements the simplest Go binary build without running tests for the target Go package.
type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		Name string
		Pkg string
		TestPkg string
		Srcs []string
		SrcsExclude []string
		VendorFirst bool
	}
}

func (tb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	output := path.Join(config.BaseOutputDir, "bin/bood", name)
	testOutput := path.Join(config.BaseOutputDir, "reports/bood", "test.txt")

	var inputs []string
	inputErrors := false
	for _, src := range tb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.SrcsExclude); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}
	if inputErrors {
		return
	}
	var testInputs []string
	inputErrors = false

	for _, src := range tb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.SrcsExclude); err == nil {
			testInputs = append(testInputs, matches...)
		} else {
			ctx.PropertyErrorf("testSrcs", "Cannot resolve files that match pattern %s", src)
			inputErrors = true
		}
	}
	if inputErrors {
		return
	}

	if tb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{output},
		Implicits:   inputs,
		Args: map[string]string{
			"workDir":    ctx.ModuleDir(),
			"output": output,
			"pkg":        tb.properties.Pkg,
		},
	})
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Initiate %s tests to Go binary", name),
			Rule:        goTest,
			Outputs:     []string{testOutput},
			Implicits:   testInputs,
			Args: map[string]string{
				"workDir":    ctx.ModuleDir(),
				"testOutput": testOutput,
				"testPkg":    tb.properties.TestPkg,
			},
		})
}

// SimpleBinFactory is a factory for go binary module type which supports Go command packages without running tests.
func SimpleBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
