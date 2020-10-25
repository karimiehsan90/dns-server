package main

import (
	"os"

	"github.com/karimiehsan90/dns-server/dns"
	"github.com/karimiehsan90/dns-server/metrics"
)

func main() {
	metricsServer := metrics.GetInstance(os.Getenv("METRICS_ADDR"))
	metricsServer.Run()
	dnsServer := dns.GetInstance(os.Getenv("REDIS_ADDR"), os.Getenv("ROOT_DNS_ADDR"), metricsServer)
	dnsServer.Run()
}
