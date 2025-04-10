package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Server struct {
	EntServer string `json:"roost_ent_server"`
	AuthToken string `json:"roost_auth_token"`
	JwtToken string `json:"roost_jwt_token"`
}

type UserConfigInfo struct{
	EntServer string `json:"roost_ent_server"`
	AuthToken string `json:"roost_auth_token"`
	//JwtToken string `json:"roost_jwt_token"`
}

type Userinfo struct{
	JwtToken string `json:"roost_jwt_token"`
	AppId string `json:"appid"`
	Response string `json:"message"`
}

func New(options ...func(*Server)) *Server {
	srv := &Server{}
	for _, o := range options {
		o(srv)
	}
	return srv
}

func WithEntServer(address string) func(*Server) {
	return func(s *Server) {
		s.EntServer = address
	}
}

func WithAuthToken(token string) func(*Server) {
	return func(s *Server) {
		s.AuthToken = token
	}
}

func WithJwtToken(token string) func(*Server) {
	return func(s *Server) {
		s.JwtToken = token
	}
}

// LoadServerFromViper returns error if unable to load token and ent server configuration
func LoadServerFromViper() error {

	var errMsg string
	if viper.Get("roost_auth_token") == nil || viper.Get("roost_auth_token").(string) == "" {
		errMsg += "Missing Auth Token. "
	}
	if viper.Get("roost_ent_server") == nil || viper.Get("roost_ent_server").(string) == "" {
		errMsg += "Missing Ent Server. "
	}
	// if viper.Get("roost_jwt_token") == nil || viper.Get("roost_jwt_token").(string) == "" {
	// 	errMsg += "Missing Jwt Token. "
	// }

	if errMsg != "" {
		errMsg += "\nPlease run the 'roost login' command to log into roost or set up the config file manually using the 'roost configure' command."
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}
