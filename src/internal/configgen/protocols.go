package configgen

import (
	"fmt"
	"maps"
	"slices"
)

var protocols = map[string]Protocol{
	"awg": &Awg{},
	"wg": &Wg{},
}

func ListProtocols() []string {
	return slices.Collect(maps.Keys(protocols))
}

func GetProtocol(name string) (Protocol, error) {
	if p, ok := protocols[name]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("no %q protocol exists", name)
}
