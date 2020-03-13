package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInvalidConfigWithoutGitlabUrl(t *testing.T) {
	config := Config{}
	assert.Nil(t, config.Validate())
}

func TestValidateInvalidUrlFormat(t *testing.T) {
	config := Config{
		GitlabUrl: "https://This is not an URL!",
	}
	assert.NotNil(t, config.Validate())
}

func TestValidateInvalidConfigWithInvalidHttpCredentials(t *testing.T) {
	config := Config{
		GitlabUrl:       "https://gitlab.com",
		HttpCredentials: "invalid credentials",
	}
	assert.NotNil(t, config.Validate())
}

func TestValidate(t *testing.T) {
	config := Config{
		GitlabUrl:       "https://gitlab.com",
		HttpCredentials: "username:password",
	}
	assert.Nil(t, config.Validate())
}

func TestIsVendorAllowedEmptyWhitelist(t *testing.T) {
	config := Config{}
	assert.True(t, config.IsVendorAllowed("vendor"))
}

func TestIsVendorAllowedValidVendor(t *testing.T) {
	config := Config{
		VendorWhitelist: []string{"atomicptr"},
	}
	assert.True(t, config.IsVendorAllowed("atomicptr"))
}

func TestIsVendorAllowedInvalidVendor(t *testing.T) {
	config := Config{
		VendorWhitelist: []string{"atomicptr"},
	}
	assert.False(t, config.IsVendorAllowed("vendor"))
}

func TestGetHttpCredentialsInvalid(t *testing.T) {
	config := Config{
		HttpCredentials: "invalid credentials",
	}

	username, password := config.GetHttpCredentials()

	assert.Empty(t, username)
	assert.Empty(t, password)
}

func TestGetHttpCredentialsSimple(t *testing.T) {
	config := Config{
		HttpCredentials: "username:password",
	}

	username, password := config.GetHttpCredentials()

	assert.EqualValues(t, username, "username")
	assert.EqualValues(t, password, "password")
}

func TestGetHttpCredentialsComplexPassword(t *testing.T) {
	config := Config{
		HttpCredentials: "Cr4zyU$3rn4m3:P4$$W:@rd!!?",
	}

	username, password := config.GetHttpCredentials()

	assert.EqualValues(t, username, "Cr4zyU$3rn4m3")
	assert.EqualValues(t, password, "P4$$W:@rd!!?")
}
