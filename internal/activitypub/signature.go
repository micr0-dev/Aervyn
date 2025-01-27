package activitypub

import (
	"Aervyn/internal/config"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
)

type SignatureHeader struct {
	KeyId     string
	Algorithm string
	Headers   []string
	Signature string
}

func ParseSignatureHeader(header string) (*SignatureHeader, error) {
	parts := strings.Split(header, ",")
	sig := &SignatureHeader{}

	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}

		value := strings.Trim(kv[1], "\"")
		switch kv[0] {
		case "keyId":
			sig.KeyId = value
		case "algorithm":
			sig.Algorithm = value
		case "headers":
			sig.Headers = strings.Split(value, " ")
		case "signature":
			sig.Signature = value
		}
	}

	return sig, nil
}

func VerifySignature(r *http.Request) error {
	if config.Development {
		// Skip signature verification in development
		return nil
	}

	sigHeader := r.Header.Get("Signature")
	if sigHeader == "" {
		return fmt.Errorf("no signature header")
	}

	sig, err := ParseSignatureHeader(sigHeader)
	if err != nil {
		return fmt.Errorf("invalid signature header: %w", err)
	}

	// Fetch actor's public key
	actorKey, err := FetchActorKey(sig.KeyId)
	if err != nil {
		return fmt.Errorf("failed to fetch actor key: %w", err)
	}

	// Verify signature
	return VerifyHTTPSignature(r, sig, actorKey)
}

func FetchActorKey(keyId string) (*rsa.PublicKey, error) {
	// Fetch actor document
	resp, err := http.Get(keyId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var actor Actor
	if err := json.NewDecoder(resp.Body).Decode(&actor); err != nil {
		return nil, err
	}

	// Parse PEM public key
	block, _ := pem.Decode([]byte(actor.PublicKey.PublicKeyPem))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub, nil
}

func VerifyHTTPSignature(r *http.Request, sig *SignatureHeader, publicKey *rsa.PublicKey) error {
	// Build string to verify
	var signatureString strings.Builder

	for i, header := range sig.Headers {
		if i > 0 {
			signatureString.WriteString("\n")
		}

		switch header {
		case "(request-target)":
			path := strings.ToLower(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			signatureString.WriteString(fmt.Sprintf("%s: %s", header, path))
		default:
			signatureString.WriteString(fmt.Sprintf("%s: %s", header, r.Header.Get(header)))
		}
	}

	// Decode signature
	signature, err := base64.StdEncoding.DecodeString(sig.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Create hash
	h := sha256.New()
	h.Write([]byte(signatureString.String()))
	digest := h.Sum(nil)

	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest, signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}
