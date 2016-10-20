package router

import (
	"github.com/asiainfoLDP/datafoundry_coupon/api"
	"github.com/asiainfoLDP/datafoundry_coupon/handler"
	"github.com/asiainfoLDP/datafoundry_coupon/log"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

const (
	Platform_Local  = "local"
	Platform_DataOS = "dataos"
)

var (
	Platform = Platform_DataOS
	logger   = log.GetLogger()
)

//==============================================================
//
//==============================================================

func handler_Index(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	api.JsonResult(w, http.StatusNotFound, api.GetError(api.ErrorCodeUrlNotSupported), nil)
}

func httpNotFound(w http.ResponseWriter, r *http.Request) {
	api.JsonResult(w, http.StatusNotFound, api.GetError(api.ErrorCodeUrlNotSupported), nil)
}

type HttpHandler struct {
	handler http.HandlerFunc
}

func (httpHandler *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if httpHandler.handler != nil {
		httpHandler.handler(w, r)
	}
}

//==============================================================
//
//==============================================================

func InitRouter() *httprouter.Router {
	router := httprouter.New()
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false

	router.POST("/", handler_Index)
	router.DELETE("/", handler_Index)
	router.PUT("/", handler_Index)
	router.GET("/", handler_Index)

	router.NotFound = &HttpHandler{httpNotFound}
	router.MethodNotAllowed = &HttpHandler{httpNotFound}

	return router
}

func NewRouter(router *httprouter.Router) {
	logger.Info("new router.")
	router.POST("/charge/v1/coupons", api.TimeoutHandle(500*time.Millisecond, handler.CreateCoupon))
	//router.DELETE("/charge/v1/coupons/:id", api.TimeoutHandle(500*time.Millisecond, handler.DeletePlan))
	//router.PUT("/charge/v1/coupons/:id", api.TimeoutHandle(500*time.Millisecond, handler.ModifyPlan))
	router.PUT("/charge/v1/coupons/use/:serial/:code", api.TimeoutHandle(500*time.Millisecond, handler.UseCoupon))
	//router.GET("/charge/v1/coupons/:id", api.TimeoutHandle(500*time.Millisecond, handler.RetrievePlan))
	//router.GET("/charge/v1/coupons", api.TimeoutHandle(500*time.Millisecond, handler.QueryPlanList))
	//router.GET("/charge/v1/query/coupons/region", api.TimeoutHandle(500*time.Millisecond, handler.RetrievePlanRegion))
}