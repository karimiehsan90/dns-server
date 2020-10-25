package dns

import (
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"net"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/karimiehsan90/dns-server/metrics"
)

type Server struct {
	redisAddr     string
	rootDnsAddr   string
	redisClient   *redis.Client
	metricsServer *metrics.Server
}

var instance *Server

func GetInstance(redisAddr string, rootDnsAddr string, metricsServer *metrics.Server) *Server {
	if instance == nil {
		instance = &Server{}
		instance.redisClient = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
		instance.metricsServer = metricsServer
		instance.redisAddr = redisAddr
		instance.rootDnsAddr = rootDnsAddr
	}
	return instance
}

func (s *Server) askFromDnsRootServer(hostname string) string {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", s.rootDnsAddr+":53")
		},
	}
	ip, _ := r.LookupHost(context.Background(), hostname)
	return ip[0]
}

func (s *Server) serveDNS(u *net.UDPConn, clientAddr net.Addr, request *layers.DNS) {
	log.Info(string(request.Questions[0].Name))
	s.metricsServer.RequestCount.Inc()
	replyMess := request
	dnsAnswer := layers.DNSResourceRecord{}
	dnsAnswer.Type = layers.DNSTypeA
	redisCommand := s.redisClient.Get(string(request.Questions[0].Name))
	ip, _ := redisCommand.Result()
	if ip == "" {
		s.metricsServer.MissCount.Inc()
		ip = s.askFromDnsRootServer(string(request.Questions[0].Name))
	} else {
		s.metricsServer.HitsCount.Inc()
	}
	a, _, _ := net.ParseCIDR(ip + "/24")
	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = a
	dnsAnswer.Name = request.Questions[0].Name
	fmt.Println(string(request.Questions[0].Name))
	dnsAnswer.Class = layers.DNSClassIN
	replyMess.QR = true
	replyMess.ANCount = 1
	replyMess.OpCode = layers.DNSOpCodeNotify
	replyMess.AA = true
	replyMess.Answers = append(replyMess.Answers, dnsAnswer)
	replyMess.ResponseCode = layers.DNSResponseCodeNoErr
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err := replyMess.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	u.WriteTo(buf.Bytes(), clientAddr)
}

func (s *Server) Run() {
	log.Info("Running")
	addr := net.UDPAddr{Port: 53, IP: net.ParseIP("0.0.0.0")}
	udpConn, _ := net.ListenUDP("udp", &addr)
	for {
		tmp := make([]byte, 1024)
		_, clientAddr, _ := udpConn.ReadFrom(tmp)
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		dns := dnsPacket.(*layers.DNS)
		s.serveDNS(udpConn, clientAddr, dns)
	}
}
