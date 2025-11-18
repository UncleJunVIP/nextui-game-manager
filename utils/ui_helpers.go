package utils

import (
	"time"

	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
)

func ShowTimedMessage(message string, delay time.Duration) {
	gaba.ProcessMessage(message, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(delay)
		return nil, nil
	})
}

func ConfirmAction(message string) bool {
	result, err := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "I Changed My Mind"},
		{ButtonName: "A", HelpText: "Yes"},
	}, gaba.MessageOptions{})

	return err == nil && result.IsSome()
}

func ConfirmBulkAction(message string) bool {
	confirm, _ := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "X", HelpText: "Remove"},
	}, gaba.MessageOptions{
		ImagePath:     "",
		ConfirmButton: gaba.InternalButtonX,
	})

	return confirm.IsSome() && !confirm.Unwrap().Cancelled
}
