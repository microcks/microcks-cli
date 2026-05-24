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
 package connectors

 import (
	 "bytes"
	 "encoding/json"
	 "fmt"
	 "net/http"
 )
 
 // GitHubIssueRequest represents the payload for creating a GitHub Issue
 type GitHubIssueRequest struct {
	 Title  string   `json:"title"`
	 Body   string   `json:"body"`
	 Labels []string `json:"labels"`
 }
 
 // CreateGitHubIssue creates a GitHub Issue using the GitHub REST API
 func CreateGitHubIssue(token, repo, title, body string) error {
	 url := fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)
 
	 issueReq := GitHubIssueRequest{
		 Title:  title,
		 Body:   body,
		 Labels: []string{"bug", "microcks-test-failure"},
	 }
 
	 payload, err := json.Marshal(issueReq)
	 if err != nil {
		 return fmt.Errorf("failed to marshal issue request: %w", err)
	 }
 
	 req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	 if err != nil {
		 return fmt.Errorf("failed to create HTTP request: %w", err)
	 }
 
	 req.Header.Set("Authorization", "Bearer "+token)
	 req.Header.Set("Accept", "application/vnd.github+json")
	 req.Header.Set("Content-Type", "application/json")
	 req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
 
	 client := &http.Client{}
	 resp, err := client.Do(req)
	 if err != nil {
		 return fmt.Errorf("failed to call GitHub API: %w", err)
	 }
	 defer resp.Body.Close()
 
	 if resp.StatusCode != http.StatusCreated {
		 return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	 }
 
	 return nil
 }
 
 // BuildIssueBody builds a structured markdown body for the GitHub Issue
 func BuildIssueBody(serviceRef, testEndpoint, runnerType, serverAddr, testResultID string, failedOps []string) string {
	 ops := ""
	 for _, op := range failedOps {
		 ops += fmt.Sprintf("- ❌ `%s`\n", op)
	 }
	 if ops == "" {
		 ops = "- ❌ Test failed (no operation details available)\n"
	 }
 
	 return fmt.Sprintf(`## 🔴 Microcks Contract Test Failed
 
 ### Failed Operations
 %s
 
 ### Test Details
 | Field | Value |
 |-------|-------|
 | **Service** | %s |
 | **Endpoint** | %s |
 | **Runner** | %s |
 
 ### Reproduction Command
 `+"```"+`bash
 microcks test '%s' %s %s \
   --microcksURL=<your-microcks-url> \
   --keycloakClientId=<client-id> \
   --keycloakClientSecret=<client-secret>
 `+"```"+`
 
 ### Full Test Results
 [View on Microcks UI](%s/#/tests/%s)
 
 ---
 *This issue was automatically created by [microcks-cli](https://github.com/microcks/microcks-cli)*`,
		 ops,
		 serviceRef, testEndpoint, runnerType,
		 serviceRef, testEndpoint, runnerType,
		 serverAddr, testResultID,
	 )
 }