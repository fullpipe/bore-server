package jwt

type Factory struct {
	signer Signer
}

func NewFactory(privateKey string) (*Factory, error) {
	signer, err := NewEdDSASigner(privateKey)
	if err != nil {
		return nil, err
	}

	return &Factory{signer: signer}, nil
}

func (g *Factory) NewBuilder() *Builder {
	return NewBuilder(g.signer)
}
