package utils

import (
	"time"

	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/BrandonKowalski/gabagool/v2/pkg/gabagool/constants"
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

	return err == nil && result != nil && result.Confirmed
}

func ConfirmBulkAction(message string) bool {
	confirm, err := gabagool.ConfirmationMessage(message, []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "X", HelpText: "Remove"},
	}, gabagool.MessageOptions{
		ImagePath:     "",
		ConfirmButton: constants.VirtualButtonX,
	})

	return err == nil && confirm != nil && confirm.Confirmed
}
