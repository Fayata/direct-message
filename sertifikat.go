package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func generateCertificate() {
	// 1. Buat Private Key (kunci rahasia)
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Gagal membuat private key: %v", err)
	}

	// 2. Atur detail sertifikat (berlaku 1 tahun, untuk localhost)
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My Local Dev"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"localhost"},
	}

	// 3. Generate sertifikatnya
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Gagal membuat sertifikat: %v", err)
	}

	// 4. Simpan sertifikat ke file cert.pem
	certOut, err := os.Create("cert.pem")
	if err != nil {
		log.Fatalf("Gagal membuat file cert.pem: %v", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Gagal menulis cert.pem: %v", err)
	}
	log.Println("File cert.pem berhasil dibuat.")

	// 5. Simpan private key ke file key.pem
	keyOut, err := os.Create("key.pem")
	if err != nil {
		log.Fatalf("Gagal membuat file key.pem: %v", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Gagal meng-encode private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Gagal menulis key.pem: %v", err)
	}
	log.Println("File key.pem berhasil dibuat.")
}
