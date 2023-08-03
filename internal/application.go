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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shopspring/decimal"
	"golang.org/x/exp/slog"

	"github.com/lrweck/clean-api/internal/account"
	"github.com/lrweck/clean-api/internal/transfer"
	"github.com/lrweck/clean-api/pkg/envutil"
	"github.com/lrweck/clean-api/pkg/memorydb"
	"github.com/lrweck/clean-api/pkg/postgres"
	"github.com/lrweck/clean-api/pkg/rest"
	app_middleware "github.com/lrweck/clean-api/pkg/rest/middleware"
	"github.com/lrweck/clean-api/pkg/slogger"
	"github.com/lrweck/clean-api/pkg/telemetry"
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
	telemetry.InitOTEL(context.Background(), common.OtelURL)

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

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (a *Application) Stop(ctx context.Context, waitCh <-chan os.Signal) error {
	signal := <-waitCh
	a.EndTime = time.Now()

	start := a.EndTime
	err := a.WebServer.Shutdown(ctx)
	took := time.Since(start)

	a.Common.Logger.Info("signal received, stopping application",
		slog.String("up_for", a.EndTime.Sub(a.StartTime).String()),
		slog.String("signal", signal.String()),
		slog.String("shutdown_took", took.String()))

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (a *Application) WaitSignal() error {
	waitCh := make(chan os.Signal, 1)

	signal.Notify(waitCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return a.Stop(ctx, waitCh)
}

type Common struct {
	Logger  *slog.Logger
	OtelURL string
}

func getCommons() *Common {

	env := envutil.CurrentEnv()
	var logger *slog.Logger

	switch env {
	case "devel":
		logger = slogger.NewTextWithOptions(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	case "production":
		logger = slogger.NewJSON()
	}

	return &Common{
		Logger:  logger,
		OtelURL: envutil.OTELExporterEndpointGo(),
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
		accService: account.NewService(storages.accStorage, nil, time.Now),
		txService:  transfer.NewService(storages.txStorage, nil, time.Now),
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

	app.JSONSerializer = rest.NewGoccyEchoSerializer()
	app.HideBanner = true
	app.HidePort = true

	configureMiddlewares(app, cm)
	configureRoutes(app, svc)

	return app
}

func configureMiddlewares(e *echo.Echo, cm *Common) {
	e.Use(middleware.CORS())
	e.Use(middleware.RequestIDWithConfig(app_middleware.RequestIDConfig()))
	e.Use(app_middleware.OpenTelemetry())
	e.Use(app_middleware.RequestMetrics())
	e.Use(app_middleware.NewLogger(cm.Logger))
	e.Use(middleware.Recover())
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
