package internal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/exp/slog"

	"github.com/lrweck/clean-api/internal/account"
	"github.com/lrweck/clean-api/internal/transfer"
	"github.com/lrweck/clean-api/pkg/memorydb"
	"github.com/lrweck/clean-api/pkg/postgres"
	"github.com/lrweck/clean-api/pkg/rest"
	app_middleware "github.com/lrweck/clean-api/pkg/rest/middleware"
	"github.com/lrweck/clean-api/pkg/slogger"
)

type Application struct {
	WebServer *echo.Echo
	Services  *Services
	Storages  *Storages
	Common    *Common
	StartTime time.Time
	EndTime   time.Time
}

func NewApplication() *Application {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()

	// writeDB, err := postgres.NewDB(ctx, os.Getenv("PG_DSN"))
	// if err != nil {
	// 	panic(fmt.Errorf("failed to connect to write database: %w", err))
	// }

	common := getCommons()
	storages := getStorages(nil)
	services := getServices(storages)
	webServer := getWebServer(services, common)

	return &Application{
		WebServer: webServer,
		Services:  services,
		Storages:  storages,
		Common:    common,
	}
}

func (a *Application) Start(port int) error {
	strPort := fmt.Sprintf(":%d", port)
	a.StartTime = time.Now()
	return a.WebServer.Start(strPort)
}

func (a *Application) Stop(ctx context.Context) error {
	a.EndTime = time.Now()
	<-ctx.Done()
	a.Common.Logger.Info("stopping application, signal received", slog.Duration("took", a.EndTime.Sub(a.StartTime)))
	return a.WebServer.Shutdown(ctx)
}

func (a *Application) WaitSignal() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return a.Stop(ctx)
}

type Common struct {
	Logger *slog.Logger
	// otel, etc
}

func getCommons() *Common {

	env := os.Getenv("ENV")

	logger := slogger.NewJSON()
	if env == "" {
		logger = slogger.NewText()
	}

	return &Common{
		Logger: logger,
	}
}

type Storages struct {
	accStorage account.Storage
	txStorage  transfer.Storage
}

type Services struct {
	accService *account.Service
	txService  *transfer.Service
}

func getServices(storages *Storages) *Services {
	return &Services{
		accService: account.NewService(storages.accStorage, uuid.New, time.Now),
		txService:  transfer.NewService(storages.txStorage, uuid.New, time.Now),
	}
}

func getStorages(db *pgxpool.Pool) *Storages {

	// accStorage := postgres.NewAccountStorage(db)
	accStorage := memorydb.NewAccountStorage()

	return &Storages{
		accStorage: accStorage,
		txStorage:  postgres.NewTxStorage(db),
	}
}

func getWebServer(svc *Services, cm *Common) *echo.Echo {
	app := echo.New()
	// goccy is muuuch faster
	app.JSONSerializer = rest.NewGoccyEchoSerializer()

	app.Use(app_middleware.NewLogger(cm.Logger))
	app.Use(middleware.RequestIDWithConfig(app_middleware.RequestID()))
	app.Use(middleware.Recover())

	configureRoutes(app, svc)

	return app
}

func configureRoutes(e *echo.Echo, svcs *Services) {
	V1 := e.Group("/v1")

	V1.POST("/accounts", rest.V1POSTAccount(svcs.accService))
	V1.GET("/accounts/:id", rest.V1GETAccount(svcs.accService))
}
