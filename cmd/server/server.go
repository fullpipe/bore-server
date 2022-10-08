package server

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/fullpipe/bore-server/graph"
	"github.com/fullpipe/bore-server/graph/generated"
	"github.com/fullpipe/bore-server/repository"
	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/urfave/cli"
	"gorm.io/gorm"
)

const defaultPort = "8080"

func NewCommand() cli.Command {
	return cli.Command{
		Name:   "server",
		Action: server,
	}

}

func server(cCtx *cli.Context) error {
	router := chi.NewRouter()

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "http://localhost:8100"},
		AllowCredentials: true,
		Debug:            true,
	}).Handler)

	db, err := gorm.Open(sqlite.Open("lite.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	bookRepo := repository.NewBookRepo(db)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graph.NewResolver(db, bookRepo)}))

	srv.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Check against your desired domains here
				return r.Host == "example.org"
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})

	router.Handle("/", playground.Handler("Bore", "/query"))
	router.Handle("/query", srv)

	return http.ListenAndServe(":8080", router)
}
