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
	"testing"

	"github.com/microcks/microcks-cli/pkg/connectors"
	"github.com/microcks/microcks-cli/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMicrocksClientPollingTimeout struct {
	connectors.MicrocksClient
}

func (m *mockMicrocksClientPollingTimeout) CreateTestResult(serviceID, testEndpoint, runnerType, secretName string, timeout int64, filteredOperations, operationsHeaders, oAuth2Context string) (string, error) {
	return "test-result-timeout-id", nil
}

func (m *mockMicrocksClientPollingTimeout) GetTestResult(testResultID string) (*connectors.TestResultSummary, error) {
	return &connectors.TestResultSummary{
		ID:         testResultID,
		Success:    false,
		InProgress: true,
	}, nil
}

func TestRunTestAndWait_TimeoutReturnsKindConnection(t *testing.T) {
	client := &mockMicrocksClientPollingTimeout{}
	params := testParams{
		waitForMillis:  -10000, // deadline <= now so the polling loop deadline has already expired
		suppressOutput: true,
	}

	success, id, err := runTestAndWait(client, params)

	assert.False(t, success)
	assert.Equal(t, "test-result-timeout-id", id)
	require.Error(t, err)
	assert.Equal(t, errors.KindConnection, errors.KindOf(err))
	assert.Equal(t, 11, ExitCodeFor(err))
	assert.Contains(t, err.Error(), "test polling timed out after 0ms waiting for result test-result-timeout-id")
}
