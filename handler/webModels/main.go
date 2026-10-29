package webModels

type SaveRequest struct {
	Message           []byte `json:"message"`
	TransmissionNonce []byte `json:"nonce"`
	Hash              string `json:"hash"`
	PublicKey         []byte `json:"publicKey"`
	Time              int    `json:"time"`
	CostFactor        int    `json:"costFactor"`
}

type ReadRequest struct {
	Hash      string `json:"hash"`
	PublicKey []byte `json:"publicKey"`
}

type DeleteRequest struct {
	Hash string `json:"hash"`
}

type ReadResponse struct {
	Message    []byte `json:"message"`
	CostFactor int    `json:"costFactor,omitempty"`
}
