package bootstrap

import (
	"context"

	emailuc "github.com/sunibgram/application/usecase/email"
	storageuc "github.com/sunibgram/application/usecase/storage"
	emailsvciface "github.com/sunibgram/domain/service/email"
	sockethubsvciface "github.com/sunibgram/domain/service/socket_hub"
	"github.com/sunibgram/infrastructure/cache"
	infraemail "github.com/sunibgram/infrastructure/email"
	"github.com/sunibgram/infrastructure/openai"
	infrasocket "github.com/sunibgram/infrastructure/socket"
	"github.com/sunibgram/infrastructure/storage"
)

const emailWorkerCount = 4

type External struct {
	Storage   *storageuc.StorageUsecase
	Email     emailsvciface.IEmailService
	OpenAI    *openai.Client
	SocketHub sockethubsvciface.IHub
}

func NewExternal(c *cache.Cache) *External {
	return &External{
		Storage:   storageuc.NewStorageUsecase(storage.NewMinioClient()),
		Email:     emailuc.NewEmailService(infraemail.NewSMTPClient(), c),
		OpenAI:    openai.NewClient(),
		SocketHub: infrasocket.NewHub(),
	}
}

func StartEmailWorkers(svc emailsvciface.IEmailService) {
	for i := 0; i < emailWorkerCount; i++ {
		go svc.StartWorker(context.Background())
	}
}
