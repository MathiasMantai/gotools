# Package Webrouter


## Webrouter with database

### Example

```go
import (
    "fmt"
    "net/http"
	"github.com/rs/cors"
	"github.com/MathiasMantai/gotools/db"
	"github.com/MathiasMantai/gotools/webrouter"
)

func main() {
	var wr = webrouter.CreateWebRouterWithDb[*db.PgSqlDb]()
	wr.RegisterRoute("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("45z45")
	})
	mux := http.NewServeMux()
	wsPort := "6060"
	wr.HandleByMux(mux)
	handler := cors.Default().Handler(mux)
	fmt.Println("=> Starting webserver on Port " + wsPort)
	http.ListenAndServe(":" + wsPort, handler)
}
```