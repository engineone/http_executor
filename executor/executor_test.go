package executor_test

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"

// 	"github.com/engineone/types"
// 	_ "github.com/lepsta/engineone/pkg/executors"
// 	http_executor "github.com/lepsta/engineone/pkg/executors/http"
// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"
// 	"github.com/palantir/stacktrace"
// )

// var _ = Describe("Http Executor", func() {
// 	var (
// 		executor *http_executor.HttpExecutor
// 		task     *types.Task
// 		wf       *engine.Workflow
// 		server   *httptest.Server
// 	)

// 	BeforeEach(func() {
// 		executor = http_executor.NewHttpExecutor()
// 		task = &types.EngineOneTask{
// 			ID:       "task1",
// 			Executor: "http",
// 			Input: map[string]interface{}{
// 				"url":     "http://example.com",
// 				"method":  "GET",
// 				"headers": map[string]interface{}{},
// 			},
// 		}

// 		wf = engine.NewWorkflowBuilder().
// 			WithID("wf1").
// 			WithName("Workflow 1").
// 			WithNamespace("test").
// 			WithTasks([]*types.EngineOneTask{
// 				task,
// 			}).
// 			Build()

// 		// Start a mock server
// 		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			w.WriteHeader(http.StatusOK)
// 			fmt.Fprint(w, "Hello, World!")
// 		}))
// 	})

// 	AfterEach(func() {
// 		// Close the mock server
// 		server.Close()
// 	})

// 	Describe("Execute", func() {
// 		Context("When the input is valid", func() {
// 			It("should make an HTTP request and return the response", func() {
// 				task.Input = map[string]interface{}{
// 					"url":     server.URL,
// 					"method":  "GET",
// 					"headers": map[string]interface{}{},
// 				}

// 				err := executor.Validate(context.TODO(), task, wf.Tasks)
// 				Expect(err).ToNot(HaveOccurred())

// 				output, err := executor.Execute(context.TODO(), task, wf.Tasks)
// 				Expect(err).ToNot(HaveOccurred())
// 				Expect(output).ToNot(BeNil())
// 				Expect(output.(map[string]interface{})["headers"]).ToNot(BeNil())
// 				Expect(output.(map[string]interface{})["body"]).To(Equal([]byte("Hello, World!")))
// 			})
// 		})

// 		Context("When the input is invalid", func() {
// 			It("should return an error", func() {
// 				task.Input = "invalid input"

// 				err := executor.Validate(context.TODO(), task, wf.Tasks)
// 				Expect(err).To(HaveOccurred())
// 				Expect(stacktrace.GetCode(err)).To(Equal(engine.ErrInvalidInput))
// 			})
// 		})

// 		Context("When the HTTP request fails", func() {
// 			It("should return an error", func() {
// 				task.Input = map[string]interface{}{
// 					"url":     "http://nonexistent",
// 					"method":  "GET",
// 					"headers": map[string]interface{}{},
// 				}

// 				_, err := executor.Execute(context.TODO(), task, wf.Tasks)
// 				Expect(err).To(HaveOccurred())
// 			})
// 		})
// 	})
// })
