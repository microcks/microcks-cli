package watcher

import (
	"fmt"
	"strconv"

	"github.com/microcks/microcks-cli/cmd"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
)

func TriggerImport(entry config.WatchEntry) {
	mainArtifact := strconv.FormatBool(entry.MainArtifact)

	args := []string{
		entry.FilePath + ":" + mainArtifact,
	}

	cfgPath, err := config.DefaultLocalConfigPath()
	if err != nil {
		fmt.Errorf("Error while loading config: %s", err.Error())
	}

	for _, context := range entry.Context {
		importCommand := cmd.NewImportCommand(&connectors.ClientOptions{
			ConfigPath: cfgPath,
			Context:    context,
		})
		importCommand.SetArgs(args)
		err = importCommand.Execute()
		if err != nil {
			fmt.Printf("Error re-importing %s: %v\n", entry.FilePath, err)
		}

		fmt.Printf("Imported '%s' in context '%s'\n", entry.FilePath, context)
	}
}

func LoadRegistry(watchFilePath string) (*config.WatchConfig, error) {
	var watchCfg *config.WatchConfig
	watchCfg, err := config.ReadLocalWatchConfig(watchFilePath)
	if err != nil {
		return nil, err
	}

	if watchCfg == nil {
		watchCfg = &config.WatchConfig{}
	}

	return watchCfg, nil
}
