package metrics

import (
	"github.com/prometheus/common/log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	bindAddr     string
	RequestCount prometheus.Counter
	HitsCount    prometheus.Counter
	MissCount    prometheus.Counter
}

var instance *Server

func GetInstance(bindAddress string) *Server {
	if instance == nil {
		instance = &Server{}
		instance.bindAddr = bindAddress
		instance.RequestCount = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "requests_count",
				Help: "Request count",
			},
		)
		instance.HitsCount = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "hits_count",
				Help: "Request hits count",
			},
		)
		instance.MissCount = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "miss_count",
				Help: "Request miss count",
			},
		)
	}
	return instance
}

func (s *Server) Run() {
	log.Info("Running")
	prometheus.MustRegister(s.RequestCount)
	prometheus.MustRegister(s.HitsCount)
	prometheus.MustRegister(s.MissCount)
	http.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	go func(msg string) {
		serve := http.ListenAndServe(s.bindAddr, nil)
		log.Info(serve)
	}("go-routine")
}
