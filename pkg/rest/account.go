package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"

	"github.com/lrweck/clean-api/internal/account"
)

type POSTAccountRequest struct {
	Name            string          `json:"name"`
	Document        string          `json:"document"`
	StartingBalance decimal.Decimal `json:"starting_balance"`
}

type GETAccountResponse struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Document  string          `json:"document"`
	Balance   decimal.Decimal `json:"starting_balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdateAt  time.Time       `json:"updated_at,omitempty"`
}

type AccountService interface {
	New(ctx context.Context, a account.NewAccount) (ulid.ULID, error)
	Retrieve(ctx context.Context, id string) (*account.Account, error)
}

func V1_POST_Account(svc AccountService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req POSTAccountRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		ctx := c.Request().Context()
		id, err := svc.New(ctx, account.NewAccount{
			Name:            req.Name,
			Document:        req.Document,
			StartingBalance: req.StartingBalance,
		})

		if err != nil {
			return handlePostAccountErrors(c, err)
		}

		return c.JSON(http.StatusCreated, echo.Map{
			"id": id,
		})
	}
}

func handlePostAccountErrors(c echo.Context, err error) error {

	if errval := new(account.ErrValidation); errors.As(err, &errval) {
		c.JSON(http.StatusBadRequest, NewUserError(errval.Error(), errval.Errors()))
		return err
	}

	c.JSON(http.StatusInternalServerError, ErrInternalServerError)
	return err
}

func V1_GET_Account(svc AccountService) echo.HandlerFunc {
	return func(c echo.Context) error {

		id := c.Param("id")

		if _, err := ulid.ParseStrict(id); err != nil {
			c.JSON(http.StatusBadRequest, ErrInvalidAccountID)
			return err
		}

		ctx := c.Request().Context()

		acc, err := svc.Retrieve(ctx, id)
		if err != nil {
			return handleGetAccountErrors(c, err)
		}

		response := GETAccountResponse{
			ID:        acc.ID.String(),
			Name:      acc.Name,
			Document:  acc.Document,
			Balance:   acc.Balance,
			CreatedAt: acc.CreatedAt,
			UpdateAt:  acc.UpdateAt,
		}

		return c.JSON(http.StatusOK, response)
	}
}

func handleGetAccountErrors(c echo.Context, err error) error {

	if errors.Is(err, account.ErrNotFound) {
		c.JSON(http.StatusNotFound, ErrAccountNotFound)
		return err
	}

	c.JSON(http.StatusInternalServerError, ErrInternalServerError)
	return err

}

var (
	ErrInvalidAccountID = echo.Map{
		"message": "invalid account id",
		"details": []string{"must be a valid uuid"},
	}

	ErrAccountNotFound = echo.Map{
		"message": "account not found",
	}

	ErrInternalServerError = echo.Map{
		"message": http.StatusText(http.StatusInternalServerError),
	}
)

func NewUserError(msg string, details []string) echo.Map {
	return echo.Map{
		"message": msg,
		"details": details,
	}
}
