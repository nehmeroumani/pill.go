package fastmux

import (
	"strings"

	"github.com/nehmeroumani/fastchain"
	"github.com/nehmeroumani/fasthttpcontext"
	"github.com/nehmeroumani/fasthttptreemux"
	"github.com/nehmeroumani/pill.go/helpers"
	"github.com/valyala/fasthttp"
)

func New(opts ...string) *Mux {
	basePath := ""
	if opts != nil && len(opts) > 0 {
		if opts[0] != "/" {
			basePath = opts[0]
		}
	}
	return &Mux{Router: fasthttptreemux.New(), basePath: basePath}
}

func wrapHandler(h fasthttp.RequestHandler) fasthttptreemux.HandlerFunc {
	return func(requestCtx *fasthttp.RequestCtx, params map[string]string) {
		fasthttpcontext.Set(requestCtx, "params", params)
		fasthttpcontext.ClearHandler(h)(requestCtx)
	}
}

type Mux struct {
	Router   *fasthttptreemux.TreeMux
	Chain    fastchain.Chain
	basePath string
}

func (this *Mux) Use(middlewares ...fastchain.Constructor) {
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

type GlobalRouter struct {
	webRouter *Mux
	apiRouter *Mux
	apiDomainName string
	apiPath   string
}

func NewGlobalRouter(WebRouter *Mux, APIRouter *Mux, APIDomainName string, APIPath string) *GlobalRouter {
	return &GlobalRouter{webRouter: WebRouter, apiRouter: APIRouter, apiDomainName: strings.ToLower(strings.TrimSpace(APIDomainName)), apiPath: strings.ToLower(strings.TrimSpace(APIPath))}
}

func (this GlobalRouter) ServeHTTP(requestCtx *fasthttp.RequestCtx) {
	if this.apiDomainName == "" {
		if strings.ToLower(helpers.BytesToString(requestCtx.Host())) == this.apiDomainName {
			this.apiRouter.ServeHTTP(requestCtx)
		} else {
			this.webRouter.ServeHTTP(requestCtx)
		}
	} else {
		if strings.HasPrefix(strings.ToLower(helpers.BytesToString(requestCtx.Path())), this.apiPath) {
			this.apiRouter.ServeHTTP(requestCtx)
		} else {
			this.webRouter.ServeHTTP(requestCtx)
		}
	}
}
