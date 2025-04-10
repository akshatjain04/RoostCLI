/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ZB-io/internal/roostcli/pkg/cluster"
	"github.com/ZB-io/internal/roostcli/pkg/config"
	"github.com/ZB-io/internal/roostcli/pkg/spinner"
	team "github.com/ZB-io/internal/roostcli/pkg/team"
	"github.com/ZB-io/internal/roostcli/pkg/utils"
	bubbletable "github.com/charmbracelet/bubbles/table"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// teamCmd represents the team command
var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "A command to interact with Roost Teams",
	Long:  ``,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(config.LoadServerFromViper())
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var teamCreate = &cobra.Command{
	Use:   "create",
	Short: "A command to create team in roost",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.LoadServerFromViper()
		cobra.CheckErr(err)
		var createteamdetails team.CreateTeam

		createteamdetails.Description, _ = cmd.Flags().GetString("Description")
		createteamdetails.FirstMembers, _ = cmd.Flags().GetStringSlice("members")
		createteamdetails.Name, _ = cmd.Flags().GetString("name")
		createteamdetails.Org, _ = cmd.Flags().GetString("org")
		createteamdetails.Visibility, _ = cmd.Flags().GetString("visibility")

		err = utils.AcceptFromPrompt(&createteamdetails)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("create cluster prompt error %q", err.Error()))
		}

		spinner := spinner.NewSpinner()
		spinner.Start("Creating the team")
		apiEndPoint := "/api/team/create"
		reqBuff, err := json.Marshal(createteamdetails)
		cobra.CheckErr(err)
		body := bytes.NewReader(reqBuff)
		authkey := "Bearer " + viper.Get("roost_auth_token").(string)
		status, respbody, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}
		if status == http.StatusCreated {
			spinner.Stop(true)
			fmt.Println("Succesfully created the team with name", createteamdetails.Name)
		} else {
			spinner.Stop(false)
			fmt.Println("Unable to Create Team", string(respbody))
		}
	},
}

var getTeamDetails = &cobra.Command{
	Use:   "list",
	Short: "A command to list all the teams you are part of",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.LoadServerFromViper()
		cobra.CheckErr(err)
		Teaminfo := teamDetails()

		if Teaminfo.Count > 0 {

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"TEAM NAME", "DESCRIPTION", "VISIBILITY", "ROLE", "TOTAL MEMBERS", "JOINED ON", "TEAMID"})
			t.SetStyle(table.StyleDouble)
			for _, teamData := range Teaminfo.Teamlist {
				t.AppendRows([]table.Row{
					{teamData.Name, teamData.Description, teamData.Visibility, teamData.MemberRole, teamData.MemberCount, teamData.JoiningDate, teamData.TeamId, teamData.Isadmin},
				})
				// t.AppendSeparator()
			}

			fmt.Print("\n")
			t.Render()
			fmt.Print("\n")
		}

	},
}

var teamDelete = &cobra.Command{
	Use:   "delete",
	Short: "A command to delete team",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		var deleteinfo team.DeleteTeam
		Teaminfo := teamDetails()
		teamname, _ := cmd.Flags().GetString("name")
		for _, teamData := range Teaminfo.Teamlist {
			if teamData.Name == teamname {
				deleteinfo.TeamID = teamData.TeamId
			}
		}
		isSet := cmd.Flags().Lookup("name").Changed

		if !isSet {
			columns := []bubbletable.Column{
				{Title: "No.", Width: 4},
				{Title: "Name", Width: 15},
				{Title: "Role", Width: 15},
				{Title: "Team-ID", Width: 40},
			}

			var rows []bubbletable.Row
			count := 0
			for _, teamData := range Teaminfo.Teamlist {
				if teamData.Isadmin == 1 {
					count++
					row := []bubbletable.Row{
						{fmt.Sprint(count), teamData.Name, teamData.MemberRole, teamData.TeamId},
					}
					rows = append(rows, row...)
				}
			}
			UserChoice := utils.TableInput(columns, rows)
			if len(UserChoice) != 0 {
				deleteinfo.TeamID = UserChoice[3]
			}
		}

		if deleteinfo.TeamID == "" {
			fmt.Println("No TeamID was selected or given from the flag", teamname)
			return
		}

		spinner := spinner.NewSpinner()
		spinner.Start("Deleting the team")
		apiEndPoint := "/api/team/delete"
		reqBuff, err := json.Marshal(deleteinfo)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		body := bytes.NewReader(reqBuff)
		authkey := "Bearer " + viper.Get("roost_auth_token").(string)
		status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}

		var apiresp team.TeamApiResponse

		json.Unmarshal(resp, &apiresp)

		if status == http.StatusCreated {
			spinner.Stop(true)
			fmt.Println("Succesfully deleted the team:", deleteinfo.TeamID)
		} else {
			spinner.Stop(false)
			fmt.Println("Unable to delete team,please check bearer token or admin settings", apiresp.TeamRespMessage)
		}

	},
}

