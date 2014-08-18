package endpoints

import (
	acluster "github.com/nqn/angstrom/cluster"
	"github.com/nqn/angstrom/payload"
	"math"
	"net/url"
	"net/http"
	"strconv"
	"github.com/golang/glog"
	"encoding/json"
	"fmt"
)

const defaultSampleLimit = 10
const defaultMinCoverage = 80.0

func Initialize(port int, angstromPath string, cluster *acluster.Cluster) {
	// Serve Web UI.
	http.Handle("/", http.FileServer(http.Dir(angstromPath + "/assets")))

	// TODO(nnielsen): Separate HTTP handling into resources.go.
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

			// TODO(nnielsen): These samples should not have been collected at all.
			if (sample.CoverageCpusPercent < defaultMinCoverage) || (sample.CoverageMemoryPercent < defaultMinCoverage) {
				continue
			}

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

	http.ListenAndServe(":" + strconv.Itoa(port), nil)
}
