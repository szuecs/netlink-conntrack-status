package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ti-mo/conntrack"
)

var (
	version = "unset"
	commit  = "unset"

	versionFlag    = false
	daemonFlag     = false
	updateInterval time.Duration
)

type conntrackServer struct {
	prometheus *prometheusMetrics
	reg        *prometheus.Registry
}

type prometheusMetrics struct {
	foundM        *prometheus.CounterVec
	invalidM      *prometheus.CounterVec
	ignoreM       *prometheus.CounterVec
	insertM       *prometheus.CounterVec
	insertFailedM *prometheus.CounterVec
	dropM         *prometheus.CounterVec
	earlyDropM    *prometheus.CounterVec
	errorM        *prometheus.CounterVec
	serchRestartM *prometheus.CounterVec
}

type Stats struct {
	Found         uint32 `json:"found"`
	Invalid       uint32 `json:"invalid"`
	Ignore        uint32 `json:"ignore"`
	Insert        uint32 `json:"insert"`
	InsertFailed  uint32 `json:"insert_failed"`
	Drop          uint32 `json:"drop"`
	EarlyDrop     uint32 `json:"early_drop"`
	Error         uint32 `json:"error"`
	SearchRestart uint32 `json:"search_restart"`
}

func (cs *conntrackServer) startMetricsServer(quit <-chan struct{}) {
	mux := http.NewServeMux()

	metricsHandler := promhttp.HandlerFor(cs.reg, promhttp.HandlerOpts{})
	mux.Handle("/metrics", metricsHandler)
	mux.Handle("/metrics/", metricsHandler)

	go func() {
		if err := http.ListenAndServe(":9090", mux); err != nil {
			log.Fatalf("Failed to start metrics listener on %s: %v", metricsHandler, err)
		}
	}()
	<-quit
}

func (cs *conntrackServer) updateStats() *Stats {
	c, err := conntrack.Dial(nil)
	if err != nil {
		log.Fatalf("failed to dial netlink: %v", err)
	}
	defer c.Close()
	connStats, err := c.Stats()
	if err != nil {
		log.Fatalf("failed to get stats: %v", err)
	}

	var stats Stats
	for _, st := range connStats {
		stats.Found += st.Found
		stats.Invalid += st.Invalid
		stats.Ignore += st.Ignore
		stats.Insert += st.Insert
		stats.InsertFailed += st.InsertFailed
		stats.Drop += st.Drop
		stats.EarlyDrop += st.EarlyDrop
		stats.Error += st.Error
		stats.SearchRestart += st.SearchRestart
	}
	return &stats
}
func (cs *conntrackServer) registerMetrics() {
	cs.reg.MustRegister(cs.prometheus.foundM)
	cs.reg.MustRegister(cs.prometheus.invalidM)
	cs.reg.MustRegister(cs.prometheus.ignoreM)
	cs.reg.MustRegister(cs.prometheus.insertM)
	cs.reg.MustRegister(cs.prometheus.insertFailedM)
	cs.reg.MustRegister(cs.prometheus.dropM)
	cs.reg.MustRegister(cs.prometheus.earlyDropM)
	cs.reg.MustRegister(cs.prometheus.errorM)
	cs.reg.MustRegister(cs.prometheus.serchRestartM)
}

func (cs *conntrackServer) updateMetrics(oldStats, stats Stats) {
	cs.prometheus.foundM.WithLabelValues().Add(float64(stats.Found - oldStats.Found))
	cs.prometheus.invalidM.WithLabelValues().Add(float64(stats.Invalid - oldStats.Invalid))
	cs.prometheus.ignoreM.WithLabelValues().Add(float64(stats.Ignore - oldStats.Ignore))
	cs.prometheus.insertM.WithLabelValues().Add(float64(stats.Insert - oldStats.Insert))
	cs.prometheus.insertFailedM.WithLabelValues().Add(float64(stats.InsertFailed - oldStats.InsertFailed))
	cs.prometheus.dropM.WithLabelValues().Add(float64(stats.Drop - oldStats.Drop))
	cs.prometheus.earlyDropM.WithLabelValues().Add(float64(stats.EarlyDrop - oldStats.EarlyDrop))
	cs.prometheus.errorM.WithLabelValues().Add(float64(stats.Error - oldStats.Error))
	cs.prometheus.serchRestartM.WithLabelValues().Add(float64(stats.SearchRestart - oldStats.SearchRestart))

}

func newPrometheusMetrics() *prometheusMetrics {
	namespace := "netlink"
	subSystem := "conntrack"

	found := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "found",
		Help:      "The total of conntrack -S found.",
	}, []string{})
	invalid := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "invalid",
		Help:      "The total of conntrack -S invalid.",
	}, []string{})
	ignore := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "ignore",
		Help:      "The total of conntrack -S ignore.",
	}, []string{})
	insert := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "insert",
		Help:      "The total of conntrack -S insert.",
	}, []string{})
	insertFailed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "insertFailed",
		Help:      "The total of conntrack -S insertFailed.",
	}, []string{})
	drop := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "drop",
		Help:      "The total of conntrack -S drop.",
	}, []string{})
	earlyDrop := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "earlyDrop",
		Help:      "The total of conntrack -S earlyDrop.",
	}, []string{})
	error := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "error",
		Help:      "The total of conntrack -S error.",
	}, []string{})
	searchRestart := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subSystem,
		Name:      "searchRestart",
		Help:      "The total of conntrack -S searchRestart.",
	}, []string{})

	pm := &prometheusMetrics{
		foundM:        found,
		invalidM:      invalid,
		ignoreM:       ignore,
		insertM:       insert,
		insertFailedM: insertFailed,
		dropM:         drop,
		earlyDropM:    earlyDrop,
		errorM:        error,
		serchRestartM: searchRestart,
	}

	return pm
}

func newConntrackServer(pm *prometheusMetrics) *conntrackServer {
	return &conntrackServer{
		prometheus: pm,
		reg:        prometheus.NewRegistry(),
	}
}

func main() {
	flag.BoolVar(&versionFlag, "version", false, "print version")
	flag.BoolVar(&daemonFlag, "daemon", false, "daemonize process")
	flag.DurationVar(&updateInterval, "update-interval", time.Second, "Set time.Duration update interval 30s")
	flag.Parse()
	if versionFlag {
		fmt.Printf("%s: %s - commit: %s\n", os.Args[0], version, commit)
		os.Exit(0)
	}

	sigs := make(chan os.Signal, 1)
	quit := make(chan struct{}, 2)
	go func() {
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		<-sigs

		log.Println("shutting down")
		if daemonFlag {
			quit <- struct{}{}
			quit <- struct{}{}
		}
	}()

	cs := newConntrackServer(newPrometheusMetrics())
	stats := cs.updateStats()

	if daemonFlag {
		cs.registerMetrics()
		cs.updateMetrics(Stats{}, *stats)
		ticker := time.NewTicker(updateInterval)
		go func() {
			for {
				select {
				case <-quit:
					ticker.Stop()
					return
				case <-ticker.C:
					oldStats := *stats
					stats = cs.updateStats()
					cs.updateMetrics(oldStats, *stats)
				}
			}
		}()

		cs.startMetricsServer(quit)
	} else {
		buf, err := json.Marshal(stats)
		if err != nil {
			log.Fatalf("Failed to marshal json: %v", err)
		}
		fmt.Printf("%s\n", buf)
	}

}
