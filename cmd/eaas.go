package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ZB-io/internal/roostcli/pkg/eaas"
	"github.com/ZB-io/internal/roostcli/pkg/spinner"
	"github.com/ZB-io/internal/roostcli/pkg/utils"
	bubbletable "github.com/charmbracelet/bubbles/table"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// eaasCmd represents the eaas command
var eaasCmd = &cobra.Command{
	Use:   "eaas",
	Short: "CLI interface to interact with roost EAAS",
	Long:  "Use 'roost eaas --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Invalid command, please check 'roost cluster --help' for more info")
		os.Exit(0)
	},
}

var eaasTriggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger an EAAS workflow in roost",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		trigger := func(eaasObj eaas.TriggerEAASObj, AppID string, GitTokenID string) {
			spinner := spinner.NewSpinner()
			spinner.Start("Triggering the EaaS application")
			apiEndPoint := "/api/application/client/git/events/add"
			
			var wfIDReq eaas.GetWorkFlowIDReq
			wfIDReq.AppID = AppID
			wfIDReq.GitTokenID = GitTokenID
			
			eaasObj.WorkflowID = eaas.GetWorkFlowID(wfIDReq)

			reqBuff, err := json.Marshal(eaasObj)
			cobra.CheckErr(err)

			body := bytes.NewReader(reqBuff)
			authkey:="Bearer "+viper.Get("roost_auth_token").(string)
			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
			cobra.CheckErr(err)
			if status != 201{
				spinner.Stop(false)
				fmt.Println("Failed to trigger the application, status code:", status)
			}
			spinner.Stop(true)
			var triggerResp eaas.EAASAPIResp
			err = json.Unmarshal(resp, &triggerResp)
			cobra.CheckErr(err)
			fmt.Println(triggerResp.Msg)
		}

		getapplist := eaas.GetEaasList(false)

		isSetName := cmd.Flags().Lookup("name").Changed
		if isSetName {
			AppName, _ := cmd.Flags().GetString("name")
			for _, AppData := range getapplist.Data {
				if AppData.Appname == AppName {
					eaasObj := eaas.TriggerEAASObj{}

					eaasObj.Branch = AppData.AppRepoBranch
					eaasObj.Type = AppData.CodeRepo
					eaasObj.UserName = AppData.CreatedBy

					repo := strings.Split(AppData.AppRepoName, "/")
					eaasObj.OwnerName = repo[0]
					var repoName string 
					for i := 1; i < len(repo); i++ {
						repoName += repo[i]
					}
					eaasObj.RepoName = repoName
					year, month, day := time.Now().Date()
					hour := time.Now().Local().Hour()
					min := time.Now().Local().Minute()

					eaasObj.Title = "trigger-" + fmt.Sprintf("%d%d%d-%d%d", year, month, day, hour, min) 

					trigger(eaasObj, "zbio", AppData.ID)
				}
			}
		} else {
			if getapplist.Count < 1 {
				fmt.Println("No applications found.")
				return
			}
			var Apps = []string{}
			for _, AppData := range getapplist.Data {
				Apps = append(Apps, AppData.Appname)
			}

			AppNameInput := utils.PromptSelectInput(Apps, "Select the application to trigger.")
			if AppNameInput != "" {
				for _, AppData := range getapplist.Data {
					if AppData.Appname == AppNameInput {
						eaasObj := eaas.TriggerEAASObj{}

						eaasObj.Branch = AppData.AppRepoBranch
						eaasObj.Type = AppData.CodeRepo
						eaasObj.UserName = AppData.CreatedBy
						repo := strings.Split(AppData.AppRepoName, "/")
						eaasObj.OwnerName = repo[0]
						var repoName string 
						for i := 1; i < len(repo); i++ {
							repoName += repo[i]
						}
						eaasObj.RepoName = repoName

						year, month, day := time.Now().Date()
						hour := time.Now().Local().Hour()
						min := time.Now().Local().Minute()
						eaasObj.Title = "trigger-" + fmt.Sprintf("%d%d%d-%d%d", year, month, day, hour, min)
						//fmt.Printf("%v\n", eaasObj)
						trigger(eaasObj, "zbio", AppData.ID)
					}
				}
			}
		}
	},
}

