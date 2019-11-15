package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zdnscloud/g53"
	"github.com/zdnscloud/vanguard/core"
)

var gMetrics = newMetrics()

const TotalView = "any"

type Metrics struct {
	reg      *prometheus.Registry
	viewQps  map[string]*Counter
	stopChan chan struct{}
}

func newMetrics() *Metrics {
	mt := &Metrics{
		reg:      prometheus.NewRegistry(),
		viewQps:  make(map[string]*Counter),
		stopChan: make(chan struct{}),
	}

	mt.reg.MustRegister(RequestCount)
	mt.reg.MustRegister(ResponseCount)
	mt.reg.MustRegister(UpdateCount)
	mt.reg.MustRegister(QPS)
	mt.reg.MustRegister(CacheSize)
	mt.reg.MustRegister(CacheHits)

	mt.reg.MustRegister(RequestCountByView)
	mt.reg.MustRegister(ResponseCountByView)
	mt.reg.MustRegister(UpdateCountByView)
	mt.reg.MustRegister(QPSByView)
	mt.reg.MustRegister(CacheSizeByView)
	mt.reg.MustRegister(CacheHitsByView)

	return mt
}

func Run() {
	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-gMetrics.stopChan:
			gMetrics.stopChan <- struct{}{}
			return
		case <-timer.C:
		}
		for view, counter := range gMetrics.viewQps {
			if view == TotalView {
				QPS.WithLabelValues("server").Set(float64(counter.Count()))
			} else {
				QPSByView.WithLabelValues("server", view).Set(float64(counter.Count()))
			}
			counter.Clear()
		}
	}
}

func Stop() {
	gMetrics.stopChan <- struct{}{}
	<-gMetrics.stopChan
	gMetrics.viewQps = make(map[string]*Counter)
}

func Handler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := promhttp.HandlerFor(gMetrics.reg, promhttp.HandlerOpts{})
		handler.ServeHTTP(w, r)
	}
}

func RecordMetrics(client core.Client) {
	if client.Request.Header.Opcode == g53.OP_QUERY {
		recordQps(TotalView)
		recordQps(client.View)
		RequestCount.WithLabelValues("server").Inc()
		RequestCountByView.WithLabelValues("server", client.View).Inc()
		if client.Response != nil {
			ResponseCount.WithLabelValues("server").Inc()
			ResponseCountByView.WithLabelValues("server", client.View).Inc()
		}
	} else if client.Request.Header.Opcode == g53.OP_UPDATE {
		if client.Response != nil {
			UpdateCount.WithLabelValues("server").Inc()
			if client.View != "" {
				UpdateCountByView.WithLabelValues("server", client.View).Inc()
			}
		}
	}
}

func recordQps(view string) {
	counter, ok := gMetrics.viewQps[view]
	if ok == false {
		counter = newCounter()
		gMetrics.viewQps[view] = counter
	}

	counter.Inc()
}

func RecordCacheHit(view string) {
	CacheHits.WithLabelValues("cache").Inc()
	CacheHitsByView.WithLabelValues("cache", view).Inc()
}

func RecordCacheSize(view string, size int, totalSize int) {
	CacheSize.WithLabelValues("cache").Set(float64(totalSize))
	CacheSizeByView.WithLabelValues("cache", view).Set(float64(size))
}
