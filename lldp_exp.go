package main


import (
	"os/exec"
	"encoding/xml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	log "github.com/Sirupsen/logrus"
	"strings"
)

type Data struct {
	Interfaces	[]Interfaces	`xml:"interface"`
}

type Interfaces	struct { 
	Name	string	`xml:"name,attr"`
	Chassis	Chassis	`xml:"chassis"`
	Port	Port	`xml:"port"`
}

type Chassis struct {
	Id	string	`xml:"id"`
	Name	string	`xml:"name"`
	Desc	string	`xml:"descr"`
	MgmtIp	string	`xml:"mgmt-ip"`
}

type Port struct {
	Id	string	`xml:"id"`
	Desc	string	`xml:"descr"`
	Ttl	string	`xml:"ttl"`
}

type lldpCollector struct {
	interfaceInfo	*prometheus.Desc
}

func newLldpCollector()	*lldpCollector	{
	return &lldpCollector	{
		interfaceInfo: prometheus.NewDesc("interfaceInfo",
			"Shows lldp neighbor information",
			[]string{"local_host","local_iface","chassis_mac","chassis_name","chassis_desc","chassis_mgmt_ip","remote_iface","port_desc"},
			nil,
		),
	}
}

func (collector *lldpCollector)	Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.interfaceInfo
}

func (collector *lldpCollector) Collect(ch chan<- prometheus.Metric) {
	cmd := exec.Command("lldpcli", "show", "neighbors", "-f", "xml")
	cmd2 := exec.Command("hostname")
	out, err := cmd.CombinedOutput()
	out2, err := cmd2.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	var lldp Data
	err = xml.Unmarshal([]byte(out), &lldp)
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	for _, inf := range lldp.Interfaces {
		ch <- prometheus.MustNewConstMetric(
			collector.interfaceInfo,
			prometheus.GaugeValue,
			float64(1),
			strings.TrimSuffix(string([]byte(out2)),"\n"),
			inf.Name,
			inf.Chassis.Id,
			inf.Chassis.Name,
			inf.Chassis.Desc,
			inf.Chassis.MgmtIp,
			inf.Port.Id,
			inf.Port.Desc)
	}
}

func main() {
  //Create a new instance of the lldpcollector and 
  //register it with the prometheus client.
  lldpcollector := newLldpCollector()
  prometheus.MustRegister(lldpcollector)

  //This section will start the HTTP server and expose
  //any metrics on the /metrics endpoint.
  http.Handle("/metrics", promhttp.Handler())
  log.Info("Beginning to serve on port :9700")
  log.Fatal(http.ListenAndServe(":9700", nil))
}
