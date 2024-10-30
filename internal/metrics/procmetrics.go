package metrics

import (
    "fmt"
    "github.com/shirou/gopsutil/v3/process"
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
    OpenFilesCount            uint64
    ThreadCount               uint64
}

func (pidData PsUtilPidData) String() string {
    return fmt.Sprintf("PID: %d; CPU Usage Percent: %f; VM Usage Bytes: %d; VM Usage Percent: %f; IO Read Count: %d; IO Read Bytes: %d; IO Write Count: %d; IO Write Bytes: %d; Open File Count: %d; Thread Count: %d",
        pidData.PID, pidData.CpuUsagePercent, pidData.VirtualMemoryUsageBytes, pidData.VirtualMemoryUsagePercent,
        pidData.IoCounterReadCount, pidData.IoCounterReadBytes, pidData.IoCounterWriteCount, pidData.IoCounterWriteBytes,
        pidData.OpenFilesCount, pidData.ThreadCount)
}
func GetPsUtilPidData() ([]PsUtilPidData, error) {
    pids, err := getPidList()
    if err != nil {
        return nil, err
    }

    ret := []PsUtilPidData{}
    for _, pid := range pids {
        proc, errProc := process.NewProcess(pid)
        if errProc != nil {
            return nil, errProc
        }

        cpuPercent, errPer := proc.CPUPercent()
        if errPer != nil {
            return nil, errPer
        }
        vmBytes, errVmBytes := proc.MemoryInfo()
        if errVmBytes != nil {
            return nil, errVmBytes
        }
        vmPercent, errVmPercent := proc.MemoryPercent()
        if errVmPercent != nil {
            return nil, errVmPercent
        }
        ioCounters, errIoCounters := proc.IOCounters()
        if errIoCounters != nil {
            return nil, errIoCounters
        }
        openFileStats, errOpenFileStats := proc.OpenFiles()
        if errOpenFileStats != nil {
            return nil, errOpenFileStats
        }
        threadStats, errThreadStats := proc.Threads()
        if errThreadStats != nil {
            return nil, errThreadStats
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
