package contracts_test

import (
	"context"
	"evm_event_indexer/api"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/internal/testutil"
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	router *gin.Engine
	ctx    = context.Background()
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	testutil.SetupTestConfig()

	dbManager := storage.Forge()
	if err := dbManager.Init(); err != nil {
		panic(fmt.Sprintf("failed to init database: %s\n", err))
	}

	server := api.NewServer()
	router = server.Handler.(*gin.Engine)

	code := m.Run()

	dbManager.Shutdown()
	os.Exit(code)
}
