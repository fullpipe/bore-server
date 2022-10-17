package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/fullpipe/bore-server/auth"
	"github.com/fullpipe/bore-server/config"
	"github.com/fullpipe/bore-server/graph"
	"github.com/fullpipe/bore-server/graph/generated"
	"github.com/fullpipe/bore-server/graph/model"
	"github.com/fullpipe/bore-server/jwt"
	"github.com/fullpipe/bore-server/repository"
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
		AllowCredentials: false,
		AllowedHeaders:   []string{"*"},
		Debug:            cfg.Debug,
	}).Handler)

	db, err := gorm.Open(sqlite.Open(cfg.LiteDB), &gorm.Config{})
	if err != nil {
		return err
	}

	jwtParser, err := jwt.NewEdDSAParser(cfg.JWT.PublicKey, "access")
	if err != nil {
		return err
	}
	router.Use(auth.JwtMiddleware(db, jwtParser))

	resoleversConfig := generated.Config{Resolvers: graph.NewResolver(db, cfg)}
	resoleversConfig.Directives.HasRole = func(ctx context.Context, obj interface{}, next graphql.Resolver, role model.Role) (interface{}, error) {
		for _, r := range auth.Roles(ctx) {
			if r == string(role) {
				return next(ctx)
			}
		}

		return nil, fmt.Errorf("Access denied")
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(resoleversConfig))

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

	// refreshParser, _ := jwt.NewEdDSAParser(cfg.JWT.PublicKey, "refresh")
	// refreshTokenHandler := RefreshTokenHandler{
	// 	refreshParser: refreshParser,
	// 	userRepo:      repository.NewUserRepo(db),
	// }
	// router.Handle("/refresh", &refreshTokenHandler)

	if cfg.Debug {
		router.Handle("/playground", playground.Handler("Bore", "/query"))
	}

	router.Handle("/query", srv)

	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), router)
}

type RefreshTokenHandler struct {
	refreshParser jwt.Parser
	userRepo      *repository.UserRepo
}

func (h *RefreshTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Это моя домашняя страница"))
}
