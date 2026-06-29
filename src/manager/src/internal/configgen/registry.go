package configgen

type Protocol interface {
	create()
	revoke()
}

type ConfigGenerator struct {
	protocol Protocol
}

func Init(p *Protocol) *ConfigGenerator {
	return &ConfigGenerator{}
}

func (cg *ConfigGenerator) SetProtocol(p Protocol) {
	cg.protocol = p
}

func (cg *ConfigGenerator) Create() {
	cg.protocol.create()
}

func (cg *ConfigGenerator) Revoke() {}
