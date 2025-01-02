package config

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestParseKeys(t *testing.T) {

	pkey := `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgF8sjeK3PDamGn0icENKSjwWpuWjmPrKEoVWXIO7os5iM1CZOn7h
qC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRzgE//5ct6YtpkWUgWEOZfXZ/X
FY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNqTQj/dCin6E75Q1c/AgMBAAEC
gYAnYTsQHPs4LYB2WIKVBS80L7c8+3U4B9aj/zjmdQQHW1CaP9yZVWuOKwgzDLVt
GzJnm6Fs34PLJwzlO80RREkmEynnTYNVJejCLwuyT1oEGV6rFsql2HcIZ073NCxN
WakwL6Ay7QHH5S+hJDHCuxAx7kKoqiIXRcvbcwpRAnE5kQJBAKISE5uw1ejUocnE
ad7M2PVTz36ZS9d/3glpRQiQ2exeFRtcsq1J6O7G5OK62UMv3tcHyjf2suY0fPAA
jlPt/jcCQQCWVTylcs1Q319VRecJxSiCPjj97AA2VO1gcgzCWQ7mTp+N8QIegrD/
ZvvHqSLt79CexWrnOI6SvPuMf+8fwas5AkAlw76L7cW6bimQ4VKmFueLKs9TuZbB
jUsIuF3cpBwThsy2RoBf/rPnR7M33cAYdsQfKPKG3dZL6/kc15RSnEc7AkAgXTdS
MxXqjDw84nCr1Ms0xuqEF/Ovvrbf5Y3DpWKkyFZnO3SGVwJ96ZDY2hvP96oFFGFA
aBehlZfeFojHYG1ZAkEAnfwWAoPmvHxDaakOMsZg9PVVHIMhJ3Uck7lU5HKofHhq
rW4FGtaAhyoIZ2DQgctfe+PMcflOzkzkg9Cpqax7Cg==
-----END RSA PRIVATE KEY-----`

	pubKey := `-----BEGIN PUBLIC KEY-----
MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgF8sjeK3PDamGn0icENKSjwWpuWj
mPrKEoVWXIO7os5iM1CZOn7hqC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRz
gE//5ct6YtpkWUgWEOZfXZ/XFY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNq
TQj/dCin6E75Q1c/AgMBAAE=
-----END PUBLIC KEY-----`

	tests := []struct {
		name       string
		privateKey string
		publicKey  string
		expectErr  error
	}{
		{
			name:       "missing_private_key",
			privateKey: "",
			publicKey:  "validPublicKey",
			expectErr:  ErrNoPrivateKey,
		},
		{
			name:       "missing_public_key",
			privateKey: "validPrivateKey",
			publicKey:  "",
			expectErr:  ErrNoPublicKey,
		},
		{
			name:       "invalid_private_key_format",
			privateKey: "invalidPrivateKey",
			publicKey:  "validPublicKey",
			expectErr:  jwt.ErrKeyMustBePEMEncoded,
		},
		{
			name:       "invalid_public_key_format",
			privateKey: "validPrivateKey",
			publicKey:  "invalidPublicKey",
			expectErr:  jwt.ErrKeyMustBePEMEncoded,
		},
		{
			name:       "valid_keys",
			privateKey: pkey,
			publicKey:  pubKey,
			expectErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := parseKeys(tt.privateKey, tt.publicKey)
			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, keys)
			}
		})
	}
}

