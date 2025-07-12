package models

import (
	"go.uber.org/zap/zapcore"
)

type AppState struct {
	Config *Config

	MenuPositionList []MenuPositionPointer
}

func (a AppState) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("config", a.Config)

	return nil
}

type MenuPositionPointer struct {
	SelectedIndex		int
	SelectedPosition    int
}
