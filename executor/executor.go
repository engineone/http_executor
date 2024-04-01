package executor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/engineone/types"
	"github.com/engineone/utils"
	validate "github.com/go-playground/validator/v10"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type Input struct {
	URL     string            `json:"url" valid:"required,url"`
	Method  string            `json:"method" valid:"required,in(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)"`
	Headers map[string]string `json:"headers" valid:"required,dictionary"`
	Body    interface{}       `json:"body"`
}

type Output struct {
	Headers map[string]string `json:"headers" valid:"required,dictionary"`
	Body    interface{}       `json:"body"`
}

type HttpExecutor struct {
	validator  *validate.Validate
	inputCache *Input
}

// NewHttpExecutor creates a new HttpExecutor
func NewHttpExecutor() *HttpExecutor {
	return &HttpExecutor{
		validator: utils.NewValidator(),
	}
}

func (e *HttpExecutor) New() types.Executor {
	return NewHttpExecutor()
}

func (e *HttpExecutor) ID() string {
	return "http"
}

func (e *HttpExecutor) Name() string {
	return "HTTP"
}

func (e *HttpExecutor) InputRules() map[string]interface{} {
	return utils.ExtractValidationRules(&Input{})
}

func (e *HttpExecutor) OutputRules() map[string]interface{} {
	return utils.ExtractValidationRules(&Output{})
}

func (e *HttpExecutor) Description() string {
	return "Http executor to make http requests to a given url with the given method and headers."
}

func (e *HttpExecutor) convertInput(input interface{}) (*Input, error) {
	if e.inputCache != nil {
		return e.inputCache, nil
	}

	e.inputCache = &Input{}
	if err := utils.ConvertToStruct(input, e.inputCache); err != nil {
		return nil, stacktrace.PropagateWithCode(err, types.ErrInvalidInput, "Error converting input to struct")
	}
	return e.inputCache, nil
}

func (e *HttpExecutor) Validate(ctx context.Context, task *types.Task, otherTasks []*types.Task) error {
	if task.Input == nil {
		return stacktrace.NewErrorWithCode(types.ErrInvalidInput, "Input is required")
	}

	input, err := e.convertInput(task.Input)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to convert input")
	}

	err = e.validator.Struct(input)
	return stacktrace.PropagateWithCode(err, types.ErrInvalidInput, "Input validation failed")
}

func (e *HttpExecutor) Execute(ctx context.Context, task *types.Task, otherTasks []*types.Task) (interface{}, error) {
	logrus.Debugf("Executing task %s in an http executor", task.ID)

	input, err := e.convertInput(task.Input)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to convert input")
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

	var req *http.Request
	if input.Method == "POST" || input.Method == "PUT" || input.Method == "PATCH" {
		var body io.Reader
		bodyContent, ok := input.Body.(string)
		if ok {
			body = strings.NewReader(bodyContent)
		} else {
			bodyBytes, _ := json.Marshal(bodyContent)
			body = strings.NewReader(string(bodyBytes))
		}

		req, err = http.NewRequest(input.Method, input.URL, body)
	} else {
		req, err = http.NewRequest(input.Method, input.URL, nil)
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create the http request")
	}

	if input.Headers != nil {
		for key, value := range input.Headers {
			req.Header.Set(key, value)
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
