package views

import (
	"github.com/dotzero/hooks/app/models"
)

// Common view struct
type Common struct {
	BaseURL string
	Recent  []*models.Hook
}

// Home view struct
type Home struct {
	Common Common
}

// Hook view struct
type Hook struct {
	Common Common
	Hook   *models.Hook
}
