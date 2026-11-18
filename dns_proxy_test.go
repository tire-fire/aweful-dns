package main

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

// upstream starts a DNS server that answers every A query with addr and every AAAA query with an
// IPv6 address, and returns its address.
func upstream(t *testing.T, addr string) string {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &dns.Server{PacketConn: conn, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		switch r.Question[0].Qtype {
		case dns.TypeA:
			m.Answer = append(m.Answer, &dns.A{Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60}, A: net.ParseIP(addr)})
		case dns.TypeAAAA:
			m.Answer = append(m.Answer, &dns.AAAA{Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60}, AAAA: net.ParseIP("2001:db8::1")})
		}
		_ = w.WriteMsg(m)
	})}
	go func() { _ = server.ActivateAndServe() }()
	t.Cleanup(func() { _ = server.Shutdown() })
	return conn.LocalAddr().String()
}

func query(t *testing.T, upstreamAddr string, qtype uint16) *dns.Msg {
	t.Helper()
	req := new(dns.Msg)
	req.SetQuestion("skerja.team03.ice.land.", qtype)
	domains := []string{"skerja.team03.ice.land", "ice.land"}
	resp, err := getResponse(req, "192.168.220.", "10.100.103.", upstreamAddr, domains[:0:0])
	if err != nil {
		// a refused answer is reported as an error and answered with an empty reply
		return new(dns.Msg)
	}
	return resp
}

func TestInternalAnswerIsMappedToTheTeamRange(t *testing.T) {
	resp := query(t, upstream(t, "192.168.220.97"), dns.TypeA)
	if len(resp.Answer) != 1 || resp.Answer[0].(*dns.A).A.String() != "10.100.103.97" {
		t.Fatalf("want 10.100.103.97, got %v", resp.Answer)
	}
}

func TestAnswerOutsideTheInternalRangeIsRefused(t *testing.T) {
	for _, addr := range []string{"10.100.142.97", "172.18.0.5", "127.0.0.1", "192.168.221.97"} {
		if resp := query(t, upstream(t, addr), dns.TypeA); len(resp.Answer) != 0 {
			t.Errorf("answer %s was passed through as %v", addr, resp.Answer)
		}
	}
}

func TestOnlyAQueriesAreAnswered(t *testing.T) {
	if resp := query(t, upstream(t, "192.168.220.97"), dns.TypeAAAA); len(resp.Answer) != 0 {
		t.Fatalf("AAAA query was answered with %v", resp.Answer)
	}
}
