package transaction

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

func TestTxFromContext(t *testing.T) {
	if tx, ok := TxFromContext(context.Background()); ok || tx != nil {
		t.Fatalf("got (%v, %v), want (nil, false)", tx, ok)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var want pgx.Tx = NewMockTx(ctrl)
	if got, ok := TxFromContext(context.WithValue(context.Background(), txCtxKey{}, want)); !ok || got != want {
		t.Fatalf("got (%v, %v), want (%v, true)", got, ok, want)
	}
}
