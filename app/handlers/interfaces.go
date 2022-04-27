package handlers

//go:generate moq -skip-ensure -out mock.go . store tpl

import (
	"io"

	"github.com/dotzero/hooks/app/models"
)

type store interface {
	Hook(name string) (*models.Hook, error)
	PutHook(hook *models.Hook) error
	RecentHooks(max int) ([]*models.Hook, error)
	Requests(hook string) ([]*models.Request, error)
	PutRequest(hook string, req *models.Request) error
	Count(name []byte) (int, error)
}

type tpl interface {
	Execute(wr io.Writer, data interface{}) error
}