var eaasListEnvCmd = &cobra.Command{
	Use: "list-environments",
	Short: "Display a list of EAAS environments",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		listenvs := func() {
			take, _ := cmd.Flags().GetInt("take")
			getEnvList := eaas.GetEaasListEnv(take)
			if getEnvList.Count > 0 {
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"App-Name", "App-Repo-Name", "App-Repo-Branch", "event type", "Created-by", "Namespace", "Assigned-Cluster", "Event-Status"})
				t.SetStyle(table.StyleDouble)

				for _, EaasData := range getEnvList.Data {
					t.AppendRows([]table.Row{
						{EaasData.AppName, EaasData.RepoName, EaasData.BranchName, EaasData.Action, EaasData.UserName, EaasData.AssignedNS, EaasData.AssignedCluster, EaasData.Status},
					})
				}

				fmt.Print("\n")
				t.Render()
				fmt.Print("\n")
			} else {
				fmt.Println("No environments found, please set up an application in Roost.")
			}
		}
		listenvs()
	},
}

var eaasEnvDetailsCmd = &cobra.Command{
	Use: "get-env-details",
	Short: "get details of an EAAS environment",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		envDetails := func() {
			take, _ := cmd.Flags().GetInt("take")
			getEnvList := eaas.GetEaasListEnv(take)
			if getEnvList.Count > 0 {
				columns := []bubbletable.Column{
					{Title: "App-Name", Width: 20},
					{Title: "Assigned-Cluster", Width: 15},
					{Title: "Repo", Width: 20},
					{Title: "Branch", Width: 20},
					{Title: "Created-By", Width: 30},
					{Title: "Assigned-namespace", Width: 40},
				}
		
				var rows []bubbletable.Row
		
				for _, envData := range getEnvList.Data {
					row := []bubbletable.Row{
						{envData.AppName, fmt.Sprint(envData.AssignedCluster), envData.RepoName, envData.BranchName, envData.UserName, envData.AssignedNS},
					}
					rows = append(rows, row...)
				}
				UserChoice := utils.TableInput(columns, rows)
				if len(UserChoice) != 0 {
					for _, listData := range getEnvList.Data {
						if UserChoice[0] == listData.AppName {
							EnvDataFormatted, err := json.MarshalIndent(listData, "", " ")
							cobra.CheckErr(err)
							fmt.Println(string(EnvDataFormatted))
						}
					}
				} else {
					fmt.Println("Please select an option")
					return
				}

			} else {
				fmt.Println("No environments found, please set up an application in Roost.")
			}
		}
		envDetails()
	},
}

var eaasListAppsCmd = &cobra.Command{
	Use:   "list-apps",
	Short: "List your EAAS applications",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		// common.ValidConfig()
		listApps := func() {
			isSetAll := cmd.Flags().Lookup("all").Changed
			getapplist := eaas.GetEaasList(isSetAll)
			//fmt.Printf("%+v\n", getapplist)
			if getapplist.Count > 0 {
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"ID", "App-Name", "App-Repo-Name", "App-Repo-Branch", "Created-By", "Created-On"})
				t.SetStyle(table.StyleDouble)

				for _, EaasData := range getapplist.Data {
					t.AppendRows([]table.Row{
						{EaasData.ID, EaasData.Appname, EaasData.AppRepoName, EaasData.AppRepoBranch, EaasData.CreatedBy, EaasData.CreatedOn},
					})
				}

				fmt.Print("\n")
				t.Render()
				fmt.Print("\n")
			} else {
				fmt.Println("No Applications found, please set up an application in Roost.")
			}
		}
		listApps()
	},
}

