package shared

import "qlova.tech/sum"

type HostType struct {
	MEGATHREAD,
	APACHE sum.Int[HostType] // Apache is an internal type
}

var HostTypes = sum.Int[HostType]{}.Sum()
