package api

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/correctinho/correct-mlt-go/qlog"
	types "github.com/correctinho/correct-types-sdk-go"
	"github.com/correctinho/correct-types-sdk-go/chain"
	_err "github.com/correctinho/correct-types-sdk-go/err"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// PostDecode - cdeodificando um JWT
func (srv *Service) PostDecode(ctx *gin.Context) {
	logger := qlog.NewProduction(ctx)
	defer logger.Sync()

	// pkcs5UnPadding - alinhando tamanho do bloco
	pkcs5UnPadding := func(origData []byte) []byte {
		length := len(origData)
		unPadding := int(origData[length-1])
		return origData[:(length - unPadding)]
	}

	// decrypt - descriptografa um dado
	decrypt := func(data interface{}, cryptoKey string) (interface{}, error) {
		aesKey := []byte(cryptoKey)
		payload := data.(string)
		content, _ := hex.DecodeString(payload)
		block, e := aes.NewCipher(aesKey)
		if e != nil {
			return "", e
		}
		if len(aesKey) < aes.BlockSize {
			return "", errors.New("cipherText < BlockSize")
		}

		iv := make([]byte, aes.BlockSize)
		if len(content)%aes.BlockSize != 0 {
			return "", errors.New("cipherText % BlockSize")
		}
		mode := cipher.NewCBCDecrypter(block, iv)
		mode.CryptBlocks(content, content)
		content = pkcs5UnPadding(content)

		isJSON := func(s string) bool {
			var js map[string]interface{}
			return json.Unmarshal([]byte(s), &js) == nil
		}

		if isJSON(string(content)) {
			var raw map[string]interface{}
			if e := json.Unmarshal(content, &raw); e != nil {
				return "", e
			}
			return raw, nil
		}
		return string(content), nil
	}

	// verify - Verifica assinatura de um token
	verify := func(token string, password string) (*types.JwtToken, error) {
		item, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Token inválido")
			}
			return []byte(password), nil
		})
		if err != nil {
			return nil, err
		}
		claims, ok := item.Claims.(jwt.MapClaims)
		if !ok {
			return nil, errors.New("Token inválido")
		}
		decodeM, e := json.Marshal(claims)
		if e != nil {
			return nil, e
		}
		var jwtUser types.JwtToken
		if e := json.Unmarshal(decodeM, &jwtUser); e != nil {
			return nil, e
		}
		return &jwtUser, nil
	}

	decode := func(token string) (*types.JwtToken, error) {
		jwtParse := jwt.Parser{SkipClaimsValidation: false}
		tokens, _, e := jwtParse.ParseUnverified(token, jwt.MapClaims{})
		if e != nil {
			return nil, e
		}
		decodeM, e := json.Marshal(tokens.Claims)
		if e != nil {
			return nil, e
		}
		var jwt types.JwtToken
		if e := json.Unmarshal(decodeM, &jwt); e != nil {
			return nil, e
		}
		return &jwt, nil
	}

	var request PostDecodeRequest
	if e := ctx.BindJSON(&request); e != nil {
		logger.Error(e.Error())
		chain.ResponseError(ctx, &_err.ErrRequiredData)
		return
	}

	jwt, e := decode(request.Token)
	if e != nil {
		logger.Error(e.Error())
		chain.ResponseError(ctx, &_err.ErrInvalidAuthenticationToken)
		return
	}

	n, _ := strconv.ParseInt(os.Getenv("JWT_CURRENT_TIME"), 10, 64)
	password := os.Getenv("JWT_CURRENT_PASS")
	cryptoKey := os.Getenv("JWT_CURRENT_CRYPTO_KEY")
	if jwt.IssuedAt < n {
		password = os.Getenv("JWT_PAST_PASS")
		cryptoKey = os.Getenv("JWT_PAST_CRYPTO_KEY")
	}

	jwt, e = verify(request.Token, password)
	if e != nil {
		logger.Fatal(e.Error())
		chain.ResponseError(ctx, &_err.ErrInvalidAuthenticationToken)
		return
	}
	if !(jwt.Data == nil || (reflect.ValueOf(jwt.Data).Kind() == reflect.Ptr && reflect.ValueOf(jwt.Data).IsNil())) {
		jwt.Data, e = decrypt(jwt.Data, cryptoKey)
		if e != nil {
			logger.Error(e.Error())
			chain.ResponseError(ctx, &_err.ErrInvalidAuthenticationToken)
			return
		}
	}

	chain.Response(ctx, http.StatusOK, jwt)
}
