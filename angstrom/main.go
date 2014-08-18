package main

import (
	"code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/mesosphere/mesos-go/mesos"
	"strconv"
	"time"
	"math"
	"net"
	"encoding/json"
	"strings"
	"container/list"
	"os"
	"path/filepath"
	"net/http"
	"github.com/golang/glog"
	"github.com/nqn/angstrom/payload"
	acluster "github.com/nqn/angstrom/cluster"
	"net/url"
)

const defaultPort = 9000
const defaultSampleLimit = 10

func main() {
	taskId := 0
	localExecutor, _ := executorPath()
	hostname, _ := os.Hostname()

	goPath := os.Getenv("GOPATH") + "/"

	master := flag.String("master", "localhost:5050", "Location of leading Mesos master")
	executorPath := flag.String("executor-uri", localExecutor, "Path to executor executable")
	address := flag.String("address", hostname, "Hostname to serve artifacts from")
	angstromPath := flag.String("angstrom-path", goPath + "src/github.com/nqn/angstrom", "Path to angstrom checkout")

	flag.Parse()

	// TODO(nnielsen): Hide in helper.
	// Determine address to listen on.
	interfaces, _ := net.Interfaces()
	for _, inter := range interfaces {
		if inter.Name == "lo" {
			continue
		}
		addr, err := inter.Addrs()
		if err == nil {
			network := addr[0].String()
			networkSplit := strings.Split(network, "/")
			address = &networkSplit[0]
			break
		}
	}

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
			time.Sleep(1 * time.Second)
		}
	}()

	slaves := list.New()
	for _, slave := range cluster.Sample.Slaves {
		slaveHostname := slave.Hostname + ":" + strconv.Itoa(slave.Port)
		slaves.PushBack(slaveHostname)
	}

	scheduleTask := func(offer mesos.Offer) *mesos.TaskInfo {
		slave := slaves.Front()
		if slave == nil {
			return nil
		}

		slaves.Remove(slave)

		// TODO(nnielsen): Map task -> monitored slave, for restart.

		return &mesos.TaskInfo{
			Name: proto.String("angstrom-task"),
			TaskId: &mesos.TaskID{
				Value: proto.String("angstrom-task-" + strconv.Itoa(taskId)),
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
						tasks = append(tasks, *task)
						driver.LaunchTasks(offer.Id, tasks)
					} else {
						driver.DeclineOffer(offer.Id)
					}

				}
			},

			FrameworkMessage: func(driver *mesos.SchedulerDriver, _executorId mesos.ExecutorID, slaveId mesos.SlaveID, data string) {
				// TODO(nnielsen): Move to cluster package.
				// TODO(nnielsen): Compute error.
				var target []payload.StatisticsInfo
				err := json.Unmarshal([]byte(data), &target)
				if err != nil {
					return
				}

				for _, stat := range target {
					frameworkId := stat.FrameworkId

					// TODO(nnielsen): Hack for now, we need to hang monitored slaves id off stats payload.
					executorId := stat.ExecutorId + ":" + slaveId.GetValue()

					var framework *acluster.Framework
					if f, ok := cluster.Sample.Frameworks[frameworkId] ; !ok {
						f = &acluster.Framework{
							Executors: make(map[string]*acluster.Executor),
						}
						cluster.Sample.Frameworks[frameworkId] = f
						framework = f
					} else {
						framework = f
					}

					var executor *acluster.Executor
					if e, ok := framework.Executors[executorId] ; !ok {
						e = &acluster.Executor {}
						framework.Executors[executorId] = e
						executor = e
					} else {
						executor = e

						// TODO(nnielsen): Save # samples.
						// Compute new values since last sample.

						limit := e.Stat.Statistics["cpus_limit"].(float64)
						_ = limit

						totalTime := stat.Statistics["timestamp"].(float64) - e.Stat.Statistics["timestamp"].(float64)
						userTime := stat.Statistics["cpus_user_time_secs"].(float64) - e.Stat.Statistics["cpus_user_time_secs"].(float64)
						systemTime := stat.Statistics["cpus_system_time_secs"].(float64) - e.Stat.Statistics["cpus_system_time_secs"].(float64)

						executor.Cpus = (userTime + systemTime) / totalTime
						executor.Memory = stat.Statistics["mem_rss_bytes"].(float64) / (1024 * 1024)
					}

					glog.V(2).Info(stat)

					executor.Stat = stat
				}
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

	// TODO(nnielsen): Separate HTTP handling into package.
	http.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		percentOf := func(a float64, b float64) float64 {
			return (a / b) * 100
		}

		glog.V(2).Infof("Request: %s", r.URL)
		glog.V(2).Infof("Total samples: %d", cluster.Archive.Len())

		var limit int64 = defaultSampleLimit
		var from int64 = 0
		var to int64 = math.MaxInt64

		query, err := url.ParseQuery(r.URL.RawQuery)
		if err == nil {
			integerField := func(field string) (int64, bool) {
				arr, ok := query[field] ; if ok {
					if len(arr) > 0 {
						val := arr[0]
						ival, err := strconv.ParseInt(val, 10, 64)
						if err == nil {
							return ival, true
						}
					}
				}
				return 0, false
			}

			if val, ok := integerField("limit") ; ok {
				limit = val
			}

			if val, ok := integerField("from") ; ok {
				from = val
			}

			if val, ok := integerField("to") ; ok {
				to = val
			}

		}

		c := make([]*payload.ClusterStateJson, 0)

		var sampleCount int64 = 0
		cluster.ArchiveLock.RLock()
		for e := cluster.Archive.Front(); e != nil; e = e.Next() {
			sample := e.Value.(*acluster.ClusterSample)

			// TODO(nnielsen): Separate into payload package.

			if (sample.Timestamp < from) || (sample.Timestamp > to) || (sampleCount >= limit) {
				continue
			}

			sampleCount++

			c = append(c, &payload.ClusterStateJson {
				TotalCpus: sample.Cpus,
				TotalMemory: sample.Memory,
				TotalDisk: sample.Disk,
				AllocatedCpus: sample.AllocatedCpus,
				AllocatedCpusPercent: percentOf(sample.AllocatedCpus, sample.Cpus),
				AllocatedMemory: sample.AllocatedMemory,
				AllocatedMemoryPercent: percentOf(sample.AllocatedMemory, sample.Memory),
				AllocatedDisk: sample.AllocatedDisk,
				AllocatedDiskPercent: percentOf(sample.AllocatedDisk, sample.Disk),
				UsedCpus: sample.UsedCpus,
				UsedCpusPercent: percentOf(sample.UsedCpus, sample.Cpus),
				UsedMemory: sample.UsedMemory,
				UsedMemoryPercent: percentOf(sample.UsedMemory, sample.Memory),
				UsedDisk: sample.UsedDisk,
				UsedDiskPercent: percentOf(sample.UsedDisk, sample.Disk),
				SlackCpus: sample.SlackCpus,
				SlackCpusPercent: percentOf(sample.SlackCpus, sample.Cpus),
				SlackMemory: sample.SlackMemory,
				SlackMemoryPercent: percentOf(sample.SlackMemory, sample.Memory),
				SlackDisk: sample.SlackDisk,
				SlackDiskPercent: percentOf(sample.SlackDisk, sample.Disk),
				Timestamp: sample.Timestamp,
			})
		}
		cluster.ArchiveLock.RUnlock()

		state := make(map[string][]*payload.ClusterStateJson)
		state["cluster"] = c

		body, err := json.Marshal(state)
		if err == nil {
			fmt.Fprintf(w, "%s", body)
		}
	})

	http.Handle("/", http.FileServer(http.Dir(*angstromPath + "/assets")))

	http.ListenAndServe(":" + strconv.Itoa(defaultPort), nil)
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
