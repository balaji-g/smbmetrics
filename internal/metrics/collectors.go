// SPDX-License-Identifier: Apache-2.0

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    collectorsNamespace = "smb"
)

func (sme *smbMetricsExporter) register() error {
    cols := []prometheus.Collector{
        sme.newSMBVersionsCollector(),
        sme.newSMBActivityCollector(),
        sme.newSMBSharesCollector(),
        sme.newSMBProcessCollector(),
    }
    for _, c := range cols {
        if err := sme.reg.Register(c); err != nil {
            sme.log.Error(err, "failed to register collector")
            return err
        }
    }
    return nil
}

type smbCollector struct {
    // nolint:structcheck
    sme *smbMetricsExporter
    dsc []*prometheus.Desc
}

func (col *smbCollector) Describe(ch chan<- *prometheus.Desc) {
    for _, d := range col.dsc {
        ch <- d
    }
}

type smbVersionsCollector struct {
    smbCollector
    clnt *kclient
}

func (col *smbVersionsCollector) Collect(ch chan<- prometheus.Metric) {
    status := 0
    vers, err := ResolveVersions(col.clnt)
    if err != nil {
        status = 1
    }
    ch <- prometheus.MustNewConstMetric(
        col.dsc[0],
        prometheus.GaugeValue,
        float64(status),
        vers.Version,
        vers.CommitID,
        vers.SambaImage,
        vers.SambaVersion,
        vers.CtdbVersion,
    )
}

func (sme *smbMetricsExporter) newSMBVersionsCollector() prometheus.Collector {
    col := &smbVersionsCollector{}
    col.sme = sme
    col.clnt, _ = newKClient()
    col.dsc = []*prometheus.Desc{
        prometheus.NewDesc(
            collectorName("metrics", "status"),
            "Current metrics-collector status versions",
            []string{
                "version",
                "commitid",
                "sambaimage",
                "sambavers",
                "ctdbvers",
            }, nil),
    }
    return col
}

type smbActivityCollector struct {
    smbCollector
}

func (col *smbActivityCollector) Collect(ch chan<- prometheus.Metric) {
    totalSessions := 0
    totalTreeCons := 0
    totalConnectedUsers := 0
    totalOpenFiles := 0
    totalOpenFilesAccessRW := 0
    smbInfo, err := NewUpdatedSMBInfo()
    if err == nil {
        totalSessions = smbInfo.TotalSessions()
        totalTreeCons = smbInfo.TotalTreeCons()
        totalConnectedUsers = smbInfo.TotalConnectedUsers()
        totalOpenFiles = smbInfo.TotalOpenFiles()
        totalOpenFilesAccessRW = smbInfo.TotalOpenFilesAccessRW()
    }
    ch <- prometheus.MustNewConstMetric(col.dsc[0],
        prometheus.GaugeValue, float64(totalSessions))

    ch <- prometheus.MustNewConstMetric(col.dsc[1],
        prometheus.GaugeValue, float64(totalTreeCons))

    ch <- prometheus.MustNewConstMetric(col.dsc[2],
        prometheus.GaugeValue, float64(totalConnectedUsers))

    ch <- prometheus.MustNewConstMetric(col.dsc[3],
        prometheus.GaugeValue, float64(totalOpenFiles))

    ch <- prometheus.MustNewConstMetric(col.dsc[4],
        prometheus.GaugeValue, float64(totalOpenFilesAccessRW))
}

func (sme *smbMetricsExporter) newSMBActivityCollector() prometheus.Collector {
    col := &smbActivityCollector{}
    col.sme = sme
    col.dsc = []*prometheus.Desc{
        prometheus.NewDesc(
            collectorName("sessions", "total"),
            "Number of currently active SMB sessions",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("tcon", "total"),
            "Number of currently active SMB tree-connections",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("users", "total"),
            "Number of currently active SMB users",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("openfiles", "total"),
            "Number of currently open files",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("openfiles", "access_rw"),
            "Number of open files with read-write access mode",
            []string{}, nil),
    }
    return col
}

type smbSharesCollector struct {
    smbCollector
}