var teamInviteMember = &cobra.Command{
	Use:   "invite-member",
	Short: "A command to invite a member in team",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.LoadServerFromViper()
		cobra.CheckErr(err)
		var invitedetails team.InviteMembers

		invitedetails.TeamID, _ = cmd.Flags().GetString("id")
		invitedetails.Username, _ = cmd.Flags().GetStringSlice("members")
		fmt.Println(invitedetails)
		spinner := spinner.NewSpinner()
		spinner.Start("Sending team invites")
		apiEndPoint := "/api/team/inviteMultiple"
		reqBuff, err := json.Marshal(invitedetails)
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		body := bytes.NewReader(reqBuff)
		authkey := "Bearer " + viper.Get("roost_auth_token").(string)
		status, respbody, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}
		if status == http.StatusCreated {
			spinner.Stop(true)
			fmt.Println("Succesfully sent invite to the following members", invitedetails.Username)
		} else {
			spinner.Stop(false)
			fmt.Println("Unable to add members", string(respbody))
		}
	},
}

var teamRemoveMember = &cobra.Command{
	Use:   "remove-member",
	Short: "A command to remove a member from team",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.LoadServerFromViper()
		cobra.CheckErr(err)
		var removedetails team.RemoveMember
		removedetails.TeamID, _ = cmd.Flags().GetString("id")
		removeMember := func(memberid string) {
			removedetails.MemberID = memberid
			spinner := spinner.NewSpinner()
			spinner.Start("Removing team member")
			apiEndPoint := "/api/team/removeMember"
			reqBuff, err := json.Marshal(removedetails)
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
			body := bytes.NewReader(reqBuff)
			authkey := "Bearer " + viper.Get("roost_auth_token").(string)
			status, respbody, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
			if err != nil {
				spinner.Stop(false)
				fmt.Println(err)
				os.Exit(0)
			}
			if status == http.StatusCreated {
				spinner.Stop(true)
				fmt.Println("Succesfully removed the member with ID", removedetails.MemberID)
			} else {
				spinner.Stop(false)
				fmt.Println("Unable to add members", string(respbody))
			}
		}
		MemberIDs, _ := cmd.Flags().GetStringSlice("members")
		for _, member := range MemberIDs {
			removeMember(member)
		}

	},
}

