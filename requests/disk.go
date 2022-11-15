package requests

import (
	"encoding/json"
)

type FTPCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Path     string `json:"path"`
}

type OAuthCredentials struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type DiskCreateRequest struct {
	Name         string         `json:"name" binding:"required,gte=1,lte=64"`
	TotalSpace   uint64         `json:"totalSpace" binding:"required,min=1"`
	ProviderUUID string         `json:"providerUUID" binding:"required"`
	VolumeUUID   string         `json:"volumeUUID" binding:"required"`
	Credentials  FTPCredentials `json:"credentials" binding:"required"`
}

type OAuthRequest struct {
	Code string `json:"code" binding:"required"`
}

type DiskUpdateRequest struct {
	Name        string         `json:"name" binding:"required,gte=1,lte=64"`
	TotalSpace  uint64         `json:"totalSpace" binding:"required,min=1"`
	Credentials FTPCredentials `json:"credentials" binding:"required"`
}

// ToString - convert FTP credentials to JSON string
//
// return type:
//   - string: credentials converted to JSON string
func (cred *FTPCredentials) ToString() string {
	if cred.Port == "" || cred.Login == "" || cred.Host == "" {
		return ""
	}

	ret, _ := json.Marshal(cred)
	return string(ret)
}

// StringToFTPCredentials - convert JSON string to FTP credentials
//
// params:
//   - cred string: JSON representation of FTP credentials
//
// return type:
//   - *FTPCredentials: converted FTP credentials
func StringToFTPCredentials(cred string) *FTPCredentials {
	var ret *FTPCredentials = &FTPCredentials{}
	_ = json.Unmarshal([]byte(cred), ret)

	return ret
}

// StringToOAuthCredentials - convert JSON string to OAuth credentials
//
// params:
//   - cred string: JSON representation of OAuth credentials
//
// return type:
//   - *OAuthCredentials: converted OAuth credentials
func StringToOAuthCredentials(cred string) *OAuthCredentials {
	var ret *OAuthCredentials = &OAuthCredentials{}
	_ = json.Unmarshal([]byte(cred), ret)

	return ret
}
