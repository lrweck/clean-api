package memorydb

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/shopspring/decimal"

	"github.com/lrweck/clean-api/internal/account"
)

func BenchmarkAccountCreate(b *testing.B) {

	accStorage := NewAccountStorage()

	acc := account.Account{
		ID:        ulid.Make(),
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
