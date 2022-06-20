package commands

import (
	"github.com/kluctl/kluctl/v2/cmd/kluctl/args"
	"github.com/kluctl/kluctl/v2/pkg/status"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"io/ioutil"
)

type renderCmd struct {
	args.ProjectFlags
	args.TargetFlags
	args.ArgsFlags
	args.ImageFlags
	args.RenderOutputDirFlags
}

func (cmd *renderCmd) Help() string {
	return `Renders all resources and configuration files and stores the result in either
a temporary directory or a specified directory.`
}

func (cmd *renderCmd) Run() error {
	if cmd.RenderOutputDir == "" {
		p, err := ioutil.TempDir(utils.GetTmpBaseDir(), "rendered-")
		if err != nil {
			return err
		}
		cmd.RenderOutputDir = p
	}

	ptArgs := projectTargetCommandArgs{
		projectFlags:         cmd.ProjectFlags,
		targetFlags:          cmd.TargetFlags,
		argsFlags:            cmd.ArgsFlags,
		imageFlags:           cmd.ImageFlags,
		renderOutputDirFlags: cmd.RenderOutputDirFlags,
	}
	return withProjectCommandContext(ptArgs, func(ctx *commandCtx) error {
		status.Info(ctx.ctx, "Rendered into %s", ctx.targetCtx.DeploymentCollection.RenderDir)
		return nil
	})
}
