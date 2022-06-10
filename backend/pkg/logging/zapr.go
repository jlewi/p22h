package logging

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func InitLogger(lvl string, dev bool) (*logr.Logger, error) {
	c := zap.NewProductionConfig()

	if dev {
		c = zap.NewDevelopmentConfig()
	}

	zapLvl := zap.NewAtomicLevel()
	err := zapLvl.UnmarshalText([]byte(lvl))
	if err != nil {
		return nil, errors.Wrapf(err, "Could not convert level %v to ZapLevel", lvl)
	}

	c.Level = zapLvl
	newLogger, err := c.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger; error %v", err))
	}

	zap.ReplaceGlobals(newLogger)

	logR := zapr.NewLogger(newLogger)
	return &logR, nil
}