func (col *smbSharesCollector) Collect(ch chan<- prometheus.Metric) {
    smbInfo, _ := NewUpdatedSMBInfo()
    serviceToMachine := smbInfo.MapServiceToMachines()
    for service, machines := range serviceToMachine {
        ch <- prometheus.MustNewConstMetric(col.dsc[0],
            prometheus.GaugeValue,
            float64(len(machines)),
            service)
    }
    machineToServices := smbInfo.MapMachineToServies()
    for machine, services := range machineToServices {
        ch <- prometheus.MustNewConstMetric(col.dsc[1],
            prometheus.GaugeValue,
            float64(len(services)),
            machine)
    }
}

func (sme *smbMetricsExporter) newSMBSharesCollector() prometheus.Collector {
    col := &smbSharesCollector{}
    col.sme = sme
    col.dsc = []*prometheus.Desc{
        prometheus.NewDesc(
            collectorName("share", "activity"),
            "Number of remote machines currently using a share",
            []string{"service"}, nil),

        prometheus.NewDesc(
            collectorName("share", "byremote"),
            "Number of shares served for remote machine",
            []string{"machine"}, nil),
    }
    return col
}

type smbProcessCollector struct {
    smbCollector
}

func (col *smbProcessCollector) Collect(ch chan<- prometheus.Metric) {

    serverUp := 0
    vmBytes := uint64(0)
    netBytesSent := uint64(0)
    netBytesRecv := uint64(0)
    ioReadCount := uint64(0)
    ioReadBytes := uint64(0)
    ioWriteCount := uint64(0)
    ioWriteBytes := uint64(0)

    psInfo, err := GetPsUtilPidData()
    if err == nil {
        serverUp = 1
        for _, ps := range psInfo {
            vmBytes += ps.VirtualMemoryUsageBytes
            netBytesRecv += ps.BytesRecv
            netBytesSent += ps.BytesSent
            ioReadCount += ps.IoCounterReadCount
            ioReadBytes += ps.IoCounterReadBytes
            ioWriteBytes += ps.IoCounterWriteBytes
            ioWriteCount += ps.IoCounterWriteCount
        }
    }
    ch <- prometheus.MustNewConstMetric(col.dsc[0],
        prometheus.GaugeValue, float64(serverUp))
    ch <- prometheus.MustNewConstMetric(col.dsc[1],
        prometheus.GaugeValue, float64(netBytesRecv))
    ch <- prometheus.MustNewConstMetric(col.dsc[2],
        prometheus.GaugeValue, float64(vmBytes))
    ch <- prometheus.MustNewConstMetric(col.dsc[3],
        prometheus.GaugeValue, float64(netBytesSent))
    ch <- prometheus.MustNewConstMetric(col.dsc[4],
        prometheus.GaugeValue, float64(ioReadCount))
    ch <- prometheus.MustNewConstMetric(col.dsc[5],
        prometheus.GaugeValue, float64(ioReadBytes))
    ch <- prometheus.MustNewConstMetric(col.dsc[6],
        prometheus.GaugeValue, float64(ioWriteCount))
    ch <- prometheus.MustNewConstMetric(col.dsc[7],
        prometheus.GaugeValue, float64(ioWriteBytes))
}

func (sme *smbMetricsExporter) newSMBProcessCollector() prometheus.Collector {
    col := &smbProcessCollector{}
    col.sme = sme
    col.dsc = []*prometheus.Desc{
        prometheus.NewDesc(
            collectorName("smbd_up", "status"),
            "SMBD Status",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("netbytes", "sent"),
            "CPU Usage Percent",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("vmusagebytes", "total"),
            "Virtual Memory Usage Bytes",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("netbytes", "recv"),
            "Virtual Memory Usage Percent",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("ioreadcount", "total"),
            "I/O Read Count",
            []string{}, nil),

        prometheus.NewDesc(
            collectorName("ioreadbytes", "total"),
            "I/O Read Bytes",
            []string{}, nil),
        prometheus.NewDesc(
            collectorName("iowritecount", "total"),
            "I/O Write Count",
            []string{}, nil),
        prometheus.NewDesc(
            collectorName("iowritebytes", "total"),
            "I/O Write Bytes",
            []string{}, nil),
    }
    return col
}

func collectorName(subsystem, name string) string {
    return prometheus.BuildFQName(collectorsNamespace, subsystem, name)
}
