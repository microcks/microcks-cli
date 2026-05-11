package output

import "github.com/microcks/microcks-cli/pkg/connectors"

// TestResultJSON is a serialized view of connectors.TestResult.
type TestResultJSON struct {
	ID             string               `json:"id" yaml:"id"`
	Version        int32                `json:"version" yaml:"version"`
	TestNumber     int32                `json:"testNumber" yaml:"testNumber"`
	TestDate       int64                `json:"testDate" yaml:"testDate"`
	TestedEndpoint string               `json:"testedEndpoint" yaml:"testedEndpoint"`
	ServiceID      string               `json:"serviceId" yaml:"serviceId"`
	ElapsedTime    int64                `json:"elapsedTime" yaml:"elapsedTime"`
	Success        bool                 `json:"success" yaml:"success"`
	InProgress     bool                 `json:"inProgress" yaml:"inProgress"`
	RunnerType     string               `json:"runnerType" yaml:"runnerType"`
	TestCases      []TestCaseResultJSON `json:"testCaseResults" yaml:"testCaseResults"`
}

type TestCaseResultJSON struct {
	Success         bool                 `json:"success" yaml:"success"`
	ElapsedTime     int64                `json:"elapsedTime" yaml:"elapsedTime"`
	OperationName   string               `json:"operationName" yaml:"operationName"`
	TestStepResults []TestStepResultJSON `json:"testStepResults" yaml:"testStepResults"`
}

type TestStepResultJSON struct {
	Success          bool   `json:"success" yaml:"success"`
	ElapsedTime      int64  `json:"elapsedTime" yaml:"elapsedTime"`
	RequestName      string `json:"requestName" yaml:"requestName"`
	Message          string `json:"message" yaml:"message"`
	EventMessageName string `json:"eventMessageName" yaml:"eventMessageName"`
}

func NewTestResultJSON(result *connectors.TestResult) TestResultJSON {
	testCases := make([]TestCaseResultJSON, 0, len(result.TestCases))
	for _, tc := range result.TestCases {
		steps := make([]TestStepResultJSON, 0, len(tc.TestStepResults))
		for _, step := range tc.TestStepResults {
			steps = append(steps, TestStepResultJSON{
				Success:          step.Success,
				ElapsedTime:      step.ElapsedTime,
				RequestName:      step.RequestName,
				Message:          step.Message,
				EventMessageName: step.EventMessageName,
			})
		}
		testCases = append(testCases, TestCaseResultJSON{
			Success:         tc.Success,
			ElapsedTime:     tc.ElapsedTime,
			OperationName:   tc.OperationName,
			TestStepResults: steps,
		})
	}

	return TestResultJSON{
		ID:             result.ID,
		Version:        result.Version,
		TestNumber:     result.TestNumber,
		TestDate:       result.TestDate,
		TestedEndpoint: result.TestedEndpoint,
		ServiceID:      result.ServiceID,
		ElapsedTime:    result.ElapsedTime,
		Success:        result.Success,
		InProgress:     result.InProgress,
		RunnerType:     result.RunnerType,
		TestCases:      testCases,
	}
}
