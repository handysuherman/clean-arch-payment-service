package repository

import (
	"testing"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/helper"
)

func TestOnConfigUpdate(t *testing.T) {
	testStore.OnConfigUpdate(helper.RandomString(12), cfg)
}

func TestOnPqConnUpdate(t *testing.T) {
	testStore.OnPqsqlUpdate(helper.RandomString(12), pqConn)
}

func TestOnRedisConnUpdate(t *testing.T) {
	testStore.OnRedisUpdate(helper.RandomString(12), rConn)
}
