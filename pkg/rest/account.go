package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"

	"github.com/lrweck/clean-api/internal/account"
)

type POSTAccount struct {
	Name            string          `json:"name"`
	Document        string          `json:"document"`
	StartingBalance decimal.Decimal `json:"starting_balance"`
}

type AccountService interface {
	New(ctx context.Context, a account.NewAccount) (*uuid.UUID, error)
	Retrieve(ctx context.Context, id uuid.UUID) (*account.Account, error)
}

func V1POSTAccount(svc AccountService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req POSTAccount
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
				"details": []string{"must be a valid uuid"},
			})
			return err
		}

		acc, err := svc.Retrieve(c.Request().Context(), id)
		if err != nil {
			return handleGetAccountErrors(c, err)
		}

		return c.JSON(http.StatusOK, acc)
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