func TestBindArg(t *testing.T) {
	tests := []struct {
		name        string
		getFlags    func() []string
		getEnv      func() map[string]string
		expectErr   bool
		expectedCnf CliConfig
	}{
		{
			name: "valid_command_flags",
			getFlags: func() []string {
				return []string{"-a=someAddress", "-l=info", "-d=testDSN", "-h=hash", "-r=123", "-b=123", "-t=123s", "-w=/upload"}
			},
			getEnv: func() map[string]string {
				return map[string]string{}
			},
			expectErr: false,
			expectedCnf: CliConfig{
				Address:         "someAddress",
				LogLevel:        "info",
				DatabaseDSN:     "testDSN",
				HashKey:         "hash",
				PrivateJWTKey:   "123",
				PublicJWTKey:    "123",
				TokenExpiration: 123 * time.Second,
				UploadPath:      "/upload",
			},
		},
		{
			name: "valid_env",
			getFlags: func() []string {
				return []string{}
			},
			getEnv: func() map[string]string {
				return map[string]string{
					"RUN_ADDRESS":      "someAddress",
					"LOG_LEVEL":        "info",
					"DATABASE_URI":     "testDSN",
					"KEY":              "hash",
					"JPKEY":            "123",
					"JPUKEY":           "123",
					"TOKEN_EXPIRATION": "123s",
					"UPLOAD_PATH":      "/upload",
				}
			},
			expectErr: false,
			expectedCnf: CliConfig{
				Address:         "someAddress",
				LogLevel:        "info",
				DatabaseDSN:     "testDSN",
				HashKey:         "hash",
				PrivateJWTKey:   "123",
				PublicJWTKey:    "123",
				TokenExpiration: 123 * time.Second,
				UploadPath:      "/upload",
			},
		},
		{
			name: "valid_both",
			getFlags: func() []string {
				return []string{"-a=someAddress", "-l=info", "-d=testDSN", "-h=hash", "-r=123", "-b=123", "-t=123s", "-w=/upload"}
			},
			getEnv: func() map[string]string {
				return map[string]string{
					"ADDRESS":          "someAddress1",
					"LOG_LEVEL":        "info1",
					"DATABASE_DSN":     "testDSN1",
					"HASH_KEY":         "hash1",
					"PRIVATE_JWT_KEY":  "1231",
					"PUBLIC_JWT_KEY":   "1231",
					"TOKEN_EXPIRATION": "1231s",
					"UPLOAD_PATH":      "/upload1",
				}
			},
			expectErr: false,
			expectedCnf: CliConfig{
				Address:         "someAddress",
				LogLevel:        "info",
				DatabaseDSN:     "testDSN",
				HashKey:         "hash",
				PrivateJWTKey:   "123",
				PublicJWTKey:    "123",
				TokenExpiration: 123 * time.Second,
				UploadPath:      "/upload",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pflag.CommandLine = pflag.NewFlagSet(tt.name, pflag.ExitOnError)
			os.Clearenv()
			viper.Reset()
			os.Args = append([]string{"cmd"}, tt.getFlags()...)
			for k, v := range tt.getEnv() {
				if setErr := os.Setenv(k, v); setErr != nil {
					assert.Error(t, setErr, "Failed to set environment variable %q", k)
					return
				}
			}
			var testCnf CliConfig
			err := parseFromViper(&testCnf)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCnf, testCnf)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		getEnv      func() map[string]string
		getFlags    func() []string
		expectedErr bool
		expectedCnf CliConfig
	}{
		{
			name: "valid_config",
			getEnv: func() map[string]string {
				return map[string]string{
					"RUN_ADDRESS":  "127.0.0.1:8080",
					"LOG_LEVEL":    "debug",
					"DATABASE_URI": "postgres://user:password@localhost/dbname",
					"KEY":          "hash",
					"JPKEY": `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgF8sjeK3PDamGn0icENKSjwWpuWjmPrKEoVWXIO7os5iM1CZOn7h
qC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRzgE//5ct6YtpkWUgWEOZfXZ/X
FY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNqTQj/dCin6E75Q1c/AgMBAAEC
gYAnYTsQHPs4LYB2WIKVBS80L7c8+3U4B9aj/zjmdQQHW1CaP9yZVWuOKwgzDLVt
GzJnm6Fs34PLJwzlO80RREkmEynnTYNVJejCLwuyT1oEGV6rFsql2HcIZ073NCxN
WakwL6Ay7QHH5S+hJDHCuxAx7kKoqiIXRcvbcwpRAnE5kQJBAKISE5uw1ejUocnE
ad7M2PVTz36ZS9d/3glpRQiQ2exeFRtcsq1J6O7G5OK62UMv3tcHyjf2suY0fPAA
jlPt/jcCQQCWVTylcs1Q319VRecJxSiCPjj97AA2VO1gcgzCWQ7mTp+N8QIegrD/
ZvvHqSLt79CexWrnOI6SvPuMf+8fwas5AkAlw76L7cW6bimQ4VKmFueLKs9TuZbB
jUsIuF3cpBwThsy2RoBf/rPnR7M33cAYdsQfKPKG3dZL6/kc15RSnEc7AkAgXTdS
MxXqjDw84nCr1Ms0xuqEF/Ovvrbf5Y3DpWKkyFZnO3SGVwJ96ZDY2hvP96oFFGFA
aBehlZfeFojHYG1ZAkEAnfwWAoPmvHxDaakOMsZg9PVVHIMhJ3Uck7lU5HKofHhq
rW4FGtaAhyoIZ2DQgctfe+PMcflOzkzkg9Cpqax7Cg==
-----END RSA PRIVATE KEY-----`,
					"JPUKEY": `-----BEGIN PUBLIC KEY-----
MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgF8sjeK3PDamGn0icENKSjwWpuWj
mPrKEoVWXIO7os5iM1CZOn7hqC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRz
gE//5ct6YtpkWUgWEOZfXZ/XFY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNq
TQj/dCin6E75Q1c/AgMBAAE=
-----END PUBLIC KEY-----`,
					"TOKEN_EXPIRATION": "600s",
					"UPLOAD_PATH":      "/var/uploads",
				}
			},
			getFlags: func() []string {
				return []string{}
			},
			expectedErr: false,
			expectedCnf: CliConfig{
				Address:     "127.0.0.1:8080",
				LogLevel:    "debug",
				DatabaseDSN: "postgres://user:password@localhost/dbname",
				HashKey:     "hash",
				PrivateJWTKey: `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgF8sjeK3PDamGn0icENKSjwWpuWjmPrKEoVWXIO7os5iM1CZOn7h
qC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRzgE//5ct6YtpkWUgWEOZfXZ/X
FY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNqTQj/dCin6E75Q1c/AgMBAAEC
gYAnYTsQHPs4LYB2WIKVBS80L7c8+3U4B9aj/zjmdQQHW1CaP9yZVWuOKwgzDLVt
GzJnm6Fs34PLJwzlO80RREkmEynnTYNVJejCLwuyT1oEGV6rFsql2HcIZ073NCxN
WakwL6Ay7QHH5S+hJDHCuxAx7kKoqiIXRcvbcwpRAnE5kQJBAKISE5uw1ejUocnE
ad7M2PVTz36ZS9d/3glpRQiQ2exeFRtcsq1J6O7G5OK62UMv3tcHyjf2suY0fPAA
jlPt/jcCQQCWVTylcs1Q319VRecJxSiCPjj97AA2VO1gcgzCWQ7mTp+N8QIegrD/
ZvvHqSLt79CexWrnOI6SvPuMf+8fwas5AkAlw76L7cW6bimQ4VKmFueLKs9TuZbB
jUsIuF3cpBwThsy2RoBf/rPnR7M33cAYdsQfKPKG3dZL6/kc15RSnEc7AkAgXTdS
MxXqjDw84nCr1Ms0xuqEF/Ovvrbf5Y3DpWKkyFZnO3SGVwJ96ZDY2hvP96oFFGFA
aBehlZfeFojHYG1ZAkEAnfwWAoPmvHxDaakOMsZg9PVVHIMhJ3Uck7lU5HKofHhq
rW4FGtaAhyoIZ2DQgctfe+PMcflOzkzkg9Cpqax7Cg==
-----END RSA PRIVATE KEY-----`,
				PublicJWTKey: `-----BEGIN PUBLIC KEY-----
MIGeMA0GCSqGSIb3DQEBAQUAA4GMADCBiAKBgF8sjeK3PDamGn0icENKSjwWpuWj
mPrKEoVWXIO7os5iM1CZOn7hqC4OgRKArfNC2BVa2zvVcrzRxRzFobyM6fblMbRz
gE//5ct6YtpkWUgWEOZfXZ/XFY5AUlngBKZtU2MS/CUX+PFXICUIVTDoCL6ngwNq
TQj/dCin6E75Q1c/AgMBAAE=
-----END PUBLIC KEY-----`,
				TokenExpiration: 600 * time.Second,
				UploadPath:      "/var/uploads",
			},
		},
		{
			name: "invalid_rsa_private_key",
			getEnv: func() map[string]string {
				return map[string]string{
					"JPKEY":     "invalid_key",
					"LOG_LEVEL": "info",
				}
			},
			getFlags: func() []string {
				return []string{}
			},
			expectedErr: true,
			expectedCnf: CliConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.getEnv() {
				if err := os.Setenv(k, v); err != nil {
					t.Fatalf("failed to set env: %s", err)
				}
			}
			os.Args = append([]string{"cmd"}, tt.getFlags()...)
			pflag.CommandLine = pflag.NewFlagSet(tt.name, pflag.ExitOnError)
			viper.Reset()

			cfg, err := NewConfig()
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, compareConfigs(&tt.expectedCnf, cfg))
			}
		})
	}
}

func compareConfigs(expected, actual *CliConfig) bool {
	if expected.PrivateJWTKey != actual.PrivateJWTKey ||
		expected.Address != actual.Address ||
		expected.PublicJWTKey != actual.PublicJWTKey ||
		expected.LogLevel != actual.LogLevel ||
		expected.HashKey != actual.HashKey ||
		expected.UploadPath != actual.UploadPath {
		return false
	}
	return true
}
