package boot

import (
	"github.com/EQEmu/spire/internal/console/cmd"
	"github.com/EQEmu/spire/internal/eqemuchangelog"
	"github.com/EQEmu/spire/internal/eqemuserver"
	"github.com/EQEmu/spire/internal/eqtraders"
	"github.com/EQEmu/spire/internal/generators"
	"github.com/EQEmu/spire/internal/model"
	"github.com/EQEmu/spire/internal/questapi"
	"github.com/EQEmu/spire/internal/spire"
	"github.com/EQEmu/spire/internal/user"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

// ProvideCommands returns the application command list in startup order.
func ProvideCommands(
	helloWorldCommand *cmd.HelloWorldCommand,
	userCreateCommand *user.CreateCommand,
	generateModelsCommand *model.GeneratorCommand,
	httpServeCommand *cmd.HttpServeCommand,
	routesListCommand *cmd.RoutesListCommand,
	spireMigrateCommand *spire.MigrateCommand,
	questApiParseCommand *questapi.ParseCommand,
	questExampleTestCommand *questapi.ExampleTestCommand,
	generateRaceModelMapsCommand *generators.RaceModelMapsCommand,
	changelogCmd *eqemuchangelog.ChangelogCommand,
	testFilesystemCmd *cmd.TestFilesystemCommand,
	spireInstallCmd *spire.InitCommand,
	userChangePasswordCmd *user.ChangePasswordCommand,
	spireCrashAnalyticsCommand *spire.CrashAnalyticsFingerprintBackfillCommand,
	eqEmuServerUpdateCommand *eqemuserver.UpdateCommand,
	eqEmuServerLauncherCommand *eqemuserver.LauncherCmd,
	eqEmuServerLauncherShimCommand *eqemuserver.LauncherShimCmd,
	scrapeEqtradersCommand *eqtraders.ScrapeCommand,
	importEqtradersCommand *eqtraders.ImportCommand,
) []*cobra.Command {
	return []*cobra.Command{
		helloWorldCommand.Command(),
		userCreateCommand.Command(),
		generateModelsCommand.Command(),
		httpServeCommand.Command(),
		routesListCommand.Command(),
		spireMigrateCommand.Command(),
		questApiParseCommand.Command(),
		questExampleTestCommand.Command(),
		generateRaceModelMapsCommand.Command(),
		changelogCmd.Command(),
		testFilesystemCmd.Command(),
		spireInstallCmd.Command(),
		userChangePasswordCmd.Command(),
		spireCrashAnalyticsCommand.Command(),
		eqEmuServerUpdateCommand.Command(),
		scrapeEqtradersCommand.Command(),
		importEqtradersCommand.Command(),
		eqEmuServerLauncherCommand.Command(),
		eqEmuServerLauncherShimCommand.Command(),
	}
}
