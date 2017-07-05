package fastmux

import (
	"github.com/nehmeroumani/fastchain"
	"github.com/nehmeroumani/fasthttpcontext"
	"github.com/nehmeroumani/fasthttptreemux"
	"github.com/valyala/fasthttp"
)

func New() *Mux {
	return &Mux{Router: fasthttptreemux.New()}
}

func wrapHandler(h fasthttp.RequestHandler) fasthttptreemux.HandlerFunc {
	return func(requestCtx *fasthttp.RequestCtx, params map[string]string) {
		fasthttpcontext.Set(requestCtx, "params", params)
		fasthttpcontext.ClearHandler(h)(requestCtx)
	}
}

type Mux struct {
	Router *fasthttptreemux.TreeMux
	Chain  fastchain.Chain
}

func (this *Mux) Use(middlewares ...fastchain.Constructor) {
	this.Chain = this.Chain.Append(middlewares...)
}
func (this *Mux) Get(p string) *route {
	return &route{mux: this, pattern: p, method: "GET", chain: this.Chain}
}
func (this *Mux) Post(p string) *route {
	return &route{mux: this, pattern: p, method: "POST", chain: this.Chain}
}
func (this *Mux) Put(p string) *route {
	return &route{mux: this, pattern: p, method: "PUT", chain: this.Chain}
}
func (this *Mux) Patch(p string) *route {
	return &route{mux: this, pattern: p, method: "PATCH", chain: this.Chain}
}
func (this *Mux) Delete(p string) *route {
	return &route{mux: this, pattern: p, method: "DELETE", chain: this.Chain}
}
func (this *Mux) Head(p string) *route {
	return &route{mux: this, pattern: p, method: "HEAD", chain: this.Chain}
}
func (this *Mux) Options(p string) *route {
	return &route{mux: this, pattern: p, method: "OPTIONS", chain: this.Chain}
}
func (this *Mux) ServeHTTP(requestCtx *fasthttp.RequestCtx) {
	this.Router.ServeHTTP(requestCtx)
}
func (this *Mux) NotFoundHandler(h func(requestCtx *fasthttp.RequestCtx)) {
	this.Router.NotFoundHandler = h
}

type route struct {
	mux     *Mux
	chain   fastchain.Chain
	pattern string
	method  string
}

func (this *route) Use(middlewares ...fastchain.Constructor) *route {
	this.chain = this.chain.Append(middlewares...)
	return this
}

func (this *route) ThenFunc(h fasthttp.RequestHandler) {
	this.mux.Router.Handle(this.method, this.pattern, wrapHandler(this.chain.ThenFunc(h)))
}

// Params(r *http.Request) is a function to get URL params from the request fasthttpcontext
func Params(requestCtx *fasthttp.RequestCtx) map[string]string {
	if params, ok := fasthttpcontext.GetOk(requestCtx, "params"); ok {
		return params.(map[string]string)
	}
	return nil
}
func GetParam(requestCtx *fasthttp.RequestCtx, key string) string {
	if params := Params(requestCtx); params != nil {
		if value, ok := params[key]; ok {
			return value
		}
	}
	return ""
}
