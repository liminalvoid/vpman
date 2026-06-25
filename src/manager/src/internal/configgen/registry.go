package configgen

type Protocol interface {
	create(cg *ConfigGenerator)
	revoke(cg *ConfigGenerator)
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

func (cg *ConfigGenerator) Create() {}
func (cg *ConfigGenerator) Revoke() {}
