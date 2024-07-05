package app

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(New, wire.Value(([]AppOptFunc)(nil)))
