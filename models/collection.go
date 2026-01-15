package models

import "nextui-game-manager/shared"

type Collection struct {
	DisplayName    string
	CollectionFile string
	Games          shared.Items
}

func (c Collection) Value() interface{} {
	return c
}
