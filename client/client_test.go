package client

import (
	"context"
	"testing"

	"github.com/3xcellent/github-metrics/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_New(t *testing.T) {
	t.Run("when token is blank", func(t *testing.T) {
		testConfig := config.APIConfig{}
		_, actErr := New(context.Background(), testConfig)

		t.Run("returns expected error", func(t *testing.T) {
			assert.EqualErrorf(t, actErr, ErrAccessTokenNotSet, "Error expected: %q, got %q", ErrAccessTokenNotSet, actErr.Error())
		})
	})

	t.Run("when token is provided", func(t *testing.T) {
		testConfig := config.APIConfig{
			Token: "github access token",
		}
		actClient, actErr := New(context.Background(), testConfig)

		t.Run("returns expected client", func(t *testing.T) {
			require.NoError(t, actErr)
			assert.Equal(t, "api.github.com", actClient.c.BaseURL.Host)
			assert.Equal(t, "uploads.github.com", actClient.c.UploadURL.Host)
		})
	})

	t.Run("when BaseURL is provided", func(t *testing.T) {
		testConfig := config.APIConfig{
			Token:   "github access token",
			BaseURL: "https://myserver.com/",
		}
		actClient, actErr := New(context.Background(), testConfig)

		t.Run("returns expected client", func(t *testing.T) {
			require.NoError(t, actErr)
			assert.Equal(t, "https://myserver.com/api/v3/", actClient.c.BaseURL.String())
			assert.Equal(t, "https://myserver.com/api/uploads/", actClient.c.UploadURL.String())
		})
	})
}
