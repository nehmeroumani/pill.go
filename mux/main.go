package mux

import (
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/context"
	"github.com/justinas/alice"
)

func New(opts ...string) *Mux {
	basePath := ""
	if opts != nil && len(opts) > 0 {
		if opts[0] != "/" {
			basePath = opts[0]
		}
	}
	return &Mux{Router: httptreemux.New(), basePath: basePath}
}

func wrapHandler(h http.Handler) httptreemux.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		context.Set(req, "params", params)
		defer func() {
			req.Body.Close()
			req.Header.Set("Connection", "close")
		}()
		context.ClearHandler(h).ServeHTTP(w, req)
	}
}

type Mux struct {
	Router   *httptreemux.TreeMux
	Chain    alice.Chain
	basePath string
}

func (this *Mux) Use(middlewares ...alice.Constructor) {
	this.Chain = this.Chain.Append(middlewares...)
}
func (this *Mux) Get(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "GET", chain: this.Chain}
}
func (this *Mux) Post(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "POST", chain: this.Chain}
}
func (this *Mux) Put(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "PUT", chain: this.Chain}
}
func (this *Mux) Patch(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "PATCH", chain: this.Chain}
}
func (this *Mux) Delete(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "DELETE", chain: this.Chain}
}
func (this *Mux) Head(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "HEAD", chain: this.Chain}
}
func (this *Mux) Options(p string) *route {
	return &route{mux: this, pattern: this.basePath + p, method: "OPTIONS", chain: this.Chain}
}
func (this *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	this.Router.ServeHTTP(w, req)
}
func (this *Mux) NotFoundHandler(h func(http.ResponseWriter, *http.Request)) {
	this.Router.NotFoundHandler = h
}

type route struct {
	mux     *Mux
	chain   alice.Chain
	pattern string
	method  string
}

func (this *route) Use(middlewares ...alice.Constructor) *route {
	this.chain = this.chain.Append(middlewares...)
	return this
}

func (this *route) Then(h http.Handler) {
	this.mux.Router.Handle(this.method, this.pattern, wrapHandler(this.chain.Then(h)))
}

func (this *route) ThenFunc(h http.HandlerFunc) {
	this.mux.Router.Handle(this.method, this.pattern, wrapHandler(this.chain.ThenFunc(h)))
}

// Params(r *http.Request) is a function to get URL params from the request context
func Params(req *http.Request) map[string]string {
	if params, ok := context.GetOk(req, "params"); ok {
		return params.(map[string]string)
	}
	return nil
}
func GetParam(r *http.Request, key string) string {
	if params := Params(r); params != nil {
		if value, ok := params[key]; ok {
			return value
		}
	}
	return ""
}

type GlobalRouter struct {
	WebRouter   *Mux
	APIRouter   *Mux
	Environment string
	DomainName  string
}

func (this GlobalRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if this.Environment == "production" {
		if strings.ToLower(req.Host) == strings.ToLower("api."+this.DomainName) {
			this.APIRouter.ServeHTTP(w, req)
		} else {
			this.WebRouter.ServeHTTP(w, req)
		}
	} else {
		if strings.HasPrefix(strings.ToLower(req.URL.Path), "/api") {
			this.APIRouter.ServeHTTP(w, req)
		} else {
			this.WebRouter.ServeHTTP(w, req)
		}
	}
}
