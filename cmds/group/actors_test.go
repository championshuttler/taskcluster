package group

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
	tcclient "github.com/taskcluster/taskcluster-client-go"
)

const fakeTaskID = "ANnmjMocTymeTID0tlNJAw"
const fakeRunID = "0"
const fakeGroupID = "e4WPAAeSdaSdKxeWzDCBA"
const badGroupID = "AAAAAAAAAAAAAAAAAAAAA"

type FakeServerSuite struct {
	suite.Suite
	testServer *httptest.Server
}

func (suite *FakeServerSuite) SetupSuite() {
	// set up a fake server that knows how to answer the `task()` method
	handler := http.NewServeMux()

	handler.HandleFunc("/v1/task/"+fakeTaskID+"/cancel", cancelHandler)
	handler.HandleFunc("/v1/task-group/"+fakeGroupID+"/list", listTaskGroupHandler)

	suite.testServer = httptest.NewServer(handler)

	// set the base URL the subcommands use to point to the fake server
	queueBaseURL = suite.testServer.URL + "/v1"
}

func (suite *FakeServerSuite) TearDownSuite() {
	suite.testServer.Close()
	queueBaseURL = ""
}

func TestFakeServerSuite(t *testing.T) {
	suite.Run(t, new(FakeServerSuite))
}

// returns the test status on request
func cancelHandler(w http.ResponseWriter, _ *http.Request) {
	status := `{
				  "status": {
				    "state": "cancelled",
				    "runs": [
				      {
				        "runId": 0,
				        "state": "cancelled",
				        "reasonCreated": "scheduled",
				        "reasonResolved": "cancelled"
				      }
				    ]
				  }
				}`
	io.WriteString(w, status)
}

func listTaskGroupHandler(w http.ResponseWriter, _ *http.Request) {
	list := `{
			  "taskGroupId": "e4WPAAeSdaSdKxeWzDCBA",
			  "tasks": [
			    {
			      "status": {
			        "taskId": "ANnmjMocTymeTID0tlNJAw",
			        "provisionerId": "aws-provisioner-v1",
			        "workerType": "github-worker",
			        "schedulerId": "taskcluster-github",
			        "taskGroupId": "e4WPJRJeSdaSdKxeWzDlNQ",
			        "deadline": "2017-03-30T15:49:31.389Z",
			        "expires": "2018-03-30T15:49:31.389Z",
			        "retriesLeft": 5,
			        "state": "pending",
			        "runs": [
			          {
			            "runId": 0,
			            "state": "pending",
			            "reasonCreated": "scheduled",
			            "reasonResolved": "failed",
			            "workerGroup": "us-west-2",
			            "workerId": "i-06936339d4f83059a",
			            "takenUntil": "2017-03-29T16:10:32.326Z",
			            "scheduled": "2017-03-29T15:49:32.292Z",
			            "started": "2017-03-29T15:50:32.412Z",
			            "resolved": "2017-03-29T15:53:27.562Z"
			          }
			        ]
			      },
						"task": {
						  "metadata": {
							  "name": "test-framework-task/opt"
							}
						}
			    }
			  ]
			}`

	io.WriteString(w, list)
}

func setUpCommand() (*bytes.Buffer, *cobra.Command) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOutput(buf)

	return buf, cmd
}

func (suite *FakeServerSuite) TestRunCancel() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	cmd.Flags().Bool("force", true, "")

	// run the command
	args := []string{fakeGroupID}
	runCancel(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("cancelling task ANnmjMocTymeTID0tlNJAw\n", string(buf.Bytes()))
}

func (suite *FakeServerSuite) TestRunStatus() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	// run the command
	args := []string{fakeGroupID}
	runStatus(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("pending: 1\n", string(buf.Bytes()))
}

func (suite *FakeServerSuite) TestRunStatusBadGroupId() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	// run the command
	args := []string{badGroupID}
	runStatus(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("", string(buf.Bytes()))
}

func (suite *FakeServerSuite) TestRunListAll() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	cmd.Flags().Bool("all", true, "")

	// run the command
	args := []string{fakeGroupID}
	runList(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("ANnmjMocTymeTID0tlNJAw test-framework-task/opt pending\n", string(buf.Bytes()))
}

func (suite *FakeServerSuite) TestRunListPending() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	cmd.Flags().Bool("pending", true, "")

	// run the command
	args := []string{fakeGroupID}
	runList(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("ANnmjMocTymeTID0tlNJAw test-framework-task/opt pending\n", string(buf.Bytes()))
}

func (suite *FakeServerSuite) TestRunListWrongFilter() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	cmd.Flags().Bool("completed", true, "")

	// run the command
	args := []string{fakeGroupID}
	runList(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("", string(buf.Bytes()))
}

func (suite *FakeServerSuite) TestRunListBadGraphId() {
	// set up to run a command and capture output
	buf, cmd := setUpCommand()
	cmd.Flags().Bool("pending", true, "")

	// run the command
	args := []string{badGroupID}
	runList(&tcclient.Credentials{}, args, cmd.OutOrStdout(), cmd.Flags())

	suite.Equal("", string(buf.Bytes()))
}
