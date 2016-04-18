package snakepit

import (
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
