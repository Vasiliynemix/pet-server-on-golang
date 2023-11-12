package server

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/server/handlers/auth/login"
	"PetProjectGo/internal/server/handlers/auth/refresh"
	"PetProjectGo/internal/server/handlers/auth/register"
	"PetProjectGo/internal/server/handlers/auth/unlogin"
	"PetProjectGo/internal/server/handlers/market/category"
	"PetProjectGo/internal/server/handlers/market/product"
	userGroup "PetProjectGo/internal/server/handlers/user"
	mwLogger "PetProjectGo/internal/server/middleware/logger"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/storage/mongodb"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
)

type Server struct {
	log      *logging.Logger
	cfg      *config.Config
	router   *chi.Mux
	mongo    *mongodb.MongoDB
	postgres *sqlx.DB
	auth     *GroupServerAuth
	user     *GroupServerUser
	market   *GroupServerMarket
}

type GroupServerAuth struct {
	register *register.HandlerRegister
	login    *login.HandlerLogin
	unlogin  *unlogin.HandlerUnLogin
	refresh  *refresh.HandlerRefresh
}

type GroupServerUser struct {
	userInfo *userGroup.HandlerUserGet
}

type GroupServerMarket struct {
	category *category.HandlerCategoryAdd
	product  *product.HandlerProductAdd
}

func NewWebServer(
	log *logging.Logger,
	cfg *config.Config,
	mongo *mongodb.MongoDB,
	postgres *sqlx.DB,
) (*Server, error) {
	userService := services.NewUserService(log, &cfg.App, mongo, postgres)
	marketCService, err := services.NewMarketCategoryService(mongo, userService)
	if err != nil {
		return nil, err
	}

	marketPService, err := services.NewMarketProductService(mongo, marketCService)
	if err != nil {
		return nil, err
	}
	return &Server{
		log:    log,
		cfg:    cfg,
		router: chi.NewRouter(),
		auth:   NewGroupAuth(cfg, log, userService),
		user:   NewGroupUser(log, userService),
		market: NewGroupMarket(log, marketCService, marketPService),
	}, nil
}

func NewGroupAuth(
	cfg *config.Config,
	log *logging.Logger,
	userService *services.UserService,
) *GroupServerAuth {
	return &GroupServerAuth{
		register: register.NewHandlerRegister(&cfg.App, log, userService),
		login:    login.NewHandlerLogin(log, userService),
		unlogin:  unlogin.NewHandlerUnLogin(log, userService),
		refresh:  refresh.NewHandlerRefresh(log, userService),
	}
}

func NewGroupUser(
	log *logging.Logger,
	userService *services.UserService,
) *GroupServerUser {
	return &GroupServerUser{
		userInfo: userGroup.NewHandlerUserGet(log, userService),
	}
}

func NewGroupMarket(
	log *logging.Logger,
	categoryService *services.MarketCategoryService,
	productService *services.MarketProductService,
) *GroupServerMarket {
	return &GroupServerMarket{
		category: category.NewHandlerCategoryAdd(log, categoryService),
		product:  product.NewHandlerProductAdd(log, productService),
	}
}

func (s *Server) Run() {
	s.log.Info("Server started", zap.String("address", s.cfg.Web.Address))

	srv := &http.Server{
		Addr:         s.cfg.Web.Address,
		Handler:      s.router,
		ReadTimeout:  s.cfg.Web.Timeout,
		WriteTimeout: s.cfg.Web.Timeout,
		IdleTimeout:  s.cfg.Web.IdleTimeout,
	}

	s.registerMiddlewares()
	s.registerRouters()

	err := srv.ListenAndServe()
	if err != nil {
		s.log.Fatal("Server error", zap.Error(err))
	}
}

func (s *Server) registerMiddlewares() {
	s.log.Info("Registering middlewares")
	s.router.Use(middleware.RequestID)
	s.router.Use(mwLogger.NewLoggerMw(s.log))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)
}

func (s *Server) registerRouters() {
	s.log.Info("Registering routers")

	s.log.Info("Registering auth group")
	s.router.Route("/auth", func(r chi.Router) {
		r.Post("/register", s.auth.register.RegisterHandler())
		r.Post("/login", s.auth.login.LoginHandler())
		r.Post("/unlogin", s.auth.unlogin.UnLoginHandler())
		r.Post("/refresh", s.auth.refresh.RefreshHandler())
	})

	s.log.Info("Registering user group")
	s.router.Route("/user", func(r chi.Router) {
		r.Get("/me", s.user.userInfo.UserGetHandler())
	})

	s.log.Info("Registering category group")
	s.router.Route("/category", func(r chi.Router) {
		r.Post("/add", s.market.category.AddCategoryHandler())
	})

	s.log.Info("Registering product group")
	s.router.Route("/product", func(r chi.Router) {
		r.Post("/add", s.market.product.AddProductHandler())
	})
}
