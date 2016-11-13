/*-
 * Copyright 2016 Zbigniew Mandziejewicz
 * Copyright 2016 Square, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jwt

import (
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/json"
	"strings"
)

// JSONWebToken represents a JSON Web Token (as specified in RFC7519).
type JSONWebToken struct {
	err     error
	payload func(k interface{}) ([]byte, error)
	Headers []jose.Header
}

type NestedJSONWebToken struct {
	enc     *jose.JSONWebEncryption
	Headers []jose.Header
}

// Claims deserializes a JSONWebToken into dest using the provided key.
func (t *JSONWebToken) Claims(key interface{}, dest ...interface{}) error {
	if t.err != nil {
		return t.err
	}

	b, err := t.payload(key)
	if err != nil {
		return err
	}

	for _, d := range dest {
		if err := json.Unmarshal(b, d); err != nil {
			return err
		}
	}

	return nil
}

func (t *NestedJSONWebToken) Decrypt(decryptionKey interface{}) *JSONWebToken {
	b, err := t.enc.Decrypt(decryptionKey)
	if err != nil {
		return &JSONWebToken{err: err}
	}

	sig, err := ParseSigned(string(b))
	if err != nil {
		return &JSONWebToken{err: err}
	}

	return sig
}

// ParseSigned parses token from JWS form.
func ParseSigned(s string) (*JSONWebToken, error) {
	sig, err := jose.ParseSigned(s)
	if err != nil {
		return nil, err
	}
	headers := make([]jose.Header, len(sig.Signatures))
	for i, signature := range sig.Signatures {
		headers[i] = signature.Header
	}

	return &JSONWebToken{
		err:     nil,
		payload: sig.Verify,
		Headers: headers,
	}, nil
}

// ParseEncrypted parses token from JWE form.
func ParseEncrypted(s string) (*JSONWebToken, error) {
	enc, err := jose.ParseEncrypted(s)
	if err != nil {
		return nil, err
	}

	return &JSONWebToken{
		err:     nil,
		payload: enc.Decrypt,
		Headers: []jose.Header{enc.Header},
	}, nil
}

// ParseSignedAndEncrypted parses signed-then-encrypted token from JWE form.
func ParseSignedAndEncrypted(s string) (*NestedJSONWebToken, error) {
	enc, err := jose.ParseEncrypted(s)
	if err != nil {
		return nil, err
	}

	if strings.ToUpper(enc.Header.ContentType) != "JWT" {
		return nil, ErrInvalidContentType
	}

	return &NestedJSONWebToken{
		enc:     enc,
		Headers: []jose.Header{enc.Header},
	}, nil
}
