package payload

type SampleRequest struct {
	Slave string `json:"slave"`
}

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

