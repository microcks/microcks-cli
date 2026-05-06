package watcher

import (
	"context"
	"fmt"
	"os"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
)

func TriggerImport(ctx context.Context, entry config.WatchEntry) {
	cfgPath, err := config.DefaultLocalConfigPath()
	if err != nil {
		fmt.Printf("[ERROR] Error while loading config: %v\n", err)
		return
	}

	fmt.Println("[INFO] Re-importing changed file: " + entry.FilePath)

	for _, context := range entry.Context {
		var mc connectors.MicrocksClient

		if _, err := os.Stat(cfgPath); err == nil {
			globalClientOpts := &connectors.ClientOptions{
				ConfigPath: cfgPath,
				Context:    context,
			}

			mc, err = connectors.NewClient(*globalClientOpts)
			if err != nil {
				fmt.Printf("[ERROR] Cannot connect to Microcks client: %v in context '%s'\n", err, context)
				continue
			}
		} else {
			mc = connectors.NewMicrocksClient(context)
		}

		_, err = mc.UploadArtifact(ctx, entry.FilePath, entry.MainArtifact)
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
