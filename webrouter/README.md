# Package Webrouter

## 1) Webrouter with database

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
    // init a web router with a postgres database container
	var wr = webrouter.CreateWebRouterWithDb[*db.PgSqlDb]()

    // register a testroute
	wr.RegisterRoute("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("45z45")
	})

    //create a mux and let routes get handled by it
	mux := http.NewServeMux()
	wr.HandleByMux(mux)
	handler := cors.Default().Handler(mux)

    //start the server
    wsPort := "6060"
	fmt.Println("=> Starting webserver on Port " + wsPort)
	http.ListenAndServe(":" + wsPort, handler)
}
```