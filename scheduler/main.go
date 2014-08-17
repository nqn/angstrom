package main

import (
	"code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
	"github.com/mesosphere/mesos-go/mesos"
	"strconv"
	"time"
	"sync"
	"net"
	"encoding/json"
	"io/ioutil"
	"strings"
	"container/list"
	"os"
	"path/filepath"
	"net/http"
	"github.com/golang/glog"
)

const archiveMaxSize = 2048
const defaultPort = 9000


// TODO(nnielsen): Move payload structs to separate json package.

type MasterInfo struct {
	Slaves []SlaveInfo `json:"slaves"`
	Frameworks []FrameworkInfo `json:"frameworks"`
}

type SlaveInfo struct {
	Pid string `json:"pid"`
	Id string `json:"id"`
	Resources map[string]interface{} `json:"Resources"`
}

type FrameworkInfo struct {
	Id string `json:"id"`
	Resources map[string]interface{} `json:"Resources"`
}

type StatisticsInfo struct {
	ExecutorId string `json:"executor_id"`
	ExecutorName string `json:"executor_name"`
	FrameworkId string `json:"framework_id"`
	Source string `json:"source"`
	Statistics map[string]interface{}
}


// TODO(nnielsen): Move Slave, Executor, Framework, ClusterSample and Cluster to separate Cluster package.
type Slave struct {
	Hostname string
	Port int
}

type Executor struct {
	Stat StatisticsInfo
	Cpus float64
	Memory float64
	Disk float64
}

type Framework struct {
	Executors map[string]*Executor
}

// TODO(nnielsen): Create struct for cpu, memory and disk stats.
type ClusterSample struct {
	Cpus float64
	Memory float64
	Disk float64
	AllocatedCpus float64
	AllocatedMemory float64
	AllocatedDisk float64
	UsedCpus float64
	UsedMemory float64
	UsedDisk float64
	SlackCpus float64
	SlackMemory float64
	SlackDisk float64
	Slaves map[string]*Slave
	Frameworks map[string]*Framework
	Timestamp int64
}

type Cluster struct {
	Master string
	Sample *ClusterSample
	Archive list.List
	ArchiveLock *sync.RWMutex
}

func NewCluster(master string) *Cluster {
	return &Cluster {
		Master: master,
		ArchiveLock: &sync.RWMutex{},
	}
}

func (c *Cluster) Update() {
	if c.Sample != nil {
		// TODO(nnielsen): Guard archive list with mutex.
		// Currently not safe taken HTTP endpoint traverse archive.

		c.ArchiveLock.Lock()

		c.Archive.PushBack(c.Sample)

		// Only keep archiveMaxSize sampels around.
		archiveSize := c.Archive.Len()
		if archiveSize > archiveMaxSize {
			remove := archiveSize - archiveMaxSize
			for i := 0; i < remove; i++ {
				c.Archive.Remove(c.Archive.Front())
			}
		}

		c.ArchiveLock.Unlock()
	}

	c.Sample = &ClusterSample {
		Slaves: make(map[string]*Slave),
		Frameworks: make(map[string]*Framework),
	}

	resp, err := http.Get("http://" + c.Master + "/master/state.json")
	if err != nil {
		glog.Fatalf("Cannot get slave list from master '" + c.Master + "'")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warning("Error reading response")
	}

	var target MasterInfo
	err = json.Unmarshal(body, &target)
	if err != nil {
		glog.Fatalf("Error deserializing RenderResult from JSON: " + err.Error())
	}

	sample := c.Sample

	sample.Cpus = 0.0
	sample.Memory = 0
	sample.Disk = 0
	for _, slave := range target.Slaves {
		slaveCPUs := slave.Resources["cpus"].(float64)
		slaveMemory := slave.Resources["mem"].(float64)
		slaveDisk := slave.Resources["disk"].(float64)

		sample.Cpus += slaveCPUs
		sample.Memory += slaveMemory
		sample.Disk += slaveDisk

		pidSplit := strings.Split(slave.Pid, "@")
		hostPort := pidSplit[1]
		hostSplit := strings.Split(hostPort, ":")

		hostname := hostSplit[0]
		port, err := strconv.Atoi(hostSplit[1])
		if err == nil {
			sample.Slaves[slave.Id] = &Slave { Hostname: hostname, Port: port }
		}
	}

	sample.AllocatedCpus = 0.0
	sample.AllocatedMemory = 0
	sample.AllocatedDisk = 0
	activeFrameworks := make(map[string]struct{})
	for _, framework := range target.Frameworks {
		activeFrameworks[framework.Id] = struct{}{}

		frameworkCPUs := framework.Resources["cpus"].(float64)
		frameworkMemory := framework.Resources["mem"].(float64)
		frameworkDisk := framework.Resources["disk"].(float64)

		sample.AllocatedCpus += frameworkCPUs
		sample.AllocatedMemory += frameworkMemory
		sample.AllocatedDisk += frameworkDisk
	}

	sample.UsedCpus = 0.0
	sample.UsedMemory = 0
	sample.UsedDisk = 0
	for frameworkId, framework := range sample.Frameworks {
		if _, ok := activeFrameworks[frameworkId] ; ! ok {
			glog.V(2).Infof("Removing inactive framework: " + frameworkId)
			delete(sample.Frameworks, frameworkId)
		} else {
			for _, executor := range framework.Executors {
				sample.UsedCpus += executor.Cpus
				sample.UsedMemory += executor.Memory
			}
		}
	}

	// Compute slack.
	sample.SlackCpus = sample.AllocatedCpus - sample.UsedCpus
	sample.SlackMemory = sample.AllocatedMemory - sample.UsedMemory
	sample.SlackDisk = sample.AllocatedDisk - sample.UsedDisk

	// Set timestamp.
	sample.Timestamp = time.Now().Unix()
}

