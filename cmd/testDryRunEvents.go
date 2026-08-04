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
	"io"
	"time"

	"github.com/microcks/microcks-cli/pkg/connectors"
)

type dryRunWatchEvent struct {
	Type         string                 `json:"type"`
	Timestamp    string                 `json:"timestamp"`
	Endpoint     string                 `json:"endpoint,omitempty"`
	Artifact     string                 `json:"artifact,omitempty"`
	Service      string                 `json:"service,omitempty"`
	TestResultID string                 `json:"testResultId,omitempty"`
	Result       *connectors.TestResult `json:"result,omitempty"`
	Message      string                 `json:"message,omitempty"`
}

type dryRunEventWriter struct {
	encoder *json.Encoder
}

func newDryRunEventWriter(w io.Writer) *dryRunEventWriter {
	return &dryRunEventWriter{encoder: json.NewEncoder(w)}
}

func (w *dryRunEventWriter) emit(event dryRunWatchEvent) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	return w.encoder.Encode(event)
}

func (w *dryRunEventWriter) emitTestResult(mc connectors.MicrocksClient, testResultID string) error {
	result, err := mc.GetFullTestResult(testResultID)
	if err != nil {
		return err
	}
	return w.emit(dryRunWatchEvent{
		Type:         "test-result",
		TestResultID: testResultID,
		Result:       result,
	})
}
