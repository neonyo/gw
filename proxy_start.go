package gw

import (
	"errors"
	"fmt"
	"github.com/soheilhy/cmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"net"
	"net/http"
)

func (p *proxy) start(addr string) error {
	if p.httpProxy.conf.hostSwitch == nil {
		return errors.New("服务选择没有设置")
	}
	p.startHTTP(addr)
	return nil
}

func (p *proxy) Stop() {

}

func (p *proxy) startHTTPWithListener(l net.Listener) {
	s := p.newHTTPServer()
	err := s.Serve(l)
	if err != nil {
		log.Fatalf("start http listeners failed with %+v", err)
	}
}

func (p *proxy) startHTTP(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("start http failed failed with %+v", err)
	}
	m := cmux.New(l)
	go p.startHTTPWithListener(m.Match(cmux.Any()))
	fmt.Printf("Starting http server at %s...\n", addr)
	err = m.Serve()
	if err != nil {
		log.Fatalf("start http failed failed with %+v", err)
	}
}

func (p *proxy) newHTTPServer() *http.Server {
	if !p.httpProxy.conf.Telemetry {
		return &http.Server{
			Handler: p.httpProxy,
		}
	}
	return &http.Server{
		Handler: otelhttp.NewHandler(p.httpProxy, "otel-go-tracer", p.mwOptions()...),
	}
}

func (p *proxy) mwOptions() []otelhttp.Option {
	var options []otelhttp.Option
	options = append(options, otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
		return fmt.Sprintf("%s", r.URL.Path)
	}))
	return options
}
