package memory_test

import (
	"context"
	"testing"

	"github.com/danilov-go/gophermart/internal/repository/memory"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_Ping(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "положительный тест",
			wantErr: false,
		},
		{
			name:    "ошибка контекста",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			var ctx context.Context
			var cancel context.CancelFunc
			if tt.wantErr {
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			} else {
				ctx = context.Background()
			}
			err := storage.Ping(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
