package api

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/correctinho/correct-mlt-go/qlog"
	types "github.com/correctinho/correct-types-sdk-go"
	"github.com/correctinho/correct-types-sdk-go/chain"
	_err "github.com/correctinho/correct-types-sdk-go/err"
)

// PostEncode - codificando dados e retornando JWT
func (srv *Service) PostEncode(ctx *gin.Context) {
	logger := qlog.NewProduction(ctx)
	defer logger.Sync()

	pkcs5Padding := func(cipherText []byte, blockSize int) []byte {
		padding := blockSize - len(cipherText)%blockSize
		padText := bytes.Repeat([]byte{byte(padding)}, padding)
		return append(cipherText, padText...)
	}

	signedString := func(expiresAt int64, password string, data, extras interface{}) (string, error) {
		claims := types.JwtToken{
			StandardClaims: &jwt.StandardClaims{
				Issuer:    "correct",
				ExpiresAt: expiresAt,
				IssuedAt:  time.Now().Unix(),
			},
			Data:   data,
			Extras: extras,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(password))
	}

	tryString := func(data interface{}) (string, error) {
		vMap, okMap := data.(map[string]interface{})
		if !okMap {
			vStr, okStr := data.(string)
			if !okStr {
				return "", errors.New("Mensagem corrompida (tipo inválido)")
			}
			return vStr, nil
		}
		body, e := json.Marshal(vMap)
		if e != nil {
			return "", e
		}
		return string(body), nil
	}

	encryptInterface := func(data interface{}) (interface{}, error) {
		aesKey := []byte(os.Getenv("JWT_CURRENT_CRYPTO_KEY"))
		content := []byte(data.(string))
		padContent := pkcs5Padding(content, aes.BlockSize)
		if len(padContent)%aes.BlockSize != 0 {
			return nil, errors.New("padContent is not a multiple of the block size")
		}
		block, e := aes.NewCipher(aesKey)
		if e != nil {
			return nil, e
		}
		cipherText := make([]byte, len(padContent))
		iv := make([]byte, aes.BlockSize)
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(cipherText[:], padContent)
		return hex.EncodeToString(cipherText), nil
	}

	encryptMap := func(data interface{}) (interface{}, error) {
		value, e := tryString(data)
		if e != nil {
			return "", e
		}
		return encryptInterface(value)
	}

	writeResponse := func(token string, exp int64) (response PostEncodeResponse) {
		response.Token = token
		response.Exp = exp
		return response
	}

	var request PostEncodeRequest
	if e := ctx.BindJSON(&request); e != nil {
		logger.Error(e.Error())
		chain.ResponseError(ctx, &_err.ErrRequiredData)
		return
	}

	var data interface{}
	if !(request.Data == nil || (reflect.ValueOf(request.Data).Kind() == reflect.Ptr && reflect.ValueOf(request.Data).IsNil())) {
		item, e := encryptMap(request.Data)
		if e != nil {
			logger.Error(e.Error())
			chain.ResponseError(ctx, &_err.ErrInternalService)
			return
		}
		data = item
	}

	seconds := request.Seconds
	if seconds == 0 {
		seconds = 300
	}

	duration := time.Duration(seconds) * time.Second

	password := os.Getenv("JWT_CURRENT_PASS")
	exp := time.Now().Add(duration).Unix()
	token, e := signedString(exp, password, data, request.Extras)
	if e != nil {
		logger.Error(e.Error())
		chain.ResponseError(ctx, &_err.ErrInternalService)
		return
	}

	response := writeResponse(token, exp)

	chain.Response(ctx, http.StatusOK, response)
}