var eaasLogsCmd = &cobra.Command{
	Use:   "get-logs",
	Short: "Get EAAS logs",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		getlogs := func(eaasObj eaas.GetLogsReq) {
			spinner := spinner.NewSpinner()
			spinner.Start("Fetching EAAS logs")
		 	apiEndPoint := "/api/application/client/git/eaas/getLogs"
		 	
			reqBuff, err := json.Marshal(eaasObj)
			cobra.CheckErr(err)
			
			body := bytes.NewReader(reqBuff)
			authkey:="Bearer "+viper.Get("roost_auth_token").(string)
			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
			
			cobra.CheckErr(err)
			if status != 201{
				spinner.Stop(false)
				fmt.Println("Failed to get logs for the selected environment:", status)
			}
			var getLogsObj eaas.GetLogsRes 
		 	err = json.Unmarshal(resp, &getLogsObj)
		 	//fmt.Println(string(resp))
			spinner.Stop(true)
		 	fmt.Printf("%+v\n", getLogsObj)
		}

		take, _ := cmd.Flags().GetInt("take")
		getEnvList := eaas.GetEaasListEnv(take)
		
		if getEnvList.Count < 1 {
			fmt.Println("No environments found")
			return
		}

		columns := []bubbletable.Column{
			{Title: "App-Name", Width: 20},
			{Title: "Assigned-Cluster", Width: 15},
			{Title: "Repo", Width: 20},
			{Title: "Branch", Width: 20},
			{Title: "Created-By", Width: 30},
			{Title: "Assigned-namespace", Width: 40},
		}

		var rows []bubbletable.Row

		for _, envData := range getEnvList.Data {
			row := []bubbletable.Row{
				{envData.AppName, fmt.Sprint(envData.AssignedCluster), envData.RepoName, envData.BranchName, envData.UserName, envData.AssignedNS},
			}
			rows = append(rows, row...)
		}

		var eaasObj eaas.GetLogsReq
		UserChoice := utils.TableInput(columns, rows)
		if len(UserChoice) != 0 {
			for _, listData := range getEnvList.Data {
				if UserChoice[0] == listData.AppName {
					eaasObj.TriggerID = listData.TriggerID
				}
			}
		} else {
			fmt.Println("Please select an option")
			return
		}

		getlogs(eaasObj)
	},
}

var eaasDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an EAAS application",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		eaasObj := eaas.DeleteAppObj{}
		deleteApp := func(ID string) {
			eaasObj.AppID = "zbio"
			eaasObj.GitTokenID = ID
			eaasObj.DeleteAssociatedWorkFlows = true 
			spinner := spinner.NewSpinner()
			spinner.Start("Deleting the requested EAAS application")
			apiEndPoint := "/api/application/client/git/token/delete"
			reqBuff, err := json.Marshal(eaasObj)
			if err != nil {
				spinner.Stop(false)
				fmt.Println(err)
				os.Exit(0)
			}
			body := bytes.NewReader(reqBuff)
			authkey:="Bearer "+viper.Get("roost_auth_token").(string)
			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
			cobra.CheckErr(err)
			if status != 201{
				spinner.Stop(false)
				fmt.Println("Failed to delete the application, status code:", status)
			}

			var RespMsg eaas.EAASAPIResp
			err = json.Unmarshal(resp, &RespMsg)
			cobra.CheckErr(err)
			spinner.Stop(true)
			fmt.Println(RespMsg.Msg)
		}
		getapplist := eaas.GetEaasList(false)

		isSetName := cmd.Flags().Lookup("name").Changed
		if isSetName {
			AppName, _ := cmd.Flags().GetString("name")
			for _, AppData := range getapplist.Data {
				if AppData.Appname == AppName {
					deleteApp(AppData.ID)
				}
			}
		} else {
			if getapplist.Count < 1 {
				fmt.Println("No applications found.")
				return
			}
			var Apps = []string{}
			for _, AppData := range getapplist.Data {
				Apps = append(Apps, AppData.Appname)
			}

			AppNameInput := utils.PromptSelectInput(Apps, "Select the application you want to delete.")
			if AppNameInput != "" {
				for _, AppData := range getapplist.Data {
					if AppData.Appname == AppNameInput {
						deleteApp(AppData.ID)
					}
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(eaasCmd)
	eaasCmd.AddCommand(eaasTriggerCmd)
	eaasCmd.AddCommand(eaasListAppsCmd)
	eaasCmd.AddCommand(eaasLogsCmd)
	eaasCmd.AddCommand(eaasDeleteCmd)
	eaasCmd.AddCommand(eaasListEnvCmd)
	eaasCmd.AddCommand(eaasEnvDetailsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eaasCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eaasCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	eaasTriggerCmd.Flags().StringP("name", "n", "", "Trigger an EAAS workflow by application name.")

	eaasLogsCmd.Flags().Int("take", 15, "Set how many environments will be fetched.")

	eaasListAppsCmd.Flags().Int("take", 10, "Set how many environments will be fetched.")

	eaasListAppsCmd.Flags().BoolP("all", "a", false, "List All EaaS Applications.")

	eaasEnvDetailsCmd.Flags().Int("take", 10, "Set how many environments will be fetched.")

	eaasDeleteCmd.Flags().StringP("name", "n", "", "Delete an application by it's name.")
}