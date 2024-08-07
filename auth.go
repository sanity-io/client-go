package sanity

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	SignatureHeaderName = "sanity-webhook-signature"
	// Sanity didn't send signed payloads prior to 2021 (2021-01-01T00:00:00.000Z)
	MinimumTimestamp = 1609459200000
)

var signatureHeaderRegex = regexp.MustCompile(`^t=(\d+)[, ]+v1=([^, ]+)$`)

func IsValidRequest(r *http.Request, webhookSecret string) (bool, error) {
	signature := r.Header.Get(SignatureHeaderName)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return false, fmt.Errorf("parse body: %w", err)
	}
	if err := r.Body.Close(); err != nil {
		return false, fmt.Errorf("close body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if valid, err := IsValidSignature(string(body), signature, webhookSecret); err != nil {
		return false, fmt.Errorf("validate signature: %w", err)
	} else if !valid {
		return false, nil
	}
	return true, nil
}

func IsValidSignature(payload string, signature string, webhookSecret string) (bool, error) {
	_, timestamp, err := decodeSignatureHeader(signature)
	if err != nil {
		return false, err
	}
	if timestamp < MinimumTimestamp {
		return false, errors.New("invalid signature timestamp, must be a unix timestamp with millisecond precision")
	}
	expectedSignature := encodeSignatureHeader(payload, timestamp, webhookSecret)
	if signature != expectedSignature {
		return false, nil
	}
	return true, nil
}

func decodeSignatureHeader(signaturePayload string) (string, int, error) {
	match := signatureHeaderRegex.FindStringSubmatch(strings.TrimSpace(signaturePayload))
	if len(match) != 3 {
		return "", 0, errors.New("invalid signature")
	}
	hashedPayload := match[2]
	timestamp, err := strconv.Atoi(match[1])
	if err != nil {
		return "", 0, err
	}
	return hashedPayload, timestamp, nil
}

func encodeSignatureHeader(
	payload string,
	timestamp int,
	secret string,
) string {
	signature := generateHS256Signature(payload, timestamp, secret)
	return fmt.Sprintf(`t=%d,v1=%s`, timestamp, signature)
}

func generateHS256Signature(payload string, timestamp int, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%d.%s", timestamp, payload)))
	return base64.RawURLEncoding.Strict().EncodeToString(h.Sum(nil))
}
