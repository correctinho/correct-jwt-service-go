package api

// PostEncodeRequest é a estrutura de solicitação para a codificação de dados.
type PostEncodeRequest struct {
	Data    interface{} `json:"data,omitempty"`    // Dados a serem codificados.
	Extras  interface{} `json:"extras,omitempty"`  // Informações extras a serem codificadas.
	Seconds int         `json:"seconds,omitempty"` // Número de segundos que o token será válido.
}

// PostDecodeRequest é a estrutura de solicitação para a decodificação de um token.
type PostDecodeRequest struct {
	Token string `json:"token,omitempty"` // Token a ser decodificado.
}
