package memorydb

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/lrweck/clean-api/internal/account"
)

func BenchmarkAccountCreate(b *testing.B) {

	accStorage := NewAccountStorage()

	acc := account.Account{
		ID:        uuid.New(),
		Name:      "luis roberto",
		Document:  "123123123",
		Balance:   decimal.NewFromInt(123123123),
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}

	b.RunParallel(func(pb *testing.PB) {

		for pb.Next() {
			accStorage.CreateAccount(context.Background(), acc)
		}

	})

}
