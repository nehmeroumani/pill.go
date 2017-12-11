package auth

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strconv"
	"time"

	"path/filepath"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type JWTAuthentication struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

var (
	tokenDuration  time.Duration = 24 * 7
	privateKeyPath string
	publicKeyPath  string
	domainName     string
	secureToken    bool
)

func Init(privateKey string, publicKey string, domain string, onlyHTTPS bool, options ...time.Duration) {
	privateKeyPath = filepath.FromSlash(privateKey)
	publicKeyPath = filepath.FromSlash(publicKey)
	domainName = domain
	secureToken = onlyHTTPS
	if options != nil {
		if len(options) > 0 {
			tokenDuration = options[0]
		}
	}
	GetJWTAuth()
}

var JWTAuth *JWTAuthentication

func GetJWTAuth() *JWTAuthentication {
	if JWTAuth == nil {
		if privateKeyPath == "" || publicKeyPath == "" {
			panic("Public key or/and private key file(s) is/are not defined")
		}
		JWTAuth = &JWTAuthentication{
			privateKey: getPrivateKey(),
			publicKey:  getPublicKey(),
		}
	}
	return JWTAuth
}

func (this *JWTAuthentication) GenerateToken(userId int, role string) (string, error) {
	claims := &Claims{}
	claims.ExpiresAt = time.Now().Add(time.Hour * tokenDuration).Unix()
	claims.IssuedAt = time.Now().Unix()
	claims.Subject = strconv.Itoa(userId)
	claims.Role = role
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	tokenString, err := token.SignedString(this.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func HashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return ""
	}
	return string(hashedPassword[:])
}

func CompareHashAndPassword(password string, userPassword string) bool {
	if userPassword != "" {
		err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(password))
		if err == nil {
			return true
		}
	}
	return false
}

func getKeyData(keyPath string) *pem.Block {
	keyFile, err := os.Open(keyPath)
	defer keyFile.Close()
	if err != nil {
		panic(err)
	}
	pemFileInfo, _ := keyFile.Stat()
	size := pemFileInfo.Size()
	buffer := bufio.NewReader(keyFile)
	pemBytes := make([]byte, size)
	buffer.Read(pemBytes)

	data, _ := pem.Decode(pemBytes)
	return data
}

func getPublicKey() *rsa.PublicKey {
	data := getKeyData(publicKeyPath)
	keyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPK, ok := keyImported.(*rsa.PublicKey)

	if !ok {
		panic(ok)
	}
	return rsaPK
}

func getPrivateKey() *rsa.PrivateKey {
	data := getKeyData(privateKeyPath)
	keyImported, err := x509.ParsePKCS8PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPK, ok := keyImported.(*rsa.PrivateKey)

	if !ok {
		panic(ok)
	}
	return rsaPK
}
