package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid bearer token",
			header:  "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantErr: false,
		},
		{
			name:    "valid bearer token lowercase",
			header:  "bearer secret-token-123",
			want:    "secret-token-123",
			wantErr: false,
		},
		{
			name:    "missing authorization header",
			header:  "",
			want:    "",
			wantErr: true,
			errMsg:  "missing Authorization header",
		},
		{
			name:    "only one part",
			header:  "Bearer",
			want:    "",
			wantErr: true,
			errMsg:  "invalid token format",
		},
		{
			name:    "three parts",
			header:  "Bearer some token",
			want:    "",
			wantErr: true,
			errMsg:  "invalid token format",
		},
		{
			name:    "wrong prefix",
			header:  "Basic dXNlcjpwYXNz",
			want:    "",
			wantErr: true,
			errMsg:  "invalid token format",
		},
		{
			name:    "token with extra spaces",
			header:  "Bearer  token  with  spaces",
			want:    "",
			wantErr: true,
			errMsg:  "invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractBearerToken(tt.header)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBuildHTTPClient(t *testing.T) {
	t.Run("default TLS verify", func(t *testing.T) {
		client, err := buildHTTPClient(false)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, defaultTimeout, client.Timeout)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		if transport.TLSClientConfig != nil {
			assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
		}
	})

	t.Run("skip TLS verify", func(t *testing.T) {
		client, err := buildHTTPClient(true)
		require.NoError(t, err)
		assert.NotNil(t, client)

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)
		require.NotNil(t, transport.TLSClientConfig)
		assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	})
}

func TestValidateTokenAndGetUserID(t *testing.T) {
	secret := []byte("test-secret-key-123456789012345")
	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
		"test-key": keyfunc.NewGivenHMAC(secret, keyfunc.GivenKeyOptions{}),
	})

	makeToken := func(issuer, subject string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss": issuer,
			"sub": subject,
		})
		token.Header["kid"] = "test-key"

		s, err := token.SignedString(secret)
		require.NoError(t, err)

		return s
	}

	tests := []struct {
		name           string
		tokenString    string
		expectedIssuer string
		wantUserID     string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "valid token",
			tokenString:    makeToken("https://test-issuer", "user-123"),
			expectedIssuer: "https://test-issuer",
			wantUserID:     "user-123",
			wantErr:        false,
		},
		{
			name:           "wrong issuer",
			tokenString:    makeToken("https://wrong-issuer", "user-123"),
			expectedIssuer: "https://test-issuer",
			wantUserID:     "",
			wantErr:        true,
			errContains:    "invalid issuer",
		},
		{
			name:           "missing subject",
			tokenString:    makeToken("https://test-issuer", ""),
			expectedIssuer: "https://test-issuer",
			wantUserID:     "",
			wantErr:        true,
			errContains:    "missing subject",
		},
		{
			name:           "invalid token signature",
			tokenString:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ0ZXN0Iiwic3ViIjoidXNlciJ9.invalid",
			expectedIssuer: "https://test-issuer",
			wantUserID:     "",
			wantErr:        true,
			errContains:    "invalid token",
		},
		{
			name:           "malformed token",
			tokenString:    "not-a-jwt",
			expectedIssuer: "https://test-issuer",
			wantUserID:     "",
			wantErr:        true,
			errContains:    "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateTokenAndGetUserID(tt.tokenString, jwks, tt.expectedIssuer)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUserID, got)
			}
		})
	}
}

func TestValidateTokenAndGetUserID_MissingIssuerClaim(t *testing.T) {
	secret := []byte("test-secret-key-123456789012345")
	jwks := keyfunc.NewGiven(map[string]keyfunc.GivenKey{
		"test-key": keyfunc.NewGivenHMAC(secret, keyfunc.GivenKeyOptions{}),
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
	})
	token.Header["kid"] = "test-key"

	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)

	_, err = validateTokenAndGetUserID(tokenString, jwks, "https://test-issuer")
	require.Error(t, err)
	assert.Equal(t, errors.New("invalid issuer"), err)
}
