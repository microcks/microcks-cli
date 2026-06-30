/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/spf13/cobra"
)

var exitFunc = os.Exit
var watchSignalChan chan os.Signal

type ServerCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type SpringComponent struct {
	Status string `json:"status"`
}

type ServerHealth struct {
	Status     string                     `json:"status"`
	Checks     []ServerCheck              `json:"checks,omitempty"`
	Components map[string]SpringComponent `json:"components,omitempty"`
}

type JSONOutput struct {
	Status    string      `json:"status"`
	LatencyMS int64       `json:"latency_ms"`
	Version   string      `json:"version,omitempty"`
	Checks    []CheckInfo `json:"checks,omitempty"`
}

type CheckInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func NewHealthCommand(globalClientOpts *connectors.ClientOptions) *cobra.Command {
	var (
		jsonOutput bool
		watch      bool
		interval   time.Duration
	)

	var healthCmd = &cobra.Command{
		Use:   "health",
		Short: "Check the health and diagnostics of the Microcks server",
		Long:  `Check the health and diagnostics of the Microcks server`,
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize config from command options.
			config.InsecureTLS = globalClientOpts.InsecureTLS
			config.CaCertPaths = globalClientOpts.CaCertPaths
			config.Verbose = globalClientOpts.Verbose

			// Determine Server URL.
			var serverURL string
			if globalClientOpts.ServerAddr != "" {
				serverURL = globalClientOpts.ServerAddr
			} else {
				localConfig, err := config.ReadLocalConfig(globalClientOpts.ConfigPath)
				if err != nil {
					fmt.Fprintln(cmd.OutOrStdout(), "Error reading config:", err)
					exitFunc(1)
					return
				}
				if localConfig == nil {
					fmt.Fprintln(cmd.OutOrStdout(), "No Microcks server URL configured. Please specify with --microcksURL or run 'microcks login'")
					exitFunc(1)
					return
				}
				ctxName := globalClientOpts.Context
				if ctxName == "" {
					ctxName = localConfig.CurrentContext
				}
				configCtx, err := localConfig.ResolveContext(ctxName)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "Error resolving context '%s': %v\n", ctxName, err)
					exitFunc(1)
					return
				}
				serverURL = configCtx.Server.Server
			}

			httpClient := getHTTPClient()
			candidates, schemeHost, apiPath := getHealthCandidates(serverURL)

			if watch {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()

				sigChan := make(chan os.Signal, 1)
				if watchSignalChan != nil {
					sigChan = watchSignalChan
				} else {
					signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				}

				// Run immediately on start
				code := executeHealthCheck(cmd.OutOrStdout(), httpClient, candidates, schemeHost, apiPath, serverURL, jsonOutput, watch)

				for {
					select {
					case <-ticker.C:
						code = executeHealthCheck(cmd.OutOrStdout(), httpClient, candidates, schemeHost, apiPath, serverURL, jsonOutput, watch)
					case <-sigChan:
						exitFunc(code)
						return
					}
				}
			} else {
				code := executeHealthCheck(cmd.OutOrStdout(), httpClient, candidates, schemeHost, apiPath, serverURL, jsonOutput, watch)
				exitFunc(code)
			}
		},
	}

	healthCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output machine-readable JSON")
	healthCmd.Flags().BoolVar(&watch, "watch", false, "Repeat health checks periodically")
	healthCmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "Watch interval")

	return healthCmd
}

func getHTTPClient() *http.Client {
	var tr *http.Transport
	if config.InsecureTLS || len(config.CaCertPaths) > 0 {
		tlsConfig := config.CreateTLSConfig()
		tr = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	} else {
		tr = &http.Transport{}
	}

	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}

func getHealthCandidates(serverAddr string) ([]string, string, string) {
	parsed, err := url.Parse(serverAddr)
	if err != nil {
		return []string{serverAddr + "/health"}, serverAddr, "/api/"
	}

	schemeHost := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// If the path doesn't contain "/api", normalize API path prefix
	apiPath := path
	if !strings.Contains(path, "/api") {
		apiPath = path + "api/"
		apiPath = strings.ReplaceAll(apiPath, "//", "/")
	}

	var candidates []string
	// Candidate 1: relative to the configured URL path (usually /api/health)
	candidates = append(candidates, schemeHost+apiPath+"health")
	// Candidate 2: /q/health on the host
	candidates = append(candidates, schemeHost+"/q/health")
	// Candidate 3: /actuator/health on the host
	candidates = append(candidates, schemeHost+"/actuator/health")

	return candidates, schemeHost, apiPath
}

