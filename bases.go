package snakepit

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/viper"
)

type (
	Handler struct {
		Constants *viper.Viper
		JSON      *JSON
	}

	Controller struct {
		Constants *viper.Viper
		Logger    *logrus.Entry
		JSON      *JSON
	}

	Validator struct {
		Logger *logrus.Entry
	}

	Interactor struct {
		Constants *viper.Viper
		Logger    *logrus.Entry
	}

	Repository struct {
		Constants *viper.Viper
		Logger    *logrus.Entry
		JSON      *JSON
	}
)

func NewHandler(
	c *viper.Viper,
	j *JSON,
) *Handler {
	return &Handler{
		Constants: c,
		JSON:      j,
	}
}

func (h *Handler) LogTime(logger *logrus.Entry, start time.Time) {
	LogTime(logger, "Handler building", start)
}

func NewController(
	c *viper.Viper,
	l *logrus.Entry,
	j *JSON,
) *Controller {
	return &Controller{
		Constants: c,
		Logger:    l,
		JSON:      j,
	}
}

func NewValidator(
	l *logrus.Entry,
) *Validator {
	return &Validator{
		Logger: l,
	}
}

func (v *Validator) LogTime(start time.Time) {
	LogTime(v.Logger, "Validation", start)
}

func NewInteractor(
	c *viper.Viper,
	l *logrus.Entry,
) *Interactor {
	return &Interactor{
		Constants: c,
		Logger:    l,
	}
}

func NewRepository(
	c *viper.Viper,
	l *logrus.Entry,
	j *JSON,
) *Repository {
	return &Repository{
		Constants: c,
		Logger:    l,
		JSON:      j,
	}
}
