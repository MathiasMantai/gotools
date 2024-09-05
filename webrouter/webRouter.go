package webrouter

import (
	"github.com/MathiasMantai/gotools/db"
	"net/http"
)

type Dbs interface {
	*db.MssqlDb | *db.SqliteDb | *db.PgSqlDb
}

type WebRouterWithDb[T Dbs] struct {
	LastRoute      string
	DbContainer    map[string](T)
	RouteContainer map[string]interface{}
}

func (wr *WebRouterWithDb[T]) InitRouteContainer() {
	wr.RouteContainer = make(map[string]interface{})
}

func (wr *WebRouterWithDb[T]) InitDbContainer() {
	wr.DbContainer = make(map[string](T))
}

func (wr *WebRouterWithDb[T]) Init() {
	wr.InitRouteContainer()
	wr.InitDbContainer()
}

func (wr *WebRouterWithDb[T]) RegisterDb(label string, db T) {
	wr.DbContainer[label] = db
}

func (wr *WebRouterWithDb[T]) GetDb(label string) T {
	return wr.DbContainer[label]
}

func (wr *WebRouterWithDb[T]) RegisterRoute(label string, routeFunc interface{}) {
	wr.RouteContainer[label] = routeFunc
}

func (wr *WebRouterWithDb[T]) RegisterRouteMap(routeFuncs map[string]interface{}) {
	for label, routeFunc := range routeFuncs {
        wr.RegisterRoute(label, routeFunc)
    }
}

func (wr *WebRouterWithDb[T]) HandleByMux(mux *http.ServeMux) {
	for label, routeFunc := range wr.RouteContainer {
		//convert and handle by mux
		// fmt.Println("Route: " + label)
		mux.HandleFunc(label, routeFunc.(func(http.ResponseWriter, *http.Request)))
	}
}

func CreateWebRouterWithDb[T Dbs]() *WebRouterWithDb[T] {
	var wr = &WebRouterWithDb[T] {
		DbContainer: make(map[string](T)),
		RouteContainer: make(map[string]interface{}),
	}
	return wr
}

/* WebRouter no db */

type WebRouter struct {
	LastRoute string `json:"last_route"`
	RouteContainer map[string]interface{} `json:"route_container"`
}

func CreateWebRouter() *WebRouter {
	return &WebRouter{
		RouteContainer: make(map[string]interface{}),
		LastRoute: "/",
	}
}