var getTeamConfig = &cobra.Command{
	Use:   "get-kubeconfig",
	Short: "A command to get KUBECONFIG of the selected team cluster",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {

		var kubeconfigteam team.TeamKubeConfigObj

		columns := []bubbletable.Column{
			{Title: "No.", Width: 4},
			{Title: "Name", Width: 15},
			{Title: "Role", Width: 15},
			{Title: "Team-ID", Width: 40},
		}

		Teaminfo := teamDetails()
		var rows []bubbletable.Row
		count := 0
		for _, teamData := range Teaminfo.Teamlist {
			if teamData.Isadmin == 1 {
				count++
				row := []bubbletable.Row{
					{fmt.Sprint(count), teamData.Name, teamData.MemberRole, teamData.TeamId},
				}
				rows = append(rows, row...)
			}
		}
		UserChoice := utils.TableInput(columns, rows)
		if len(UserChoice) != 0 {
			kubeconfigteam.TeamId = UserChoice[3]
			kubeconfigteam.Clustertype = []string{"roost", "managed"}
		} else {
			fmt.Println("Please select an option")
			return
		}

		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		kubeConfigDir := filepath.Join(home, ".kube", "roostteamconfig")
		var getKubeConfig []team.TeamKubeConfigResponse
		spinner := spinner.NewSpinner()
		spinner.Start("Getting the kubeconfig of the attached team cluster")
		apiEndPoint := "/api/application/getTeamCluster"
		reqBuff, err := json.Marshal(kubeconfigteam)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}
		body := bytes.NewReader(reqBuff)
		authkey := "Bearer " + viper.Get("roost_auth_token").(string)
		status, respbody, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
		cobra.CheckErr(err)
		err = json.Unmarshal(respbody, &getKubeConfig)
		cobra.CheckErr(err)
		if !utils.FileOrFolderExists(kubeConfigDir) {
			spinner.Stop(false)
			err := os.MkdirAll(kubeConfigDir, 0755)
			cobra.CheckErr(err)
		}
		kubeConfigPath := filepath.Join(kubeConfigDir, UserChoice[1])

		err = os.WriteFile(kubeConfigPath, []byte(getKubeConfig[0].Kubeconfig), 0644)
		cobra.CheckErr(err)
		var apiresp cluster.ClusterApiResponse
		if status == http.StatusCreated {
			spinner.Stop(true)
			fmt.Printf("The kubeconfig file is present in $HOME/.kube/roostteamconfig/%s.\nUse 'export KUBECONFIG=$HOME/.kube/roostconfig/%s'.\n", UserChoice[1], UserChoice[1])
		} else {
			spinner.Stop(false)
			json.Unmarshal(respbody, &apiresp)
			fmt.Println("Unable to get the kubeconfig of the requested cluster: ", string(apiresp.ClusterRespMessage))
		}

	},
}

var addTeamCluster = &cobra.Command{
	Use:   "add-cluster",
	Short: "A command to add cluster in roost teams",
	Long:  "Use 'roost cluster stop --help' for more info",
	Run: func(cmd *cobra.Command, args []string) {
		Teaminfo := teamDetails()
		var teamclusteradd team.ClusterAdd
		teamclusteradd.AwsCredentials.CredentialInputType = "input"

		clusterListData := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
		if clusterListData.Count < 1 || len(clusterListData.Clusters) < 1 {
			fmt.Println("No clusters are found")
			return
		}
		var ActiveClusters = []string{}
		for _, clusterData := range clusterListData.Clusters {
			if clusterData.IsActive { //&& clusterData.EnvType=="K8s"{
				ActiveClusters = append(ActiveClusters, clusterData.CustomerToken)
			}
		}
		if len(ActiveClusters) < 1 {
			fmt.Println("No running clusters are found")
			return
		}

		clusterInput := utils.PromptSelectInput(ActiveClusters, "Select the cluster you want to add to team")
		if clusterInput == "" {
			return
		}

		for _, clusterData := range clusterListData.Clusters {
			if clusterData.CustomerToken == clusterInput {
				teamclusteradd.ClusterId = clusterData.Id
				teamclusteradd.CustomerToken = clusterData.CustomerToken
				teamclusteradd.CustomerEmail = clusterData.CustomerEmail
			}
		}

		columns := []bubbletable.Column{
			{Title: "No.", Width: 4},
			{Title: "Name", Width: 15},
			{Title: "Role", Width: 15},
			{Title: "Team-ID", Width: 40},
		}

		var rows []bubbletable.Row
		count := 0
		for _, teamData := range Teaminfo.Teamlist {
			if teamData.Isadmin == 1 {
				count++
				row := []bubbletable.Row{
					{fmt.Sprint(count), teamData.Name, teamData.MemberRole, teamData.TeamId},
				}
				rows = append(rows, row...)
			}
		}
		UserChoice := utils.TableInput(columns, rows)
		if len(UserChoice) != 0 {
			teamclusteradd.TeamId = UserChoice[3]
			teamclusteradd.RbacScope = "namespace"
		} else {
			fmt.Println("Please select an option")
		}

		spinner := spinner.NewSpinner()
		spinner.Start("Attaching selected cluster to team")
		apiEndPoint := "/api/application/register/team"
		reqBuff, err := json.Marshal(teamclusteradd)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}
		body := bytes.NewReader(reqBuff)
		authkey := "Bearer " + viper.Get("roost_auth_token").(string)
		status, respbody, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}
		if status != http.StatusCreated {
			spinner.Stop(false)
			fmt.Println("Error in team update", string(respbody))
			return
		}

		spinner.Stop(true)
		var clusterInfo team.UpdateClusterInfo

		clusterInfo.Teamconfig.TeamClusterId = teamclusteradd.ClusterId
		clusterInfo.TeamId = teamclusteradd.TeamId
		clusterInfo.Teamconfig.Restrictuseraccess = true

		utils.AcceptFromPrompt(&clusterInfo.Teamconfig)
	
		spinner.Start("Updating team details")
		apiEndPoint = "/api/team/update"
		reqBuff, err = json.Marshal(clusterInfo)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}

		body = bytes.NewReader(reqBuff)
		status, respbody, err = utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, body)
		if err != nil {
			spinner.Stop(false)
			fmt.Println(err)
			os.Exit(0)
		}

		if status == http.StatusCreated {
			spinner.Stop(true)
			fmt.Printf("Succesfully added the cluster to team %v with ID %v", teamclusteradd.TeamId, teamclusteradd.ClusterId)
		} else {
			spinner.Stop(false)
			fmt.Println("Unable to add cluster", string(respbody))
		}

	},
}

