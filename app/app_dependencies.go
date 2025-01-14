package app

import (
	"context"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
)

type ApplicationServices struct {
	checkpointBuffer *request.DefaultBuffer
}

func (app *ApplicationServices) WithBuffer(ctx context.Context) *ApplicationServices {
	if app.checkpointBuffer == nil {
		app.checkpointBuffer = request.NewDefaultBuffer(ctx, nil)
	}

	return app
}

func (app *ApplicationServices) CheckpointBuffer() *request.DefaultBuffer {
	return app.checkpointBuffer
}
