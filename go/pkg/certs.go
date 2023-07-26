package pkg

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	mrand "math/rand"
	"net"
	"os"
	"time"
)

func GenCerts(dirName string) error {

	// WORK IN PROGRESS!

	// setup authority
	rootCert := &x509.Certificate{
		SerialNumber: big.NewInt(int64(mrand.Intn(500))),
		Subject: pkix.Name{
			Organization:  []string{"Rancher"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Green Pastures"},
			StreetAddress: []string{"123 Cattle Drive"},
			PostalCode:    []string{"94016"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 3),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		SubjectKeyId: []byte{5, 4, 1, 6, 8},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	certPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	rootPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// server cert
	certBytes, err := x509.CreateCertificate(rand.Reader, rootCert, rootCert, &certPriv.PublicKey, rootPrivKey)
	if err != nil {
		return err
	}

	certPem := new(bytes.Buffer)
	certPrivPem := new(bytes.Buffer)
	if err = pem.Encode(certPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}); err != nil {
		return err
	}
	if err = pem.Encode(certPrivPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPriv),
	}); err != nil {
		return err
	}

	// write the keys to disk
	os.WriteFile(fmt.Sprintf("%s/%s/%s", baseDir, dirName, "cert.crt"), certPem.Bytes(), os.ModePerm)
	os.WriteFile(fmt.Sprintf("%s/%s/%s", baseDir, dirName, "key.key"), certPrivPem.Bytes(), os.ModePerm)

	// Generate client certificate
	clientPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	var clientTemplate = x509.Certificate{
		SerialNumber:   big.NewInt(1),
		NotBefore:      time.Now().Add(-10 * time.Second),
		NotAfter:       time.Now().AddDate(10, 0, 0),
		KeyUsage:       x509.KeyUsageCRLSign,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IsCA:           false,
		MaxPathLenZero: true,
		IPAddresses:    []net.IP{net.IP([]byte{10, 10, 0, 4})},
		DNSNames:       []string{},
	}

	clientCertBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, rootCert, &clientPrivKey.PublicKey, rootPrivKey)
	if err != nil {
		panic("Failed to create certificate:" + err.Error())
	}

	clientCertPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	clientCertPem := new(bytes.Buffer)
	clientCertPrivPem := new(bytes.Buffer)
	if err = pem.Encode(clientCertPem, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertBytes,
	}); err != nil {
		return err
	}
	if err = pem.Encode(clientCertPrivPem, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientCertPriv),
	}); err != nil {
		return err
	}

	os.WriteFile(fmt.Sprintf("%s/%s/%s", baseDir, dirName, "client-cert.crt"), clientCertPem.Bytes(), os.ModePerm)
	os.WriteFile(fmt.Sprintf("%s/%s/%s", baseDir, dirName, "client-key.key"), clientCertPrivPem.Bytes(), os.ModePerm)

	return nil
}
