package main

import (
	"code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/mesosphere/mesos-go/mesos"
	"strconv"
	"time"
	"net"
	"encoding/json"
	"strings"
	"container/list"
	"os"
	"path/filepath"
	"net/http"
	"github.com/golang/glog"
	"github.com/nqn/angstrom/payload"
	"github.com/nqn/angstrom/endpoints"
	acluster "github.com/nqn/angstrom/cluster"
)

const defaultPort = 9000
const updateInterval = 1 * time.Second

func main() {
	taskId := 0
	localExecutor, _ := executorPath()

	goPath := os.Getenv("GOPATH") + "/"

	master := flag.String("master", "localhost:5050", "Location of leading Mesos master")
	executorPath := flag.String("executor-uri", localExecutor, "Path to executor executable")
	address := flag.String("address", "localhost", "Hostname to serve artifacts from")
	angstromPath := flag.String("angstrom-path", goPath + "src/github.com/nqn/angstrom", "Path to angstrom checkout")

	flag.Parse()

	serveExecutorArtifact := func(path string) string {
		serveFile := func(pattern string, filename string) {
			http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, filename)
			})
		}

		// Create base path (http://foobar:5000/<base>)
		pathSplit := strings.Split(path, "/")
		var base string
		if len(pathSplit) > 0 {
			base = pathSplit[len(pathSplit)-1]
		} else {
			base = path
		}
		serveFile("/"+base, path)

		hostURI := fmt.Sprintf("http://%s:%d/%s", *address, defaultPort, base)

		glog.V(2).Infof("Serving '%s'\n", hostURI)

		return hostURI
	}

	executorURI := serveExecutorArtifact(*executorPath)
	executable := true

	executor := &mesos.ExecutorInfo{
		ExecutorId: &mesos.ExecutorID{Value: proto.String("default")},
		Command: &mesos.CommandInfo{
			Value: proto.String("./angstrom-executor"),
			Uris: []*mesos.CommandInfo_URI{
				&mesos.CommandInfo_URI{Value: &executorURI, Executable: &executable},
			},
		},
		Name:   proto.String("Angstrom Executor"),
		Source: proto.String("angstrom"),
	}

	cluster := acluster.NewCluster(*master)

	cluster.Update()

	// Keep updating cluster state
	go func() {
		for {
			cluster.Update()
			time.Sleep(updateInterval)
		}
	}()

	slaves := list.New()
	for _, slave := range cluster.Sample.Slaves {
		slaveHostname := slave.Hostname + ":" + strconv.Itoa(slave.Port)
		slaves.PushBack(slaveHostname)
	}

	// TODO(nnielsen): Move this and callbacks to dedicated scheduler package / struct.
	taskToSlave := make(map[string]string)

	scheduleTask := func(offer mesos.Offer) *mesos.TaskInfo {
		slave := slaves.Front()
		if slave == nil {
			return nil
		}

                glog.V(2).Infof("Scheduling slave '%s'", slave.Value.(string))

		slaves.Remove(slave)

		task := "angstrom-task-" + strconv.Itoa(taskId)
		taskToSlave[task] = slave.Value.(string)

		return &mesos.TaskInfo{
			Name: proto.String("angstrom-task"),
			TaskId: &mesos.TaskID{
				Value: proto.String(task),
			},
			SlaveId:  offer.SlaveId,
			Executor: executor,
			Data: []byte("{\"slave\": \"" + slave.Value.(string) + "\"}"),
			Resources: []*mesos.Resource{
				mesos.ScalarResource("cpus", 0.5),
				mesos.ScalarResource("mem", 32),
			},
		}
	}

	driver := mesos.SchedulerDriver{
		Master: *master,
		Framework: mesos.FrameworkInfo{
			Name: proto.String("Angstrom metrics"),
			User: proto.String(""),
		},

		Scheduler: &mesos.Scheduler{
			ResourceOffers: func(driver *mesos.SchedulerDriver, offers []mesos.Offer) {
				for _, offer := range offers {
					taskId++

					tasks := make([]mesos.TaskInfo, 0)

					task := scheduleTask(offer) ; if task != nil {
                                                glog.V(2).Infof("Launching on offer %v", offer.Id)

						tasks = append(tasks, *task)
						driver.LaunchTasks(offer.Id, tasks)
					} else {
                                                glog.V(2).Infof("Declining offer %v", offer.Id)
						driver.DeclineOffer(offer.Id)
					}

				}
			},

			FrameworkMessage: func(driver *mesos.SchedulerDriver, _executorId mesos.ExecutorID, slaveId mesos.SlaveID, data string) {
				var target []payload.StatisticsInfo
				err := json.Unmarshal([]byte(data), &target)
				if err != nil {
					return
				}

				cluster.AddSlaveSamples(slaveId, target)
			},

			StatusUpdate: func(driver *mesos.SchedulerDriver, status mesos.TaskStatus) {
				// TODO(nnielsen): Readd slave task to queue in case of any terminal state.
				if *status.State == mesos.TaskState_TASK_RUNNING {
				} else if *status.State == mesos.TaskState_TASK_FINISHED {
				}
			},
		},
	}

	driver.Init()
	defer driver.Destroy()
	driver.Start()

	endpoints.Initialize(defaultPort, *angstromPath, cluster)

        glog.V(2).Infof("Waiting for threads to join")

	driver.Join()
}

func executorPath() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}

	path := dir + "/angstrom-executor"
	return path, nil
}
