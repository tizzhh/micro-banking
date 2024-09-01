package suite

import (
	"context"
	"net"
	"strconv"
	"testing"

	currencyv1 "github.com/tizzhh/micro-banking/gen/go/protos/proto/currency"
	"github.com/tizzhh/micro-banking/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Suite struct {
	*testing.T
	Cfg            *config.Config
	CurrencyClient currencyv1.CurrencyClient
}

const (
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.Get()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	grpc, err := grpc.NewClient(grpcAddress(cfg.GRPC.CurrencyPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:              t,
		Cfg:            cfg,
		CurrencyClient: currencyv1.NewCurrencyClient(grpc),
	}
}

func grpcAddress(port int) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(port))
}
