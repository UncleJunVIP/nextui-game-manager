package utils

import (
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool/constants"
)

func ShowTimedMessage(message string, delay time.Duration) {
	gabagool.ProcessMessage(message, gabagool.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(delay)
		return nil, nil
	})
}

func ConfirmAction(message string) bool {
	result, err := gabagool.ConfirmationMessage(message, []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "I Changed My Mind"},
		{ButtonName: "A", HelpText: "Yes"},
	}, gabagool.MessageOptions{})

	return err == nil && result.IsSome()
}

func ConfirmBulkAction(message string) bool {
	confirm, _ := gabagool.ConfirmationMessage(message, []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "X", HelpText: "Remove"},
	}, gabagool.MessageOptions{
		ImagePath:     "",
		ConfirmButton: constants.VirtualButtonX,
	})

	return confirm.IsSome() && !confirm.Unwrap().Cancelled
}
