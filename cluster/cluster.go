package cluster

import (
	"container/list"
	"sync"
	"github.com/nqn/angstrom/payload"
	"github.com/mesosphere/mesos-go/mesos"
	"net/http"
	"github.com/golang/glog"
	"io/ioutil"
	"encoding/json"
	"strings"
	"strconv"
	"time"
)

const archiveMaxSize = 2048

type Slave struct {
	Hostname string
	Port int
}

type Executor struct {
	Stat payload.StatisticsInfo
	UsedCpus float64
	UsedMemory float64
	UsedDisk float64
	LimitCpus float64
	LimitMemory float64
	LimitDisk float64
}

type Framework struct {
	Executors map[string]*Executor
}

// TODO(nnielsen): Create struct for cpu, memory and disk stats.
// TODO(nnielsen): Add custom types for percentages.
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
	CoverageCpus float64
	CoverageCpusPercent float64
	CoverageMemory float64
	CoverageMemoryPercent float64
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
	// TODO(nnielsen): Post-pone sample if insufficient usage data has been presented.
	// Difficult when slave list is being generated in same go and need to be available
	// up front.
	// TODO(nnielsen): Report sample accuracy / coverage.
	// Store sample
	if c.Sample != nil {
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

	// TODO(nnielsen): Used (and collected) data needs to be differentiated from static / snapshot like ones! This is a hack where we compute usage based on previous sample.
	previousSample := c.Sample

	c.Sample = &ClusterSample {
		Slaves: make(map[string]*Slave),
		Frameworks: make(map[string]*Framework),
	}

	sample := c.Sample

	resp, err := http.Get("http://" + c.Master + "/master/state.json")
	if err != nil {
		glog.Fatalf("Cannot get slave list from master '" + c.Master + "'")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Warning("Error reading response")
	}

	var target payload.MasterInfo
	err = json.Unmarshal(body, &target)
	if err != nil {
		glog.Fatalf("Error deserializing RenderResult from JSON: " + err.Error())
	}

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

		sample.AllocatedCpus += frameworkCPUs
		sample.AllocatedMemory += frameworkMemory
	}

	sample.UsedCpus = 0.0
	sample.UsedMemory = 0
	sample.UsedDisk = 0

	if previousSample != nil {
		for frameworkId, framework := range previousSample.Frameworks {
			if _, ok := activeFrameworks[frameworkId] ; ok {
				for _, executor := range framework.Executors {
					sample.UsedCpus += executor.UsedCpus
					sample.UsedMemory += executor.UsedMemory

					sample.CoverageCpus += executor.LimitCpus
					sample.CoverageMemory += executor.LimitMemory
					// We cannot record disk usage at the moment.
				}
			}
		}

		sample.CoverageCpusPercent = ((sample.CoverageCpus / sample.AllocatedCpus) * 100)
		sample.CoverageMemoryPercent = ((sample.CoverageMemory / sample.AllocatedMemory) * 100)

		// Compute slack.
		sample.SlackCpus = sample.AllocatedCpus - sample.UsedCpus
		sample.SlackMemory = sample.AllocatedMemory - sample.UsedMemory
		sample.SlackDisk = sample.AllocatedDisk - sample.UsedDisk
	}

	// Set timestamp.
	// Timestamp is in milliseconds.
	sample.Timestamp = time.Now().UnixNano() / 1e6

}

func (c *Cluster) AddSlaveSamples(slaveId mesos.SlaveID, target []payload.StatisticsInfo) {

	// TODO(nnielsen): Compute error.
	for _, stat := range target {
		frameworkId := stat.FrameworkId

		// TODO(nnielsen): Hack for now, we need to hang monitored slaves id off stats payload.
		executorId := stat.ExecutorId + ":" + slaveId.GetValue()

		var framework *Framework
		if f, ok := c.Sample.Frameworks[frameworkId] ; !ok {
			f = &Framework{
				Executors: make(map[string]*Executor),
			}
			c.Sample.Frameworks[frameworkId] = f
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

			// Compute new values since last sample.
			// TODO(nnielsen): Defer this to angstrom task and get finer resolution average, min, max and derivation.
			executor.LimitCpus = e.Stat.Statistics["cpus_limit"].(float64)
			executor.LimitMemory = e.Stat.Statistics["mem_limit_bytes"].(float64) / (1024 * 1024)

			totalTime := stat.Statistics["timestamp"].(float64) - e.Stat.Statistics["timestamp"].(float64)
			userTime := stat.Statistics["cpus_user_time_secs"].(float64) - e.Stat.Statistics["cpus_user_time_secs"].(float64)
			systemTime := stat.Statistics["cpus_system_time_secs"].(float64) - e.Stat.Statistics["cpus_system_time_secs"].(float64)

			executor.UsedCpus = (userTime + systemTime) / totalTime
			executor.UsedMemory = stat.Statistics["mem_rss_bytes"].(float64) / (1024 * 1024)
		}

		glog.V(2).Info(stat)

		executor.Stat = stat
	}
}
