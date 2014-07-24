package main

import (
	"code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/mesosphere/mesos-go/mesos"
	"strconv"
	"net/http"
	"log"
	"encoding/json"
	"io/ioutil"
	"strings"
	"container/list"
	"os"
	"path/filepath"
)

type MasterInfo struct {
	Slaves []SlaveInfo `json:"slaves"`
}

type SlaveInfo struct {
	Pid string `json:"pid"`
}

// TODO(nnielsen): Reintroduce custom executor.
func main() {
	taskId := 0
	slaves := list.New()
	localExecutor, _ := executorPath()

	master := flag.String("master", "localhost:5050", "Location of leading Mesos master")
	executorUri := flag.String("executor-uri", localExecutor, "URI of executor executable")
	flag.Parse()

	executor := &mesos.ExecutorInfo{
		ExecutorId: &mesos.ExecutorID{Value: proto.String("default")},
		Command: &mesos.CommandInfo{
			Value: proto.String("./executor"),
			Uris: []*mesos.CommandInfo_URI{
				&mesos.CommandInfo_URI{Value: executorUri},
			},
		},
		Name:   proto.String("Test Executor (Go)"),
		Source: proto.String("go_test"),
	}

	resp, err := http.Get("http://" + *master + "/master/state.json")
	if err != nil {
		log.Panic("Cannot get slave list from master '" + *master + "'")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response")
	}

	var target MasterInfo
	err = json.Unmarshal(body, &target)
	if err != nil {
		log.Panic("Error deserializing RenderResult from JSON: " + err.Error())
	}

	for _, slave := range target.Slaves {
		pidSplit := strings.Split(slave.Pid, "@")
		slaves.PushBack(pidSplit[1])
	}
	// TODO(nnielsen): Partition node list and pack in angstrom tasks.

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
							Data: []byte("{\"slave\": \"localhost:5051\"}"),
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
				fmt.Println("Received task status")

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

func executorPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	path := dir + "/executor"
	return path, nil
}
