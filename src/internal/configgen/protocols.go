package configgen

import (
	"fmt"
)

var protocols = map[string]Protocol{
	"awg": &Awg{},
	"wg": &Wg{},
	"wireguard": &Wg{},
}

func ListProtocols() map[string]Protocol {
	return protocols
}

func GetProtocol(name string) (Protocol, error) {
	if p, ok := protocols[name]; ok {
		return p, nil
	}

	return nil, fmt.Errorf("no %q protocol exists", name)
}
