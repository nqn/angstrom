Ångström
========

Scalable, self-hosted metrics collection and global cluster state for Mesos.

![Ångström](http://cl.ly/image/2F0F1g0S1a0Z/angstrom-2.png)


## Build instructions

```bash
$ go get github.com/nqn/angstrom/{angstrom,angstrom-executor}
$ $(GOPATH)/bin/angstrom -master <mesos-master-ip:port>
```

## WebUI

`http://<angstrom framework ip>:9000/`

## API

`http://<angstrom framework ip>:9000/resources(?limit=10&from=timestamp&to=timestamp)`

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

