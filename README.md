Ångström
========

Scalable, self-hosted metrics collection and global cluster state for Mesos.

![Ångström](http://cl.ly/image/3P2301053q1a/angstrom.png)


## Build instructions

```bash
$ go get github.com/nqn/angstrom/{scheduler,executor}
```

## API

http://_angstrom framework ip_:9000/resources

```json
{
  "cluster": [
    {
      "TotalCpus": 2,
      "TotalMemory": 5376,
      "TotalDisk": 9948,
      "AllocatedCpus": 0,
      "AllocatedCpusPercent": 0,
      "AllocatedMemory": 0,
      "AllocatedMemoryPercent": 0,
      "AllocatedDisk": 0,
      "AllocatedDiskPercent": 0,
      "UsedCpus": 0,
      "UsedCpusPercent": 0,
      "UsedMemory": 0,
      "UsedMemoryPercent": 0,
      "UsedDisk": 0,
      "UsedDiskPercent": 0,
      "SlackCpus": 0,
      "SlackCpusPercent": 0,
      "SlackMemory": 0,
      "SlackMemoryPercent": 0,
      "SlackDisk": 0,
      "SlackDiskPercent": 0,
      "Timestamp": 1408319101
    },
    {
      "TotalCpus": 2,
      "TotalMemory": 5376,
      "TotalDisk": 9948,
      "AllocatedCpus": 0,
      "AllocatedCpusPercent": 0,
      "AllocatedMemory": 0,
      "AllocatedMemoryPercent": 0,
      "AllocatedDisk": 0,
      "AllocatedDiskPercent": 0,
      "UsedCpus": 0,
      "UsedCpusPercent": 0,
      "UsedMemory": 0,
      "UsedMemoryPercent": 0,
      "UsedDisk": 0,
      "UsedDiskPercent": 0,
      "SlackCpus": 0,
      "SlackCpusPercent": 0,
      "SlackMemory": 0,
      "SlackMemoryPercent": 0,
      "SlackDisk": 0,
      "SlackDiskPercent": 0,
      "Timestamp": 1408319101
    }
  ]
}
```