func teamDetails() team.TeamListResponse {
	spinner := spinner.NewSpinner()
	spinner.Start("Fetching teams")
	apiEndPoint := "/api/team/getMyTeams"
	var getTeamList team.TeamListResponse
	authkey := "Bearer " + viper.Get("roost_auth_token").(string)
	status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, authkey, &bytes.Reader{})
	if err != nil {
		spinner.Stop(false)
		fmt.Println(err)
		os.Exit(0)
	}

	if status != http.StatusCreated {
		spinner.Stop(false)
		fmt.Println("Unable to fetch team list please check the Bearer token", err)
		return team.TeamListResponse{}
	}
	err = json.Unmarshal(resp, &getTeamList)
	cobra.CheckErr(err)

	spinner.Stop(true)
	return getTeamList
}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamCreate)
	teamCmd.AddCommand(getTeamDetails)
	teamCmd.AddCommand(teamDelete)
	teamCmd.AddCommand(teamInviteMember)
	teamCmd.AddCommand(teamRemoveMember)
	teamCmd.AddCommand(addTeamCluster)
	teamCmd.AddCommand(getTeamConfig)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// teamCmd.PersistentFlags().String("foo", "", "A help for foo")
	teamCreate.Flags().String("description", "", "Describe your team")
	teamCreate.Flags().String("name", "", "Name your team")
	teamCreate.Flags().String("visibility", "private", "To Specify Visibility")
	teamCreate.Flags().String("org", "", "To Specify Organisation")
	teamCreate.Flags().StringSlice("members", []string{}, "Specify Members in team")

	teamDelete.Flags().String("name", "", "To specify team name to be deleted")
	//teamDelete.MarkFlagRequired("id")

	teamInviteMember.Flags().String("id", "", "To Specify of Team to be invited")
	teamInviteMember.Flags().StringSlice("User", []string{}, "Specify Members in team to be invited")
	teamInviteMember.MarkFlagRequired("id")
	teamInviteMember.MarkFlagRequired("members")

	teamRemoveMember.Flags().String("id", "", "To Specify of Team to be invited")
	teamRemoveMember.Flags().StringSlice("members", []string{}, "Specify Members in team to be invited")
	teamRemoveMember.MarkFlagRequired("id")
	teamRemoveMember.MarkFlagRequired("members")

	//teamCreate.MarkFlagsMutuallyExclusive()
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// teamCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
