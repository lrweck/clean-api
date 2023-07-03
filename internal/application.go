package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shopspring/decimal"
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

	decimal.MarshalJSONWithoutQuotes = true

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
	a.Common.Logger.Info("starting application", slog.Int("port", port))

	err := a.WebServer.Start(strPort)

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (a *Application) Stop(ctx context.Context, waitCh <-chan os.Signal) error {
	signal := <-waitCh
	a.EndTime = time.Now()

	start := time.Now()
	err := a.WebServer.Shutdown(ctx)
	took := time.Since(start)

	a.Common.Logger.Info("signal received, stopping application",
		slog.Duration("up_for", a.EndTime.Sub(a.StartTime)),
		slog.String("signal", signal.String()),
		slog.Duration("shutdown_took", took))

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (a *Application) WaitSignal() error {
	waitCh := make(chan os.Signal, 1)
	signal.Notify(waitCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return a.Stop(ctx, waitCh)
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
	return &Storages{
		accStorage: memorydb.NewAccountStorage(), // postgres.NewAccountStorage(db)
		txStorage:  postgres.NewTxStorage(db),
	}
}

func getWebServer(svc *Services, cm *Common) *echo.Echo {
	app := echo.New()
	// goccy is muuuch faster
	// app.JSONSerializer = rest.NewGoccyEchoSerializer()

	app.HideBanner = true
	app.HidePort = true

	app.Use(app_middleware.NewLogger(cm.Logger))
	app.Use(middleware.RequestIDWithConfig(app_middleware.RequestIDConfig()))
	app.Use(middleware.Recover())

	configureRoutes(app, svc)

	return app
}

func configureRoutes(e *echo.Echo, svcs *Services) {
	V1 := e.Group("/v1")

	accounts := V1.Group("/accounts")
	accounts.POST("", rest.V1_POST_Account(svcs.accService))
	accounts.GET("/:id", rest.V1_GET_Account(svcs.accService))

	// transfers := V1.Group("/transfers")
	// transfers.POST("/", rest.V1POSTTransfer(svcs.txService))
	// transfers.GET("/:id", rest.V1GETTransfer(svcs.txService))

}
