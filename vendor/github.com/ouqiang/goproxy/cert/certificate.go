// Copyright 2018 ouqiang authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package cert 证书管理
package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"

	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

var (
	defaultRootCAPem = []byte(`-----BEGIN CERTIFICATE-----
MIICJzCCAcygAwIBAgIITWWCIQf8/VIwCgYIKoZIzj0EAwIwUzEOMAwGA1UEBhMF
Q2hpbmExDzANBgNVBAgTBkZ1SmlhbjEPMA0GA1UEBxMGWGlhbWVuMRAwDgYDVQQK
EwdHb3Byb3h5MQ0wCwYDVQQDEwRNYXJzMB4XDTIyMDMyNTA1NDgwMFoXDTQyMDQy
NTA1NDgwMFowUzEOMAwGA1UEBhMFQ2hpbmExDzANBgNVBAgTBkZ1SmlhbjEPMA0G
A1UEBxMGWGlhbWVuMRAwDgYDVQQKEwdHb3Byb3h5MQ0wCwYDVQQDEwRNYXJzMFkw
EwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEf0mhVJmuTmxnLimKshdEE4+PYdxvBfQX
mRgsFV5KHHmxOrVJBFC/nDetmGowkARShWtBsX1Irm4w6i6Qk2QliKOBiTCBhjAO
BgNVHQ8BAf8EBAMCAQYwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMBIG
A1UdEwEB/wQIMAYBAf8CAQIwHQYDVR0OBBYEFBI5TkWYcvUIWsBAdffs833FnBrI
MCIGA1UdEQQbMBmBF3FpbmdxaWFubHVkYW9AZ21haWwuY29tMAoGCCqGSM49BAMC
A0kAMEYCIQCk1DhW7AmIW/n/QLftQq8BHZKLevWYJ813zdrNr5kXlwIhAIVvqglY
9BkYWg4NEe/mVO4C5Vtu4FnzNU9I+rFpXVSO
-----END CERTIFICATE-----
`)
	defaultRootKeyPem = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIAXeEHO0FtFqQhTvsn/DT4g3rEos97+1Nibp9RfKOKhroAoGCCqGSM49
AwEHoUQDQgAEf0mhVJmuTmxnLimKshdEE4+PYdxvBfQXmRgsFV5KHHmxOrVJBFC/
nDetmGowkARShWtBsX1Irm4w6i6Qk2QliA==
-----END EC PRIVATE KEY-----
`)
)

var (
	defaultRootCA  *x509.Certificate
	defaultRootKey *ecdsa.PrivateKey
)

func init() {
	var err error
	block, _ := pem.Decode(defaultRootCAPem)
	defaultRootCA, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("加载根证书失败: %s", err))
	}
	block, _ = pem.Decode(defaultRootKeyPem)
	defaultRootKey, err = x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("加载根证书私钥失败: %s", err))
	}
}

// Certificate 证书管理
type Certificate struct {
	cache             Cache
	defaultPrivateKey *ecdsa.PrivateKey
}

type Pair struct {
	Cert            *x509.Certificate
	CertBytes       []byte
	PrivateKey      *ecdsa.PrivateKey
	PrivateKeyBytes []byte
}

func NewCertificate(cache Cache, useDefaultPrivateKey ...bool) *Certificate {
	c := &Certificate{
		cache: cache,
	}
	if len(useDefaultPrivateKey) > 0 && useDefaultPrivateKey[0] {
		priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		if err != nil {
			panic(err)
		}
		c.defaultPrivateKey = priv
	}

	return c
}

// GenerateTlsConfig 生成TLS配置
func (c *Certificate) GenerateTlsConfig(host string) (*tls.Config, error) {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if c.cache != nil {
		fields := strings.Split(host, ".")
		sudDomains := []string{host, strings.Join(fields[1:], ".")}
		for _, item := range sudDomains {
			// 先从缓存中查找证书
			if cert := c.cache.Get(item); cert != nil {
				tlsConf := &tls.Config{
					Certificates: []tls.Certificate{*cert},
				}

				return tlsConf, nil
			}
		}
	}
	pair, err := c.GeneratePem(host, 1, defaultRootCA, defaultRootKey)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(pair.CertBytes, pair.PrivateKeyBytes)
	if err != nil {
		return nil, err
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if c.cache != nil {
		// 缓存证书
		if len(pair.Cert.IPAddresses) > 0 {
			c.cache.Set(host, &cert)
		}
		for _, item := range pair.Cert.DNSNames {
			item = strings.TrimPrefix(item, "*.")
			c.cache.Set(item, &cert)
		}

	}

	return tlsConf, nil
}

// GeneratePem 生成证书
func (c *Certificate) GeneratePem(host string, expireDays int, rootCA *x509.Certificate, rootKey *ecdsa.PrivateKey) (*Pair, error) {
	var priv *ecdsa.PrivateKey
	var err error

	if c.defaultPrivateKey != nil {
		priv = c.defaultPrivateKey
	} else {
		priv, err = ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	}
	if err != nil {
		return nil, err
	}
	tmpl := c.template(host, expireDays)
	derBytes, err := x509.CreateCertificate(crand.Reader, tmpl, rootCA, &priv.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	serverCert := pem.EncodeToMemory(certBlock)
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	p := &Pair{
		Cert:            tmpl,
		CertBytes:       serverCert,
		PrivateKey:      priv,
		PrivateKeyBytes: serverKey,
	}

	return p, nil
}

// GenerateCA 生成根证书
func (c *Certificate) GenerateCA() (*Pair, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			CommonName:   "Mars",
			Country:      []string{"China"},
			Organization: []string{"Goproxy"},
			Province:     []string{"FuJian"},
			Locality:     []string{"Xiamen"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0),
		NotAfter:              time.Now().AddDate(20, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		EmailAddresses:        []string{"qingqianludao@gmail.com"},
	}

	derBytes, err := x509.CreateCertificate(crand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	serverCert := pem.EncodeToMemory(certBlock)

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	}
	serverKey := pem.EncodeToMemory(keyBlock)

	p := &Pair{
		Cert:            tmpl,
		CertBytes:       serverCert,
		PrivateKey:      priv,
		PrivateKeyBytes: serverKey,
	}

	return p, nil
}

func (c *Certificate) template(host string, expireYears int) *x509.Certificate {
	fv := fnv.New32a()
	_, _ = fv.Write([]byte(host))

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(int64(fv.Sum32())),
		Subject: pkix.Name{
			CommonName:   host,
			Country:      []string{"China"},
			Organization: []string{"Goproxy"},
			Province:     []string{"FuJian"},
			Locality:     []string{"Xiamen"},
		},
		NotBefore:             time.Now().AddDate(-1, 0, 0),
		NotAfter:              time.Now().AddDate(expireYears, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment,
		EmailAddresses:        []string{"qingqianludao@gmail.com"},
	}
	hosts := strings.Split(host, ",")
	for _, item := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
			continue
		}

		fields := strings.Split(item, ".")
		fieldNum := len(fields)
		for i := 0; i <= (fieldNum - 2); i++ {
			cert.DNSNames = append(cert.DNSNames, "*."+strings.Join(fields[i:], "."))
		}
		if fieldNum == 2 {
			cert.DNSNames = append(cert.DNSNames, item)
		}
	}

	return cert
}

// DefaultRootCAPem 根证书
func DefaultRootCAPem() []byte {
	return defaultRootCAPem
}
