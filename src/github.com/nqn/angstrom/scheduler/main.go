package main

import (
	"code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/mesosphere/mesos-go/mesos"
	"path/filepath"
	"strconv"

//	"os"
)

// TODO(nnielsen): Reintroduce custom executor.
func main() {
	taskId := 0
	exit := make(chan bool)
	localExecutor, _ := executorPath()

	master := flag.String("master", "localhost:5050", "Location of leading Mesos master")
	// executorUri := flag.String("executor-uri", localExecutor, "URI of executor executable")
	flag.Parse()

	// executor := &mesos.ExecutorInfo{
	// 	ExecutorId: &mesos.ExecutorID{Value: proto.String("default")},
	// 	Command: &mesos.CommandInfo{
	// 		Value: proto.String("./example_executor"),
	// 		Uris: []*mesos.CommandInfo_URI{
	// 			&mesos.CommandInfo_URI{Value: executorUri},
	// 		},
	// 	},
	// 	Name:   proto.String("Test Executor (Go)"),
	// 	Source: proto.String("go_test"),
	// }

	driver := mesos.SchedulerDriver{
		Master: *master,
		Framework: mesos.FrameworkInfo{
			Name: proto.String("Angstrom metrics framework"),
			User: proto.String(""),
		},

		Scheduler: &mesos.Scheduler{
			ResourceOffers: func(driver *mesos.SchedulerDriver, offers []mesos.Offer) {
				for _, offer := range offers {
					taskId++

					tasks := []mesos.TaskInfo{
						mesos.TaskInfo{
							Name: proto.String("angstrom-task"),
							TaskId: &mesos.TaskID{
								Value: proto.String("angstrom-task-" + strconv.Itoa(taskId)),
							},
							SlaveId:  offer.SlaveId,
							Executor: executor,
							Resources: []*mesos.Resource{
								mesos.ScalarResource("cpus", 1),
								mesos.ScalarResource("mem", 512),
							},
						},
					}

					driver.LaunchTasks(offer.Id, tasks)
				}
			},

			StatusUpdate: func(driver *mesos.SchedulerDriver, status mesos.TaskStatus) {
				fmt.Println("Received task status: " + *status.Message)

				if *status.State == mesos.TaskState_TASK_FINISHED {
				}
			},
		},
	}

	driver.Init()
	defer driver.Destroy()

	driver.Start()
	driver.Join()
}

// func executorPath() (string, error) {
// 	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
// 	if err != nil {
// 		return "", err
// 	}
//
// 	path := dir + "/example_executor"
// 	return path, nil
// }