func executeHealthCheck(writer io.Writer, client *http.Client, candidates []string, schemeHost string, apiPath string, serverURL string, jsonOutput bool, watch bool) int {
	sh, statusCode, latency, err := runHealthCheck(client, candidates)

	// Measure and detect version, Keycloak, etc.
	version := "unknown"
	if err == nil {
		version = detectServerVersion(client, schemeHost, apiPath)
	}

	keycloakEnabled := false
	keycloakReachable := false
	kcStatus, kcURL := checkKeycloak(client, schemeHost, apiPath)
	if kcStatus == "enabled" {
		keycloakEnabled = true
		keycloakReachable = true
	} else if kcStatus == "unreachable" {
		keycloakEnabled = true
		keycloakReachable = false
	}

	var checks []CheckInfo
	dbConnected := true
	asyncConnected := true

	if sh != nil {
		for _, c := range sh.Checks {
			nameLower := strings.ToLower(c.Name)
			statusUpper := strings.ToUpper(c.Status)
			isUp := statusUpper == "UP"

			friendlyName := c.Name
			if strings.Contains(nameLower, "database") || strings.Contains(nameLower, "mongodb") || strings.Contains(nameLower, "db") {
				friendlyName = "Database"
				if !isUp {
					dbConnected = false
				}
			} else if strings.Contains(nameLower, "keycloak") {
				friendlyName = "Keycloak"
			} else if strings.Contains(nameLower, "minion") || strings.Contains(nameLower, "async") || strings.Contains(nameLower, "kafka") || strings.Contains(nameLower, "producer") {
				friendlyName = "Async Minion"
				if !isUp {
					asyncConnected = false
				}
			}

			checks = append(checks, CheckInfo{
				Name:   friendlyName,
				Status: statusUpper,
			})
		}
	}

	// Ensure we report keycloak status explicitly in checks if not present
	hasKeycloakCheck := false
	for _, c := range checks {
		if c.Name == "Keycloak" {
			hasKeycloakCheck = true
			break
		}
	}
	if !hasKeycloakCheck {
		kcVal := "DOWN"
		if keycloakEnabled && keycloakReachable {
			kcVal = "UP"
		} else if !keycloakEnabled {
			kcVal = "DISABLED"
		}
		checks = append(checks, CheckInfo{
			Name:   "Keycloak",
			Status: kcVal,
		})
	}

	// Ensure Database status is reflected
	hasDbCheck := false
	for _, c := range checks {
		if c.Name == "Database" {
			hasDbCheck = true
			break
		}
	}
	if !hasDbCheck {
		dbVal := "UP"
		if sh != nil && strings.ToUpper(sh.Status) != "UP" {
			dbVal = "DOWN"
			dbConnected = false
		}
		checks = append(checks, CheckInfo{
			Name:   "Database",
			Status: dbVal,
		})
	}

	// Ensure Async Minion status is reflected
	hasAsyncCheck := false
	for _, c := range checks {
		if c.Name == "Async Minion" {
			hasAsyncCheck = true
			break
		}
	}
	if !hasAsyncCheck {
		asyncVal := "UP"
		if sh != nil && strings.ToUpper(sh.Status) != "UP" {
			asyncVal = "DOWN"
			asyncConnected = false
		}
		checks = append(checks, CheckInfo{
			Name:   "Async Minion",
			Status: asyncVal,
		})
	}

	// Compute Overall Status
	overallStatus := "HEALTHY"
	exitCode := 0

	if err != nil {
		overallStatus = "UNHEALTHY"
		exitCode = 1
	} else {
		serverHealthy := strings.ToUpper(sh.Status) == "UP"
		allSubsystemsUp := dbConnected && (!keycloakEnabled || keycloakReachable) && asyncConnected

		if !serverHealthy {
			overallStatus = "UNHEALTHY"
			exitCode = 2 // Degraded subsystem or partial health
		} else if !allSubsystemsUp {
			overallStatus = "DEGRADED"
			exitCode = 2 // Degraded subsystem or partial health
		}
	}

	if jsonOutput {
		jo := JSONOutput{
			Status:    "UP",
			LatencyMS: latency.Milliseconds(),
			Version:   version,
			Checks:    checks,
		}
		if err != nil {
			jo.Status = "DOWN"
			jo.Checks = nil
		} else if overallStatus == "UNHEALTHY" {
			jo.Status = "DOWN"
		}
		outBytes, _ := json.MarshalIndent(jo, "", "  ")
		fmt.Fprintln(writer, string(outBytes))
		return exitCode
	}

	if watch {
		fmt.Fprint(writer, "\033[H\033[2J") // ANSI clear screen and move cursor to top-left
	}

	fmt.Fprintln(writer, "Microcks Server Health Check")
	fmt.Fprintln(writer, "============================")
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Server URL: %s\n", serverURL)
	fmt.Fprintln(writer)

	if err != nil {
		friendlyErr := err.Error()
		if strings.Contains(friendlyErr, "connection refused") {
			friendlyErr = "Connection refused"
		} else if strings.Contains(friendlyErr, "no such host") {
			friendlyErr = "DNS resolution failed"
		} else if strings.Contains(friendlyErr, "timeout") {
			friendlyErr = "Connection timed out"
		}
		fmt.Fprintf(writer, "Unreachable (%s)\n", friendlyErr)
		fmt.Fprintln(writer)
		fmt.Fprintln(writer, "Overall Status: UNHEALTHY")
		return 1
	}

	fmt.Fprintf(writer, "Reachable (HTTP %d, %v)\n", statusCode, latency.Round(time.Millisecond))
	fmt.Fprintln(writer, "API endpoint responding")

	if keycloakEnabled {
		if keycloakReachable {
			fmt.Fprintln(writer, "Keycloak: enabled and reachable")
		} else {
			fmt.Fprintf(writer, "Keycloak: enabled but unreachable (%s)\n", kcURL)
		}
	} else {
		fmt.Fprintln(writer, "Keycloak: disabled")
	}

	if asyncConnected {
		fmt.Fprintln(writer, "Async Minion: connected")
	} else {
		fmt.Fprintln(writer, "Async Minion: disconnected")
	}

	if dbConnected {
		fmt.Fprintln(writer, "Database: connected")
	} else {
		fmt.Fprintln(writer, "Database: disconnected")
	}

	if version != "unknown" {
		fmt.Fprintf(writer, "Version: %s\n", version)
	}

	fmt.Fprintln(writer)
	if overallStatus == "HEALTHY" {
		fmt.Fprintln(writer, "Overall Status: HEALTHY")
	} else if overallStatus == "DEGRADED" {
		fmt.Fprintln(writer, "Overall Status: DEGRADED")
	} else {
		fmt.Fprintln(writer, "Overall Status: UNHEALTHY")
	}

	return exitCode
}

