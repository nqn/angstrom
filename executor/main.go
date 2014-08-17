package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"
	"time"

	"code.google.com/p/goprotobuf/proto"
	"github.com/mesosphere/mesos-go/mesos"
)

type StatisticsInfo struct {
	ExecutorId string `json:"executor_id"`
	ExecutorName string `json:"executor_name"`
	FrameworkId string `json:"framework_id"`
	Source string `json:"source"`
	Statistics map[string]interface{}
}

type SampleRequest struct {
	Slave string `json:"slave"`
}


func sample(slave string) *[]StatisticsInfo {
	resp, err := http.Get("http://" + slave + "/monitor/statistics.json")
	if err != nil {
		log.Panic("Cannot get statistics from slave: '" + slave + "'")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response")
	}

	var monitor []StatisticsInfo
	err = json.Unmarshal(body, &monitor)
	if err != nil {
		log.Println("Could not parse monitor: " + err.Error())
		return nil
	}
	log.Println(monitor)
	return &monitor
}

func taskHandler(driver *mesos.ExecutorDriver, taskInfo mesos.TaskInfo) {
	for {
		var request SampleRequest
		err := json.Unmarshal(taskInfo.Data, &request)
		if err != nil {
			log.Println("Could not parse request: " + err.Error())
		} else {
			log.Println(request.Slave)

			samples := sample(request.Slave)
			if samples == nil {
				continue
			}

			// TODO(nnielsen): Return type should be (node_count, available, allocated, used).
			body, err := json.Marshal(samples)
			if err != nil {
				continue
			}
			log.Println("Sending framework data...")
			driver.SendFrameworkMessage(string(body))

			// TODO(nnielsen): Do local aggregation for finer grained samples.

			// TODO(nnielsen): Add terminal state.
			// driver.SendStatusUpdate(&mesos.TaskStatus{
			// 	TaskId:  taskInfo.TaskId,
			// 	State:   mesos.NewTaskState(mesos.TaskState_TASK_FINISHED),
			// 	Message: proto.String("Angstrom task YYY sampling completed"),
			// })
		}

		time.Sleep(1 * time.Second)
	}

}

func main() {
	driver := mesos.ExecutorDriver{
		Executor: &mesos.Executor{
			Registered: func(
				driver *mesos.ExecutorDriver,
				executor mesos.ExecutorInfo,
				framework mesos.FrameworkInfo,
				slave mesos.SlaveInfo) {
				fmt.Println("Angstrom executor registered!")
			},

			LaunchTask: func(driver *mesos.ExecutorDriver, taskInfo mesos.TaskInfo) {
				fmt.Println("Launch sample task!")

				// TODO(nnielsen): Launched tasks corresponds to resource samples in round robin fashion.
				driver.SendStatusUpdate(&mesos.TaskStatus{
					TaskId:  taskInfo.TaskId,
					State:   mesos.NewTaskState(mesos.TaskState_TASK_RUNNING),
					Message: proto.String("Angstrom task " + *taskInfo.TaskId.Value + " is sampling slave XXX"),
				})

				go taskHandler(driver, taskInfo)
			},
		},
	}

	driver.Init()
	defer driver.Destroy()

	driver.Run()
	driver.Join()
}
