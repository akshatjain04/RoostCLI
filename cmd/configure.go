/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ZB-io/internal/roostcli/pkg/config"
	"github.com/ZB-io/internal/roostcli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Roost cli",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		var roostTokenPromptContent, roostServerPromptContent string

		var cfginput config.Server

		var userinput config.UserConfigInfo 

		var entServer, authToken string

		if viper.Get("roost_auth_token") != nil && viper.Get("roost_auth_token").(string) != "" {
			authToken = viper.Get("roost_auth_token").(string)
			roostTokenPromptContent = authToken
		} else {
			roostTokenPromptContent = ""
		}

		if viper.Get("roost_ent_server") != nil && viper.Get("roost_ent_server").(string) != "" {
			entServer = viper.Get("roost_ent_server").(string)
			roostServerPromptContent = entServer
		} else {
			roostServerPromptContent = ""
		}

		// if viper.Get("roost_jwt_token") != nil && viper.Get("roost_jwt_token").(string) != "" {
		// 	jwtToken = viper.Get("roost_jwt_token").(string)
		// 	roostJWTPromptContent = jwtToken
		// } else {
		// 	roostJWTPromptContent = ""
		// }

		userinput.AuthToken = roostTokenPromptContent
		userinput.EntServer = roostServerPromptContent
		//userinput.JwtToken = roostJWTPromptContent
		utils.AcceptFromPrompt(&userinput)

		cfginput.AuthToken = userinput.AuthToken
		cfginput.EntServer = userinput.EntServer

		cfg := config.New(config.WithAuthToken(cfginput.AuthToken), config.WithEntServer(cfginput.EntServer))

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		roostConfigDir := fmt.Sprintf("%s/.roost", home)
		roostConfigPath := fmt.Sprintf("%s/.roost/config", home)

		if !utils.FileOrFolderExists(roostConfigDir) {
			err := os.MkdirAll(roostConfigDir, 0755)
			cobra.CheckErr(err)
		}

		configData, err := json.MarshalIndent(cfg, "", " ")
		cobra.CheckErr(err)

		err = os.WriteFile(roostConfigPath, configData, 0644)
		cobra.CheckErr(err)

		//api call to roost-reg
		//userinfo:=getjwttoken(userinput)
		


		//cfg = config.New(config.WithAuthToken(cfginput.AuthToken), config.WithEntServer(cfginput.EntServer), config.WithJwtToken(userinfo.JwtToken))

		// home, err = os.UserHomeDir()
		// cobra.CheckErr(err)

		// roostConfigDir = fmt.Sprintf("%s/.roost", home)
		// roostConfigPath = fmt.Sprintf("%s/.roost/config", home)

		// if !utils.FileOrFolderExists(roostConfigDir) {
		// 	err := os.MkdirAll(roostConfigDir, 0755)
		// 	cobra.CheckErr(err)
		// }

		// configData, err = json.MarshalIndent(cfg, "", " ")
		// cobra.CheckErr(err)

		// err = os.WriteFile(roostConfigPath, configData, 0644)
		// cobra.CheckErr(err)
	},
}

// func getjwttoken(userinput config.UserConfigInfo)config.Userinfo{

// 			var userinfo config.Userinfo

// 			apiEndPoint := "/api/application/auth/createToken"
// 			reqBuff, err := json.Marshal(userinput)
// 			cobra.CheckErr(err)
// 			body := bytes.NewReader(reqBuff)

// 			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
// 			cobra.CheckErr(err)

// 			json.Unmarshal(resp, &userinfo)

// 			if status == http.StatusCreated {
// 				fmt.Println("Succesfully configured roostcli")
// 			} else {
// 				fmt.Println("Unable to configure roostcli", string(userinfo.Response))
// 			}

// 		return userinfo
// }

func init() {
	rootCmd.AddCommand(configureCmd)
}