type ClusterStateJson struct {
	TotalCpus float64
	TotalMemory float64
	TotalDisk float64

	AllocatedCpus float64
	AllocatedCpusPercent float64
	AllocatedMemory float64
	AllocatedMemoryPercent float64
	AllocatedDisk float64
	AllocatedDiskPercent float64

	UsedCpus float64
	UsedCpusPercent float64
	UsedMemory float64
	UsedMemoryPercent float64
	UsedDisk float64
	UsedDiskPercent float64

	SlackCpus float64
	SlackCpusPercent float64
	SlackMemory float64
	SlackMemoryPercent float64
	SlackDisk float64
	SlackDiskPercent float64

	Timestamp int64
}

func main() {
	taskId := 0
	localExecutor, _ := executorPath()
	hostname, _ := os.Hostname()

	master := flag.String("master", "localhost:5050", "Location of leading Mesos master")
	executorPath := flag.String("executor-uri", localExecutor, "Path to executor executable")
	address := flag.String("address", hostname, "Hostname to serve artifacts from")

	flag.Parse()

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
			Value: proto.String("./executor"),
			Uris: []*mesos.CommandInfo_URI{
				&mesos.CommandInfo_URI{Value: &executorURI, Executable: &executable},
			},
		},
		Name:   proto.String("Test Executor (Go)"),
		Source: proto.String("go_test"),
	}

	cluster := NewCluster(*master)

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
				// TODO(nnielsen): Compute error.
				var target []StatisticsInfo
				err := json.Unmarshal([]byte(data), &target)
				if err != nil {
					return
				}

				for _, stat := range target {
					frameworkId := stat.FrameworkId

					// TODO(nnielsen): Hack for now, we need to hang monitored slaves id off stats payload.
					executorId := stat.ExecutorId + ":" + slaveId.GetValue()

					var framework *Framework
					if f, ok := cluster.Sample.Frameworks[frameworkId] ; !ok {
						f = &Framework{
							Executors: make(map[string]*Executor),
						}
						cluster.Sample.Frameworks[frameworkId] = f
						framework = f
					} else {
						framework = f
					}

					var executor *Executor
					if e, ok := framework.Executors[executorId] ; !ok {
						e = &Executor {}
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

	http.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		percentOf := func(a float64, b float64) float64 {
			return (a / b) * 100
		}

		glog.V(2).Infof("Request: %s", r.URL)
		glog.V(2).Infof("Total samples: %d", cluster.Archive.Len())

		// TODO(nnielsen): Support 'from' field, specifying samples in time range to serve.
		// TODO(nnielsen): Support 'limit' field. Default limit should be 1.
		c := make([]*ClusterStateJson, 0)

		cluster.ArchiveLock.RLock()
		for e := cluster.Archive.Front(); e != nil; e = e.Next() {
			sample := e.Value.(*ClusterSample)
			c = append(c, &ClusterStateJson {
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

		state := make(map[string][]*ClusterStateJson)
		state["cluster"] = c

		body, err := json.Marshal(state)
		if err == nil {
			fmt.Fprintf(w, "%s", body)
		}
	})

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
