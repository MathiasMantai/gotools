package db

import (
	"net/http"
	"github.com/MathiasMantai/gotools/db"
)

type Dbs interface {
	*db.MssqlDb | *db.SqliteDb | *db.PgSqlDb
}

type WebRouter[T Dbs] struct {
	LastRoute string
	DbContainer map[string](T)
	RouteContainer map[string]interface{}
}

func (wr *WebRouter[T]) InitRouteContainer() {
	wr.RouteContainer = make(map[string]interface{})
}

func (wr *WebRouter[T]) InitDbContainer() {
    wr.DbContainer = make(map[string](T))
}

func (wr *WebRouter[T]) Init() {
	wr.InitRouteContainer()
    wr.InitDbContainer()
}

func (wr *WebRouter[T]) RegisterDb(label string, db T) {
	wr.DbContainer[label] = db
}

func (wr *WebRouter[T]) GetDb(label string) T {
    return wr.DbContainer[label]
}

func (wr *WebRouter[T]) RegisterRoute(label string, routeFunc interface{}) {
	wr.RouteContainer[label] = routeFunc
}

func (wr *WebRouter[T]) HandleByMux(mux *http.ServeMux) {
	for label, routeFunc := range wr.RouteContainer {
		//convert and handle by mux
		mux.HandleFunc(label, routeFunc.(func(http.ResponseWriter, *http.Request)))
	}
}

func CreateWebRouter[T Dbs]() *WebRouter[T] {
	return new(WebRouter[T])
}