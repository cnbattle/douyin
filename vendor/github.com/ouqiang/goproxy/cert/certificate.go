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
	crand "crypto/rand"
	"math/rand"
	"strings"

	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

var (
	defaultRootCAPem = []byte(`
-----BEGIN CERTIFICATE-----
MIIFdDCCA1ygAwIBAgIBATANBgkqhkiG9w0BAQsFADBZMQ4wDAYDVQQGEwVDaGlu
YTEPMA0GA1UECBMGRnVKaWFuMQ8wDQYDVQQHEwZYaWFtZW4xDTALBgNVBAoTBE1h
cnMxFjAUBgNVBAMTDWdvLW1pdG0tcHJveHkwIBcNMTgwMzE4MDkwMDQ0WhgPMjA2
ODAzMTgwOTAwNDRaMFkxDjAMBgNVBAYTBUNoaW5hMQ8wDQYDVQQIEwZGdUppYW4x
DzANBgNVBAcTBlhpYW1lbjENMAsGA1UEChMETWFyczEWMBQGA1UEAxMNZ28tbWl0
bS1wcm94eTCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBANiuppEbanTv
iCs47AFIAy+AVXDhaInal4fGmN+kG1txO4YPygKGrdjokCZtkL6ZK61izFg6BLX+
p65j8wnAPZPZr3Zu5vlcDM7baO9ddxtnXm/fACPEuMIvgmG7zxE9CeX3LY7tsq10
hg8uKMnYGTy5Ce0hkuYn8Od0yHseGFWCmaCAHIcshbvQFxPGn42X/zWrEHDEgWtG
fOlamBBTSbNza11H8udLkXlr+N+vv/P/eKjpeIf/xzPCdiUOxdD+NHCeeSgho3Sm
P0T6ia4L7MVW0XUg7CseVVh+9TddO6QefmM1+AsWU/ektD+cUMtlWoDXE8idlpoZ
cMVJfq/6Sa9nG280fCPjd4wFLqbR67BHQkoPjQ1vmRgs4xvD04m796dRPpTDepb/
xvTTMcwgAC5tur/E5SHpr8hx9X6xGPfUUMiKyBQlSgLH4V02SjAmScxqt5AWZcT/
syLHg7BhjxwBGoCwcE8zWHCJarQ0t28Z7ptyL3DXPaJ7Vd2CvLJrekvtnm9B28aU
9KOC9JL3DKzFaRrhTYb0VNLfoLV8kRJCzZI6HAwiKcAAEIXi8on6YwqLvEIxo5AL
0gTeIf/nJU2W4OY640fIdwEvcaH4Wj2bKMRaTWvQGM1TJe4hoCN/c3mVopotCb44
IGC5R0XmVImVxZmdyCXJAfY1jYrWHA2ZAgMBAAGjRTBDMA4GA1UdDwEB/wQEAwIB
BjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBSfjEyzebvckLQu+eZjlmJF
W0/ZmjANBgkqhkiG9w0BAQsFAAOCAgEAXHGvSFwtcqX6lD9rmLTPeViTIan5hG5T
sEWsPp/kI6j579OavwCr7mk4nUjsKFaOEzN78C4k6s4gDWWePoJhlsXB4KefriW4
gWevzwgRTrIlaw3EDb+fCle0HcsCI/wwxDD8eafUxEyqBGrhLJFiUIxvOcD+yP2Q
mX3Z89Pd38Qvkb9zneJdXo2wHMq0MGKlTPdE04rg1OsuPNnSwRhtem9/E4eCtumF
JoQEQtp440wpvrbZljR18Ahd+xNh6dyaD0prnrUEGsUkC1hMb3nUWmw6dZEA5rCv
8aW5ZMm9Jr7pW7yzrm8J4II1bY5v6i7+qvOFDAf1nEnVshcSCiHu6xzgtwoGtsP8
mSOquiWwiceJL6q8xh6nOD3SYm2mZwA1n7Nl3mRJE/RgbwJNkveMrmZ6CKUm3N/x
eqd5yhTLsD7sf3+d4B7i6fAZ+csccWaDuquVI9cXi2OoMKgIFeeVwJ1FCeLY0Nah
nPlNUA0h7xKeDIHtlGsSOng6uiEVVVXGS+j9V6h+Z55AsuOAoHYOBDoXfr0Y4Bww
irCRNyFcDrKoyILOOUiPxoEcclrwUBTB78JxVA8xKTbAh0aZQRZOZOz49qF4gA1d
1riiUHJIG2sD+54UEdFoR5nhZ4/RLGqQ/Kmch5VnPp7De4OzSMd/KkQDWEjR+AA1
CDPlL4gNB6s=
-----END CERTIFICATE-----
`)
	defaultRootKeyPem = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIJKQIBAAKCAgEA2K6mkRtqdO+IKzjsAUgDL4BVcOFoidqXh8aY36QbW3E7hg/K
Aoat2OiQJm2QvpkrrWLMWDoEtf6nrmPzCcA9k9mvdm7m+VwMztto7113G2deb98A
I8S4wi+CYbvPET0J5fctju2yrXSGDy4oydgZPLkJ7SGS5ifw53TIex4YVYKZoIAc
hyyFu9AXE8afjZf/NasQcMSBa0Z86VqYEFNJs3NrXUfy50uReWv436+/8/94qOl4
h//HM8J2JQ7F0P40cJ55KCGjdKY/RPqJrgvsxVbRdSDsKx5VWH71N107pB5+YzX4
CxZT96S0P5xQy2VagNcTyJ2WmhlwxUl+r/pJr2cbbzR8I+N3jAUuptHrsEdCSg+N
DW+ZGCzjG8PTibv3p1E+lMN6lv/G9NMxzCAALm26v8TlIemvyHH1frEY99RQyIrI
FCVKAsfhXTZKMCZJzGq3kBZlxP+zIseDsGGPHAEagLBwTzNYcIlqtDS3bxnum3Iv
cNc9ontV3YK8smt6S+2eb0HbxpT0o4L0kvcMrMVpGuFNhvRU0t+gtXyREkLNkjoc
DCIpwAAQheLyifpjCou8QjGjkAvSBN4h/+clTZbg5jrjR8h3AS9xofhaPZsoxFpN
a9AYzVMl7iGgI39zeZWimi0JvjggYLlHReZUiZXFmZ3IJckB9jWNitYcDZkCAwEA
AQKCAgAlLBkhLa3er7URjStXsO3y+TYvLkxL0fdK8LQLMdELp+pJPm4ubsJmQsdw
AD3jpM1ManWZ8SIbwrsrfLQWCSfHNIIYdEAlqTf9SMDAx60GQ358/Km+eSIlFhdt
AtYsI+eNzxC+w2JyxVm2Qvn2Xp89vpTIXIkh+NooKu21yVztVoFaen/qZKXwqWs8
FkgK93dt0pH4do2pRKdrNQJ/UnqDUZqqnww5x8oGJZLFdRYeGsatW5g05Jlc9NBl
3Rnsl5+Rbm5khxjOizKxd7Wk6SDOXe2DBYnef86uZuFUhScVKbIO/RQ3erYe9t+B
RiTKL/INxlf7g6VxfEnPXqNgNzTqlFaMnj77g8xYmIdEGBX7SY9oWdK1YGT/D9YH
pbFNPnufUy08652KEqARsTS3AH762upG0AhDyn+M1x76b1yJE8EsBCWGUgGZmZQ5
siTGxOkuYGbFBik3dYhqVNQjHblQ+KHk5lNz43WMka2u5+21nRch1AnxLA/iAKh+
C8s3e8vUraRmJCkSs7QNWhNEjU32Rjbcz5c6u0ZKs27ysx1zGkA2EJ9nx0i0eHPq
wEWreRk7VQk052rbn4NCxbvXZlczGlhMXU8J82GZBs1iga4lbtIU9v4BDLTZzj2V
Gsje67rP5A86TK3kdCcI0RK0cVpnwFlI3Y7SToiFcuftDAx/zQKCAQEA/4iUtZin
aqD32oiP82DV/lbijc2/pFYpjobsz831D/anI0c8Lm7wVBiSUyvSxIJsC9NC7CaM
FJy8JRI49mudyDY7kblza4i05yQR94EBzffL857vKXC4+mDK3CaiOkT91vBjnwFp
L8zMRoAsYlqJhhOgXNLBGAT8ka7KFEYNys7qLvYhYO4cfOmnTLzySZ1DcAcF+LQG
xarr3wp8M5E2lUqEIs0d5NNbNkYA6uRmoQS3FYhydlIHEEE1eH43iKabClb8AdM1
wbCzBxh+inXg/pEp0s8WafKOeVvH0ZgzSekdxi3Xe11S+8kfxiYOUmHh6yXMUSKB
+W2UKrJf+l+f0wKCAQEA2RPpzTPqEVSIK3i9mp5bmsLJAGQhHhECrrq8LWcVB5wL
S7ah28fTwD0wciuKE7Lfgd60a6coxU9bemiZZnPe+Gb4mR2w1aCGV7zsYNDbPh0O
QBqfZtC3EplpLFxCdWzk5mel4ugYytbCIU1iJYynwydl2Mbs3F2CxEa4nowQJV0/
gtbxCKJG6gIUu1xLnpm0VT5PU0wwroPgog4Oy6A+O2IIIRDZ6c6fmU43AEnM8XoY
ZBaTlSlvynoOg8ux6f3Jt1YwUSqurC8mjUy+oE/mjSpWLEGDw8fgSTdBGq6ekOYP
mmCOPBYXE+nYjFAFGJuUxfJzf8uCTvrqWuIXKcFlYwKCAQEAqCbCZPV9RaeDMiUn
ROp2JxYZo2K/N28TjZyv/Nb06npO5eIcchnCwDQjJePyoCmK3AU7RpbfGzlAfcyN
+2o5u+QkMvKsRxkAohGUWSBlhZoIddoiW0y4DNrg4xnxKxL3TxeFFr8g7rl/uuzh
SB932+jSYAK32gx9/4fbppeqv8iFRj3lHRnTWUeQNekoLtTz6aZVgaFFy5F8AZuu
u2hVWMxeQ2BiyY9juEU8mVWPS2oE6ICPgdjcmQ+wFghIlv27jIRM9Q59k2WpiYPO
0WJcmmf/858eir14j9ebmArlxT9Hvn+wCpgQ4WsqI4QrbH7I4apP1xw0F2TKWYZj
rih6zQKCAQBBeIYNg9jWvT4Mjm/xEE3kkVb6LTjnzo2WkW9r6iknkGK/xSdwGAa/
djUEWilc45gRnU+hIFtllxeqBZ4ujkfzd2sHEzNgWvfpwmswkA1v4GeJ4f2tjsmI
bIiR/ol0zREEhMI9e27uznLihGpTlOaML3fCN8z8cZ+c/w9zkh7Uhhk/pwAvcHIe
5d3G3IFaJlWDWDWok9Qi7ldzyPWhaIUcd+anwmNW5yCvpi1kgt2y/vYYSc7dMBAt
az6xdWAFiKusBeywrkTcXaQs/baIt1B7xwcSdff9tmzo6CdUmtHsNdcC4phDew4e
zWqodwHyeAoY4ZUAOCrnEzpXitUdnNytAoIBAQCxVKnh7exzwp4VOqXxTd0dFgCM
qfk0zJDVlvV1XA+NH9tuG/cs1YTUB4oQeV3APhuPlfI70zfwt1BxbqKlc9e9eRTB
0QbBsfF4f/86y6YYBt7FlAbulNo8gaQbKHdxx65J0T/QlNFgKBoo3IepQUFshkaU
42faTEsWdQHlFALPvFrukzYA67B2mCrf0wBaS24Qx3xBzS/TJeVqJwDTE7+C2bez
bLBpLIyXUddQx8XAHeLRUtEMvY1q7O49Ruz1kHFuVrzUKzDJNU9Piqa8NtkSvApQ
4yQMI4Q3+P45ovUKxkTC+XP+qUZjML2WQaEiNn3KAK/1L5/1y/s4Weqqakgh
-----END RSA PRIVATE KEY-----
`)
)

var (
	defaultRootCA  *x509.Certificate
	defaultRootKey *rsa.PrivateKey
)

func init() {
	var err error
	block, _ := pem.Decode(defaultRootCAPem)
	defaultRootCA, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("加载根证书失败: %s", err))
	}
	block, _ = pem.Decode(defaultRootKeyPem)
	defaultRootKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(fmt.Errorf("加载根证书私钥失败: %s", err))
	}
}

// Certificate 证书管理
type Certificate struct {
	cache             Cache
	defaultPrivateKey *rsa.PrivateKey
}

type Pair struct {
	Cert            *x509.Certificate
	CertBytes       []byte
	PrivateKey      *rsa.PrivateKey
	PrivateKeyBytes []byte
}

func NewCertificate(cache Cache, useDefaultPrivateKey ...bool) *Certificate {
	c := &Certificate{
		cache: cache,
	}
	if len(useDefaultPrivateKey) > 0 && useDefaultPrivateKey[0] {
		priv, err := rsa.GenerateKey(crand.Reader, 2048)
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
		// 先从缓存中查找证书
		if cert := c.cache.Get(host); cert != nil {
			tlsConf := &tls.Config{
				Certificates: []tls.Certificate{*cert},
			}

			return tlsConf, nil
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
		c.cache.Set(host, &cert)
	}

	return tlsConf, nil
}

// Generate 生成证书
func (c *Certificate) GeneratePem(host string, expireDays int, rootCA *x509.Certificate, rootKey *rsa.PrivateKey) (*Pair, error) {
	var priv *rsa.PrivateKey
	var err error

	if c.defaultPrivateKey != nil {
		priv = c.defaultPrivateKey
	} else {
		priv, err = rsa.GenerateKey(crand.Reader, 2048)
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

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
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
	priv, err := rsa.GenerateKey(crand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			CommonName:   "Mars",
			Country:      []string{"China"},
			Organization: []string{"goproxy"},
			Province:     []string{"FuJian"},
			Locality:     []string{"Xiamen"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0),
		NotAfter:              time.Now().AddDate(1, 0, 0),
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

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
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
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(rand.Int63()),
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore:             time.Now().AddDate(-1, 0, 0),
		NotAfter:              time.Now().AddDate(expireYears, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		EmailAddresses:        []string{"qingqianludao@gmail.com"},
	}
	hosts := strings.Split(host, ",")
	for _, item := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			cert.IPAddresses = append(cert.IPAddresses, ip)
		} else {
			cert.DNSNames = append(cert.DNSNames, item)
		}
	}

	return cert
}

// RootCA 根证书
func DefaultRootCAPem() []byte {
	return defaultRootCAPem
}
