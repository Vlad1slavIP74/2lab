package testcoverage

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	//"github.com/Vlad1slavIP74/2lab/build/gomodule"
	"path"
)

var (
	pctx    = blueprint.NewPackageContext("github.com/Vlad1slavIP74/2lab/build/testcoverage")
	goBuild = pctx.StaticRule("prebuild", blueprint.RuleParams{
		Command: "mkdir -p $fileName",
	}, "fileName")
	goTestCoverage = pctx.StaticRule("testCoverage", blueprint.RuleParams{
		Command: "cd $workDir && go test -coverprofile $outputCoverage && go tool cover -html=$outputCoverage -o $outputHtml",
	}, "workDir", "outputCoverage", "outputHtml")
)

type testedCoverageModule struct {
	blueprint.SimpleName

	properties struct {
		Name        string
		Pkg         string
		Deps        []string
		Srcs        []string
		SrcsExclude []string
		Binary      []string
	}
}

func (tcm *testedCoverageModule) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return tcm.properties.Binary
}

func (tcm *testedCoverageModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)

	pathToCoverageReports := path.Join("..", config.BaseOutputDir, "reports/coverage", fmt.Sprintf("%s.out", name))
	pathToCoverageHtml := path.Join("..", config.BaseOutputDir, "reports/coverage", fmt.Sprintf("%s.html", name))

	var inputs []string
	for _, src := range tcm.properties.Srcs {
		fmt.Println(src)
		if matches, err := ctx.GlobWithDeps(src, tcm.properties.SrcsExclude); err == nil {
			inputs = append(inputs, matches...)

		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			return
		}
	}
	//ctx.Build(pctx, blueprint.BuildParams{
	//	Description: fmt.Sprintf("Build %s as Go coverage", name),
	//	Rule:        goBuild,
	//	Outputs:     []string{output},
	//	Implicits:   inputs,
	//	Args: map[string]string{
	//		"workDir": ctx.ModuleDir(),
	//		"output":  output,
	//		"pkg":     tcm.properties.Pkg,
	//	},
	//})
	//
	//output := path.Join(config.BaseOutputDir, "config", name)

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Test coverage for %s", name),
		Rule:        goTestCoverage,
		Outputs:     []string{config.BaseOutputDir},
		Implicits:   inputs,
		Args: map[string]string{
			"workDir":        "testcoverage",
			"outputCoverage": pathToCoverageReports,
			"outputHtml":     pathToCoverageHtml,
		},
	})
}

func TestCoverageFactory() (blueprint.Module, []interface{}) {
	mType := &testedCoverageModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
