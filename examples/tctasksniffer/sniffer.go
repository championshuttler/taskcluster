// Package tctasksniffer provides an example application that listens to Pulse
// for task definition messages and task running messages, and reports on them.
package main

import (
	"errors"
	"fmt"
	"github.com/petemoore/pulse-go/pulse"
	"github.com/petemoore/taskcluster-client-go/queueevents"
	"github.com/streadway/amqp"
)

func main() {
	// Passing all empty strings:
	// empty user => use PULSE_USERNAME env var
	// empty password => use PULSE_PASSWORD env var
	// empty url => connect to production
	conn := pulse.NewConnection("", "", "")
	conn.Consume(
		"taskprocessing", // queue name
		func(message interface{}, delivery amqp.Delivery) { // callback function to pass messages to
			switch t := message.(type) {
			case *queueevents.TaskDefinedMessage:
				fmt.Println("Task " + t.Status.TaskId + " defined")
				fmt.Println(string(delivery.Body))
			case *queueevents.TaskRunningMessage:
				fmt.Println("Task " + t.Status.TaskId + " running, (taken until " + t.TakenUntil + " by worker " + t.WorkerId + ")")
			default:
				panic(errors.New(fmt.Sprintf("Unrecognised message type %T!", t)))
			}
			fmt.Println("===========")
			delivery.Ack(false) // acknowledge message *after* processing
		},
		1,     // prefetch 1 message at a time
		false, // don't auto-acknowledge messages
		// queueevents.TaskDefined{WorkerType: "gaia"},
		// queueevents.TaskRunning{ProvisionerId: "aws-provisioner"})
		queueevents.TaskDefined{},
		queueevents.TaskRunning{})
	conn.Consume( // a second workflow to manage concurrently
		"", // empty name implies anonymous queue
		func(message interface{}, delivery amqp.Delivery) { // simpler callback than before
			fmt.Println("A buildbot message was received")
			fmt.Println("===========")
		},
		1,    // prefetch
		true, // auto acknowledge, so no need to call delivery.Ack
		pulse.Bind( // routing key and exchange to get messages from
			"#", // get *all* normalized buildbot messages
			"exchange/build/normalized"))
	// wait forever
	forever := make(chan bool)
	<-forever
}