func runHealthCheck(client *http.Client, candidates []string) (*ServerHealth, int, time.Duration, error) {
	var lastErr error
	for _, urlStr := range candidates {
		start := time.Now()
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		latency := time.Since(start)

		if resp.StatusCode == http.StatusNotFound {
			lastErr = fmt.Errorf("HTTP 404 Not Found")
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, latency, err
		}

		var sh ServerHealth
		if err := json.Unmarshal(body, &sh); err != nil {
			// Try to parse basic status if it's not a full health check structure
			var basic map[string]any
			if json.Unmarshal(body, &basic) == nil {
				if statusVal, ok := basic["status"].(string); ok {
					sh.Status = statusVal
					return &sh, resp.StatusCode, latency, nil
				}
			}
			return nil, resp.StatusCode, latency, fmt.Errorf("failed to parse JSON response: %w", err)
		}

		// Spring Boot Actuator fallback parsing
		if len(sh.Checks) == 0 && len(sh.Components) > 0 {
			for name, comp := range sh.Components {
				sh.Checks = append(sh.Checks, ServerCheck{
					Name:   name,
					Status: comp.Status,
				})
			}
		}

		// Default Overall Status to UP if not set but we parsed the JSON successfully
		if sh.Status == "" {
			sh.Status = "UP"
		}

		return &sh, resp.StatusCode, latency, nil
	}
	return nil, 0, 0, lastErr
}

func checkKeycloak(client *http.Client, schemeHost string, apiPath string) (string, string) {
	urlStr := schemeHost + apiPath + "keycloak/config"
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "disabled", ""
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "disabled", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "disabled", ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "disabled", ""
	}

	var configResp map[string]any
	if err := json.Unmarshal(body, &configResp); err != nil {
		return "disabled", ""
	}

	enabledVal, ok := configResp["enabled"].(bool)
	if !ok || !enabledVal {
		return "disabled", ""
	}

	authServerURL, _ := configResp["auth-server-url"].(string)
	realmName, _ := configResp["realm"].(string)
	if authServerURL == "" || realmName == "" {
		return "enabled", ""
	}

	targetURL := authServerURL + "/realms/" + realmName + "/"
	// Quick reachability test on Keycloak
	kcClient := &http.Client{
		Timeout: 2 * time.Second,
	}
	kcReq, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "unreachable", targetURL
	}
	kcResp, err := kcClient.Do(kcReq)
	if err != nil {
		return "unreachable", targetURL
	}
	defer kcResp.Body.Close()

	return "enabled", targetURL
}

func detectServerVersion(client *http.Client, schemeHost string, apiPath string) string {
	// Try /api/info or /actuator/info
	infoURLs := []string{
		schemeHost + apiPath + "info",
		schemeHost + "/actuator/info",
		schemeHost + "/api/info",
	}
	for _, infoURL := range infoURLs {
		req, err := http.NewRequest("GET", infoURL, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				var infoData map[string]any
				if err := json.Unmarshal(body, &infoData); err == nil {
					if v := getNestedString(infoData, "app", "version"); v != "" {
						return v
					}
					if v := getNestedString(infoData, "build", "version"); v != "" {
						return v
					}
					if v, ok := infoData["version"].(string); ok && v != "" {
						return v
					}
				}
			}
		}
	}

	// Try features/config (since it's unprotected)
	featuresURL := schemeHost + apiPath + "features/config"
	req, err := http.NewRequest("GET", featuresURL, nil)
	if err == nil {
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				var featuresData map[string]any
				if err := json.Unmarshal(body, &featuresData); err == nil {
					if v, ok := featuresData["version"].(string); ok && v != "" {
						return v
					}
				}
			}
		}
	}

	return "unknown"
}

func getNestedString(data map[string]any, keys ...string) string {
	var current any = data
	for _, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = m[key]
	}
	if s, ok := current.(string); ok {
		return s
	}
	return ""
}
