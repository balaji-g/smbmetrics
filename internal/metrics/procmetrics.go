package metrics

import (
    "fmt"
    "github.com/shirou/gopsutil/v4/net"
    "github.com/shirou/gopsutil/v4/process"
    "log"
    "strconv"
    "strings"
)

type PsUtilPidData struct {
    PID                       int64
    CpuUsagePercent           float64
    VirtualMemoryUsageBytes   uint64
    VirtualMemoryUsagePercent float64
    IoCounterReadCount        uint64
    IoCounterReadBytes        uint64
    IoCounterWriteCount       uint64
    IoCounterWriteBytes       uint64
    BytesSent                 uint64
    BytesRecv                 uint64
    OpenFilesCount            uint64
    ThreadCount               uint64
}

func (pidData PsUtilPidData) String() string {
    return fmt.Sprintf("PID: %d; CPU Usage Percent: %f; VM Usage Bytes: %d; VM Usage Percent: %f; IO Read Count: %d; IO Read Bytes: %d; IO Write Count: %d; IO Write Bytes: %d; Open File Count: %d; Thread Count: %d",
        pidData.PID, pidData.CpuUsagePercent, pidData.VirtualMemoryUsageBytes, pidData.VirtualMemoryUsagePercent,
        pidData.IoCounterReadCount, pidData.IoCounterReadBytes, pidData.IoCounterWriteCount, pidData.IoCounterWriteBytes,
        pidData.BytesSent, pidData.BytesRecv,
        pidData.OpenFilesCount, pidData.ThreadCount)
}
func GetPsUtilPidData() ([]PsUtilPidData, error) {
    pids, err := getPidList()
    if err != nil {
        log.Printf("getPidList() failed: %v", err)
        return nil, err
    }

    ret := []PsUtilPidData{}
    for _, pid := range pids {
        proc, err := process.NewProcess(pid)
        if err != nil {
            log.Printf("NewProcess failed %v:%v", err, pid)
            continue
        }

        cpuPercent, err := proc.CPUPercent()
        if err != nil {
            log.Printf("CPUPercent failed %v:%v", err, pid)
            continue
        }
        vmBytes, err := proc.MemoryInfo()
        if err != nil {
            log.Printf("MemoryInfo failed %v:%v", err, pid)
            continue
        }
        vmPercent, err := proc.MemoryPercent()
        if err != nil {
            log.Printf("MemoryPercent failed %v:%v", err, pid)
            continue
        }
        ioCounters, err := proc.IOCounters()
        if err != nil {
            log.Printf("IOCounters failed %v:%v", err, pid)
            //continue
        }
        if ioCounters == nil {
            ioCounters = &process.IOCountersStat{}
        }
        netCounters, err := net.IOCounters(false)
        if err != nil {
            log.Printf("NetCounters failed %v:%v", err, pid)
        }
        openFileStats, err := proc.OpenFiles()
        if err != nil {
            log.Printf("OpenFiles failed %v:%v", err, pid)
            continue
        }
        threadStats, err := proc.Threads()
        if err != nil {
            log.Printf("Threads failed %v:%v", err, pid)
            continue
        }

        entry := PsUtilPidData{
            int64(pid),
            cpuPercent,
            vmBytes.VMS,
            float64(vmPercent),
            ioCounters.ReadCount,
            ioCounters.ReadBytes,
            ioCounters.WriteCount,
            ioCounters.WriteBytes,
            netCounters[0].BytesSent,
            netCounters[0].BytesRecv,
            uint64(len(openFileStats)),
            uint64(len(threadStats)),
        }
        ret = append(ret, entry)
    }
    return ret, nil
}

func getPidList() ([]int32, error) {
    pidListInByte, err := executeCommand("pgrep", "smbd")
    if err != nil {
        return nil, err
    }

    pidListLines := strings.Split(string(pidListInByte), "\n")
    var pidList []int32
    for _, line := range pidListLines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        pid, errConv := strconv.ParseInt(line, 10, 32)
        if errConv != nil {
            return nil, errConv
        } else {
            pidList = append(pidList, int32(pid))
        }
    }
    return pidList, nil
}
