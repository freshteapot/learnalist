package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/freshteapot/e2elog"
	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	openapiClient *e2e.OpenApiClient
	e2eClient     e2e.Client
)

func TestE2e(t *testing.T) {
	logger := logging.GetLogger()
	logContext := logger.WithField("context", "e2e")

	e2eStanClientID := "e2e"
	event.SetDefaultSettingsForCMD()
	os.Setenv("EVENTS_STAN_CLIENT_ID", e2eStanClientID)
	os.Setenv("EVENTS_STAN_CLUSTER_ID", "test-cluster")
	os.Setenv("EVENTS_NATS_SERVER", "127.0.0.1")
	event.SetupEventBus(logContext)

	event.GetBus().Start(event.TopicNotifications)

	openapiClient = e2e.NewOpenApiClient(e2e.LOCAL_SERVER)
	e2eClient = e2e.NewClient("http://localhost:1234")
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")

	logs := openapiClient.GetLogs()

	logFile := "/tmp/learnalist/e2e.log"
	os.Remove(logFile)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	for _, httpLog := range logs {
		b, _ := json.Marshal(httpLog)
		f.Write(b)
		f.Write([]byte("\n"))
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

	b, _ := ioutil.ReadFile("/tmp/openapi/one/learnalist.yaml")
	s, _ := openapi3.NewSwaggerLoader().LoadSwaggerFromData(b)
	summary, _ := e2elog.Coverage(s, logFile, "/api/v1")
	fmt.Println(summary.Coverage)
}
