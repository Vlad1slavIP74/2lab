package testcoverage

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
)

var (
	pctx = blueprint.NewPackageContext("github.com/Vlad1slavIP74/2lab/build/testcoverage")

	goTestCoverage = pctx.StaticRule("testCoverage", blueprint.RuleParams{
		Command: "go test -coverprofile=$outputCoverage && go tool cover -html=$outputCoverage -o $outputHtml",
	}, "outputCoverage", "outputHtml")
)

type testedCoverageModule struct {
	blueprint.SimpleName

	properties struct {
		Name string
		Pkg  string
		// TestPkg string
		Srcs        []string
		SrcsExclude []string
		// Deps []string
	}
}

func (tb *testedCoverageModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)

	pathToCoverageReports := path.Join(config.BaseOutputDir, "reports", fmt.Sprintf("%s.out", name))
	pathToCoverageHtml := path.Join(config.BaseOutputDir, "reports", fmt.Sprintf("%s.html", name))

	var inputs []string

	for _, src := range tb.properties.Srcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.SrcsExclude); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			return
		}
	}
	//print all dependencies to make sure they exist
	// fmt.Printf("dependencies:\n")
	//
	//   ctx.VisitDirectDeps(func(d blueprint.Module) {
	//     fmt.Printf("\t%s ", d.Name())
	//   })

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Test coverage for %s", name),
		Rule:        goTestCoverage,
		Outputs:     []string{config.BaseOutputDir},
		Implicits:   inputs,
		Args: map[string]string{
			"outputCoverage": pathToCoverageReports,
			"outputHtml":     pathToCoverageHtml,
		},
	})
}

func TestCoverageFactory() (blueprint.Module, []interface{}) {
	mType := &testedCoverageModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
