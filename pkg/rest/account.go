package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"

	"github.com/lrweck/clean-api/internal/account"
)

type POSTAccountRequest struct {
	Name            string          `json:"name"`
	Document        string          `json:"document"`
	StartingBalance decimal.Decimal `json:"starting_balance"`
}

type GETAccountResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Document  string          `json:"document"`
	Balance   decimal.Decimal `json:"starting_balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdateAt  time.Time       `json:"updated_at,omitempty"`
}

type AccountService interface {
	New(ctx context.Context, a account.NewAccount) (*uuid.UUID, error)
	Retrieve(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

func V1POSTAccount(svc AccountService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req POSTAccountRequest
		if err := c.Bind(&req); err != nil {
			return err
		}

		id, err := svc.New(c.Request().Context(), account.NewAccount{
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
		c.JSON(http.StatusBadRequest, echo.Map{
			"message": errval.Error(),
			"details": errval.Unwrap(),
		})
		return err
	}

	c.JSON(http.StatusInternalServerError, echo.Map{
		"message": http.StatusText(http.StatusInternalServerError),
	})
	return err

}

func V1GETAccount(svc AccountService) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, echo.Map{
				"message": "invalid account id",
				"details": []string{"account id must be a valid uuid"},
			})
			return err
		}

		acc, err := svc.Retrieve(c.Request().Context(), id)
		if err != nil {
			return handleGetAccountErrors(c, err)
		}

		response := GETAccountResponse{
			ID:        acc.ID,
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
		c.JSON(http.StatusNotFound, echo.Map{
			"message": "account not found",
		})
		return err
	}

	c.JSON(http.StatusInternalServerError, echo.Map{
		"message": http.StatusText(http.StatusInternalServerError),
	})
	return err

}
