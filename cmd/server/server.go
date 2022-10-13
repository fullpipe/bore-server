package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/fullpipe/bore-server/config"
	"github.com/fullpipe/bore-server/graph"
	"github.com/fullpipe/bore-server/graph/generated"
	"github.com/glebarez/sqlite"
	"github.com/go-chi/chi"
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
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	router := chi.NewRouter()

	// Add CORS middleware around every request
	// See https://github.com/rs/cors for full option listing
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "http://localhost:8100", "http://localhost:4200"},
		AllowCredentials: true,
		Debug:            cfg.Debug,
	}).Handler)

	db, err := gorm.Open(sqlite.Open(cfg.LiteDB), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graph.NewResolver(db, cfg)}))

	// srv.AddTransport(&transport.Websocket{
	// 	Upgrader: websocket.Upgrader{
	// 		CheckOrigin: func(r *http.Request) bool {
	// 			// Check against your desired domains here
	// 			return r.Host == "example.org"
	// 		},
	// 		ReadBufferSize:  1024,
	// 		WriteBufferSize: 1024,
	// 	},
	// })

	fs := http.FileServer(http.Dir(cfg.BooksDir))
	router.Handle("/books/*", http.StripPrefix("/books", fs))

	if cfg.Debug {
		router.Handle("/playground", playground.Handler("Bore", "/query"))
	}

	router.Handle("/query", srv)

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), router)
}
