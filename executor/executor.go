package executor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/engineone/types"
	"github.com/engineone/utils"
	"github.com/palantir/stacktrace"
	"github.com/sergeyglazyrindev/govalidator"
	"github.com/sirupsen/logrus"
)

type HttpExecutor struct {
	inputRules  map[string]interface{}
	outputRules map[string]interface{}
}

// NewHttpExecutor creates a new HttpExecutor
func NewHttpExecutor() *HttpExecutor {
	return &HttpExecutor{
		// Example:
		// --------
		// url: http://localhost:8080
		// method: POST
		// headers:
		// 	Content-Type: application/json
		// body: |
		// 	{
		// 		"name": "John Doe",
		// 		"age": 25
		// 	}
		inputRules: map[string]interface{}{
			"url":     "required,url",
			"method":  "required,in(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)",
			"headers": "required,dictionary",
			"body":    "",
		},
		// Example:
		// --------
		// headers:
		// 	Content-Type: application/json
		// body: |
		// 	{
		// 		"name": "John Doe",
		// 		"age": 25
		// 	}
		outputRules: map[string]interface{}{
			"headers": "required,dictionary",
			"body":    "",
		},
	}
}

func (e *HttpExecutor) New() *HttpExecutor {
	return NewHttpExecutor()
}

func (e *HttpExecutor) ID() string {
	return "http"
}

func (e *HttpExecutor) Name() string {
	return "HTTP"
}

func (e *HttpExecutor) InputRules() map[string]interface{} {
	return e.inputRules
}

func (e *HttpExecutor) OutputRules() map[string]interface{} {
	return e.outputRules
}

func (e *HttpExecutor) Description() string {
	return "Http executor to make http requests to a given url with the given method and headers."
}

func (e *HttpExecutor) Validate(ctx context.Context, task *types.Task, otherTasks []*types.Task) error {
	if task.Input == nil {
		return stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input is required")
	}

	input, ok := task.Input.(map[string]interface{})
	if !ok {
		return stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input must be an object")
	}

	_, err := govalidator.ValidateMap(input, e.inputRules)
	return stacktrace.PropagateWithCode(err, types.ErrInvalidTask, "Input validation failed")
}

func (e *HttpExecutor) Execute(ctx context.Context, task *types.Task, otherTasks []*types.Task) (interface{}, error) {
	logrus.Debugf("Executing task %s in an http executor", task.ID)

	input, ok := task.Input.(map[string]interface{})
	if !ok {
		return nil, stacktrace.NewErrorWithCode(types.ErrInvalidTask, "Input must be an object")
	}

	// See if we need to render the tamplate input
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to marshal the input")
	}

	if strings.Contains(string(inputBytes), "{{") {
		inputString, err := utils.RenderInputTemplate(string(inputBytes), task, otherTasks)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to render the input template")
		}

		if err := json.Unmarshal([]byte(inputString), &input); err != nil {
			return nil, stacktrace.Propagate(err, "Failed to unmarshal the input")
		}
	}

	url, _ := input["url"].(string)
	method, _ := input["method"].(string)

	var req *http.Request
	if method == "POST" || method == "PUT" || method == "PATCH" {
		var body io.Reader
		bodyContent, ok := input["body"].(string)
		if ok {
			body = strings.NewReader(bodyContent)
		} else {
			bodyBytes, _ := json.Marshal(input["body"])
			body = strings.NewReader(string(bodyBytes))
		}

		req, err = http.NewRequest(method, url, body)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the http request")
	}

	headers, ok := input["headers"].(map[string]interface{})
	if ok {
		for key, value := range headers {
			req.Header.Set(key, value.(string))
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to send the http request")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to read the response body")
	}

	output := map[string]interface{}{
		"headers": resp.Header,
		"body":    data,
	}
	return output, nil
}
