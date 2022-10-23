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

func (cred *FTPCredentials) ToString() string {
	if cred.Port == "" || cred.Login == "" || cred.Host == "" {
		return ""
	}

	ret, _ := json.Marshal(cred)
	return string(ret)
}

func StringToFTPCredentials(cred string) *FTPCredentials {
	var ret *FTPCredentials = &FTPCredentials{}
	_ = json.Unmarshal([]byte(cred), ret)

	return ret
}

func StringToOAuthCredentials(cred string) *OAuthCredentials {
	var ret *OAuthCredentials = &OAuthCredentials{}
	_ = json.Unmarshal([]byte(cred), ret)

	return ret
}

type DiskCreateRequest struct {
	Name         string         `json:"name" binding:"required,gte=1,lte=64"`
	ProviderUUID string         `json:"providerUUID" binding:"required"`
	VolumeUUID   string         `json:"volumeUUID" binding:"required"`
	Credentials  FTPCredentials `json:"credentials" binding:"required"`
}

type OAuthRequest struct {
	VolumeUUID   string `json:"volumeUUID" binding:"required"`
	DiskUUID     string `json:"diskUUID" binding:"required"`
	ProviderUUID string `json:"providerUUID" binding:"required"`
	Code         string `json:"code" binding:"required"`
}

type DiskUpdateRequest struct {
	Name        string         `json:"name"`
	Credentials FTPCredentials `json:"credentials"`
}
