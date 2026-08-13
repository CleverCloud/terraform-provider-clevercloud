package postgresql

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// selfSignedCertPEM writes a valid self-signed certificate to a temporary file
// and returns its path, to simulate a root CA discoverable by pgx.
func selfSignedCertPEM(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %s", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %s", err)
	}

	path := filepath.Join(t.TempDir(), "root.crt")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, certPEM, 0o600); err != nil {
		t.Fatalf("failed to write certificate: %s", err)
	}

	return path
}

func TestLocalePgConfig_SkipsCertificateVerification(t *testing.T) {
	cfg, err := localePgConfig("db.example.com", 5432, "mydb", "user", "password")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if cfg.TLSConfig == nil {
		t.Fatal("expected a TLS configuration (encrypted connection), got nil")
	}
	if !cfg.TLSConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify: managed databases serve self-signed certificates")
	}
	if cfg.TLSConfig.VerifyPeerCertificate != nil {
		t.Error("expected no custom certificate verification")
	}
	if len(cfg.Fallbacks) != 0 {
		t.Errorf("expected no fallback configurations, got %d", len(cfg.Fallbacks))
	}
}

// TestLocalePgConfig_IgnoresDiscoverableRootCA pins the regression behind the
// 150 seconds lost per operation on dedicated plans: pgx escalates
// sslmode=require to verify-ca when a root CA is discoverable through the
// environment, which then rejects the self-signed server certificate.
func TestLocalePgConfig_IgnoresDiscoverableRootCA(t *testing.T) {
	t.Setenv("PGSSLROOTCERT", selfSignedCertPEM(t))

	cfg, err := localePgConfig("db.example.com", 5432, "mydb", "user", "password")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if cfg.TLSConfig == nil {
		t.Fatal("expected a TLS configuration (encrypted connection), got nil")
	}
	if !cfg.TLSConfig.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify even with PGSSLROOTCERT set")
	}
	if cfg.TLSConfig.RootCAs != nil {
		t.Error("expected the discovered root CA pool to be ignored")
	}
	if cfg.TLSConfig.VerifyPeerCertificate != nil {
		t.Error("expected the verify-ca escalation to be neutralised")
	}
}

func TestIsRetryableLocaleError(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		retryable bool
	}{
		{"nil", nil, false},
		{"x509 verification", errors.New("failed to connect: x509: certificate signed by unknown authority"), false},
		{"tls handshake", errors.New("tls: handshake failure"), false},
		{"invalid password", errors.New("failed SASL auth (FATAL: password authentication failed for user (SQLSTATE 28P01))"), false},
		{"invalid authorization", errors.New("FATAL: role does not exist (SQLSTATE 28000)"), false},
		{"connection refused", errors.New("dial tcp 10.0.0.1:5432: connect: connection refused"), true},
		{"timeout", errors.New("dial tcp 10.0.0.1:5432: i/o timeout"), true},
		{"dns", errors.New("lookup db.example.com: no such host"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryableLocaleError(tc.err); got != tc.retryable {
				t.Errorf("isRetryableLocaleError(%v) = %t, expected %t", tc.err, got, tc.retryable)
			}
		})
	}
}
