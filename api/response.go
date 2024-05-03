package api

// PostEncodeResponse é a estrutura de resposta para a solicitação de codificação de dados.
type PostEncodeResponse struct {
	Token string `json:"token,omitempty"` // Token gerado após a codificação dos dados.
	Exp   int64  `json:"exp,omitempty"`   // Número de segundos que o token será válido.
}
