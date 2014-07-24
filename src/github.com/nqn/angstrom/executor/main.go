package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"

	"code.google.com/p/goprotobuf/proto"
	"github.com/mesosphere/mesos-go/mesos"
)

type MonitorInfo struct {
	StaticsArray []StatisticsInfo
}

type StatisticsInfo struct {
	ExecutorId string `json:"executor_id"`
	ExecutorName string `json:"executor_name"`
	FrameworkId string `json:"framework_id"`
	Source string `json:"source"`
	Statistics map[string]interface{}
}

type Statistics struct {
	// CpusLimit int `json:"cpus_limit"`
	// CpusNrPeriods int `json:"cpus_nr_periods"`
	// CpusNrThrottled int
	// CpusSystemTimeSecs float64 `json:"cpus_system_time_secs"`
	// CpusThrottledTimeSecs float64
	// CpusUserTimeSecs float64`json:"cpus_user_time_secs"`
	// MemAnonBytes int
	// MemFileBytes int
	// MemLimitBytes int
	// MemMappedFileBytes int
	// MemRss_bytes int
	// Timestamp float64 `json:"timestamp"`
}

type SampleRequest struct {
	Slave string `json:"slave"`
}

func main() {
	driver := mesos.ExecutorDriver{
		Executor: &mesos.Executor{
			Registered: func(
				driver *mesos.ExecutorDriver,
				executor mesos.ExecutorInfo,
				framework mesos.FrameworkInfo,
				slave mesos.SlaveInfo) {
				fmt.Println("Executor registered!")
			},

			LaunchTask: func(driver *mesos.ExecutorDriver, taskInfo mesos.TaskInfo) {
				fmt.Println("Launch task!")

				var request SampleRequest
				err := json.Unmarshal(taskInfo.Data, &request)
				if err != nil {
					log.Println("Could not parse request: " + err.Error())
				} else {
					log.Println(request.Slave)
					resp, err := http.Get("http://" + request.Slave + "/monitor/statistics.json")
					if err != nil {
						log.Panic("Cannot get statistics from slave: '" + request.Slave + "'")
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
					}

					log.Println(monitor)
				}

				// TODO(nnielsen): Launched tasks corresponds to resource samples in round robin fashion.
				driver.SendStatusUpdate(&mesos.TaskStatus{
					TaskId:  taskInfo.TaskId,
					State:   mesos.NewTaskState(mesos.TaskState_TASK_RUNNING),
					Message: proto.String("Go task is running!"),
				})

				driver.SendStatusUpdate(&mesos.TaskStatus{
					TaskId:  taskInfo.TaskId,
					State:   mesos.NewTaskState(mesos.TaskState_TASK_FINISHED),
					Message: proto.String("Go task is done!"),
				})
			},
		},
	}

	driver.Init()
	defer driver.Destroy()

	driver.Run()
}
