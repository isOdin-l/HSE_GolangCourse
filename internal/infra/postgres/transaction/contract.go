package transaction

//go:generate go tool mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}
//go:generate go tool mockgen -destination mocks_pgx_test.go -package ${GOPACKAGE} github.com/jackc/pgx/v5 Tx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txPool interface {
	BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
}
