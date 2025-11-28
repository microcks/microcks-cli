package watcher

import (
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
)

func TriggerImport(entry config.WatchEntry) {
	// Retrieve config to get client options.
	cfgPath, err := config.DefaultLocalConfigPath()
	if err != nil {
		fmt.Errorf("Error while loading config: %s", err.Error())
	}

	fmt.Println("[INFO] Re-importing changed file: " + entry.FilePath)

	for _, context := range entry.Context {

		// Prepare Microcks client.
		var mc connectors.MicrocksClient

		// If config path exist, instantiate client with it.
		if _, err := os.Stat(cfgPath); err == nil {
			globalClientOpts := &connectors.ClientOptions{
				ConfigPath: cfgPath,
				Context:    context,
			}

			mc, err = connectors.NewClient(*globalClientOpts)
			if err != nil {
				fmt.Printf("[ERROR] Cannot connect to Microcks client: %v in context '%s'\n", err, context)
			}
		} else {
			// We have no config file, so just create a client with context as server URL.
			mc = connectors.NewMicrocksClient(context)
		}

		_, err = mc.UploadArtifact(entry.FilePath, entry.MainArtifact)
		if err != nil {
			fmt.Printf("[WARN] Error re-importing %s: %v\n", entry.FilePath, err)
		} else {
			fmt.Printf("[INFO] Successfully re-imported %s in context '%s'\n", entry.FilePath, context)
		}
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
