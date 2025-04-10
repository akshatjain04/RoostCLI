/*
Package cmd Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ZB-io/internal/roostcli/pkg/cluster"
	"github.com/ZB-io/internal/roostcli/pkg/config"
	"github.com/ZB-io/internal/roostcli/pkg/spinner"
	"github.com/ZB-io/internal/roostcli/pkg/utils"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//ent server

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "A command to interact with Roost clusters",
	Long:  ``,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(config.LoadServerFromViper())
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		if args[0] != "help" {
			fmt.Printf("%v is not a valid command\n", args[0])
		}
		cmd.Help()
	},
	Example: `
	roost cluster create
	roost cluster list
	roost cluster ui
	roost cluster get-details
	roost cluster get-kubeconfig
	roost cluster stop
	roost cluster delete
	
	`,
}

var clusterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Launch a Roost Cluster.",
	Long:  `A command to start a Roost cluster, it prompts the user for the cluster specifications, if not provided then default values of the specifications are used.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		clusterObj := cluster.CreateClusterRequest{}
		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}
		isSet := cmd.Flags().Lookup("alias").Changed
		if !isSet {
			clusterObj.Alias = fmt.Sprintf("roostcli-%d", time.Now().Unix())
		} else {
			clusterObj.Alias, _ = cmd.Flags().GetString("alias")
		}
		clusterObj.Namespace, _ = cmd.Flags().GetString("namespace")
		clusterObj.Ami, _ = cmd.Flags().GetString("ami")
		clusterObj.InstanceType, _ = cmd.Flags().GetString("instance-type")
		clusterObj.DiskSize, _ = cmd.Flags().GetString("disk-size")
		clusterObj.Region, _ = cmd.Flags().GetString("region")
		clusterObj.ClusterExpiry, _ = cmd.Flags().GetInt("expiry")
		clusterObj.K8sVersion, _ = cmd.Flags().GetString("k8s")
		clusterObj.WorkerNodes, _ = cmd.Flags().GetInt("nodes")
		clusterObj.Email, _ = cmd.Flags().GetString("email")

		err = utils.AcceptFromPrompt(&clusterObj)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("create cluster prompt error %q", err.Error()))
		}

		clusterObj.RoostAuthToken = viper.Get("roost_auth_token").(string)
		spinner := spinner.NewSpinner()
		spinner.Start("Creating a cluster")
		apiEndPoint := "/api/application/client/launchCluster"
		reqBuff, err := json.Marshal(clusterObj)
		cobra.CheckErr(err)

		var clusterResp cluster.ClusterApiResponse
		body := bytes.NewReader(reqBuff)
		status, response, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
		cobra.CheckErr(err)
		if status != http.StatusCreated {
			json.Unmarshal(response, &clusterResp)
			spinner.Stop(false)
			fmt.Println("Unable to create cluster: ", clusterResp.ClusterRespMessage)
		} else {
			spinner.Stop(true)
			fmt.Println("cluster creation in progress, It may take 5 min to comeup.\nRequested Cluster alias: ", clusterObj.Alias)
		}
	},
	Example: `
	roost cluster create
	roost cluster create --email test@mail.com
	roost cluster create --alias example.Alias
	`,
}

var clusterStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a Roost cluster",
	Long:  "A command to stop a roost cluster, provides a list of all the currently running/requested clusters. The cluster which needs to be stopped can then be selected from the list or its ID or alias can be provided as flags.",
	Run: func(cmd *cobra.Command, args []string) {
		clusterObj := cluster.ClusterStopObj{}

		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}

		clusterStop := func(clusterAlias string) {
			clusterObj.Alias = clusterAlias
			clusterObj.RoostAuthToken = viper.Get("roost_auth_token").(string)
			spinner := spinner.NewSpinner()
			spinner.Start("stopping the requested cluster")
			apiEndPoint := "/api/application/client/stopLaunchedCluster"
			reqBuff, err := json.Marshal(clusterObj)
			cobra.CheckErr(err)
			body := bytes.NewReader(reqBuff)

			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
			cobra.CheckErr(err)

			var clusterresp cluster.ClusterApiResponse
			json.Unmarshal(resp, &clusterresp)

			if status == http.StatusCreated {
				spinner.Stop(true)
				fmt.Println("Succesfully stopped the cluster with alias", clusterObj.Alias)
			} else {
				spinner.Stop(false)
				fmt.Println("Unable to stop cluster:", string(clusterresp.ClusterRespMessage))
			}
		}

		isSetID := cmd.Flags().Lookup("id").Changed
		if isSetID {
			ClusterIDs, _ := cmd.Flags().GetInt32Slice("id")
			for _, ClusterID := range ClusterIDs {
				clusterinfo, err := cluster.GetClusterDetails(int(ClusterID), "")
				cobra.CheckErr(err)
				clusterStop(clusterinfo.CustomerToken)
			}
		}

		isSetAlias := cmd.Flags().Lookup("alias").Changed
		if isSetAlias {
			ClusterAliasArr, _ := cmd.Flags().GetStringSlice("alias")
			for _, ClusterAlias := range ClusterAliasArr {
				clusterStop(ClusterAlias)
			}
		}

		if !isSetAlias && !isSetID {
			listResponse := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
			if listResponse.Count < 1 || len(listResponse.Clusters) < 1 {
				fmt.Println("No running clusters are found")
				return
			}

			var clusterNames []string
			for _, clusterData := range listResponse.Clusters {
				if clusterData.StatusMsg == "Request in Progress ..." || clusterData.IsActive == true {
					clusterNames = append(clusterNames, clusterData.CustomerToken)
				}
			}
			if len(clusterNames) < 1 {
				fmt.Println("No running clusters are found")
				return
			}

			clusterAliasInput := utils.PromptSelectInput(clusterNames, "Select the cluster you want to stop")
			if clusterAliasInput == "" {
				return
			}
			clusterStop(clusterAliasInput)
		}
	},
	Example: `
		roost cluster stop
		roost cluster stop --id 1
		roost cluster stop --id 1,2,3
		roost cluster stop --alias ExampleAlias
		roost cluster stop --alias ExampleAlias1. ExampleAlias2
	`,
}

var clusterDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Roost cluster",
	Long:  `A command to delete a roost cluster, provides a list of currently available clusters. The cluster to be deleted can then be selected from the given list or its ID or alias can be provided as flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		clusterObj := cluster.ClusterStopObj{}
		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}

		clusterDelete := func(clusterAlias string) {
			clusterObj.Alias = clusterAlias
			clusterObj.RoostAuthToken = viper.Get("roost_auth_token").(string)
			spinner := spinner.NewSpinner()
			spinner.Start("Deleting the requested cluster")
			apiEndPoint := "/api/application/client/deleteLaunchedCluster"
			reqBuff, err := json.Marshal(clusterObj)
			cobra.CheckErr(err)

			body := bytes.NewReader(reqBuff)
			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
			cobra.CheckErr(err)
			var apiresp cluster.ClusterApiResponse
			json.Unmarshal(resp, &apiresp)

			home, err := os.UserHomeDir()
			kubeConfigDir := filepath.Join(home, ".kube", "roostconfig")
			kubeConfigPath := filepath.Join(kubeConfigDir, clusterAlias)
			if utils.FileOrFolderExists(kubeConfigPath) {
				err = os.Remove(kubeConfigPath)
				if err != nil {
					spinner.Stop(false)
					fmt.Println("error deleting the downloaded kubeconfig of the file")
				}
			}
			if status == http.StatusCreated {
				spinner.Stop(true)
				fmt.Println("Succesfully deleted the cluster with Alias", clusterAlias)
			} else {
				spinner.Stop(false)
				fmt.Println("Unable to delete cluster:", string(apiresp.ClusterRespMessage))
			}
		}

		isSetID := cmd.Flags().Lookup("id").Changed
		if isSetID {
			ClusterIDs, _ := cmd.Flags().GetInt32Slice("id")
			for _, ClusterID := range ClusterIDs {
				clusterinfo, err := cluster.GetClusterDetails(int(ClusterID), "")
				cobra.CheckErr(err)
				clusterDelete(clusterinfo.CustomerToken)
			}
		}

		isSetAlias := cmd.Flags().Lookup("alias").Changed
		if isSetAlias {
			ClusterAliasArr, _ := cmd.Flags().GetStringSlice("alias")
			for _, ClusterAlias := range ClusterAliasArr {
				clusterDelete(ClusterAlias)
			}
		}

		if !isSetAlias && !isSetID {
			clusterListData := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
			if clusterListData.Count < 1 || len(clusterListData.Clusters) < 1 {
				fmt.Println("No clusters are found")
				return
			}
			var custToken = []string{}
			for _, clusterData := range clusterListData.Clusters {
				custToken = append(custToken, clusterData.CustomerToken)
			}

			clusterAliasInput := utils.PromptSelectInput(custToken, "Select the cluster you want to get delete")
			if clusterAliasInput == "" {
				return
			}
			clusterDelete(clusterAliasInput)
		}
	},
	Example: `
	roost cluster delete
	roost cluster delete --id 1
	roost cluster delete --id 1,2,3
	roost cluster delete --alias ExampleAlias
	roost cluster delete --alias ExampleAlias1. ExampleAlias2
	`,
}

var clusterKubeconfigCmd = &cobra.Command{
	Use:   "get-kubeconfig",
	Short: "Get KUBECONFIG of the roost provisioned cluster",
	Long:  `A command to get the kubeconfig of a roost provisioned cluster, provides a list of all the running clusters, the cluster for which the kubeconfig is to be downloaded can then be selected from the provided list or its ID or alias can be provided as flags. The kubeconfig file, once fetched will then be stored in '$HOME/.kube/config'`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}
		var kubeConfigPath string
		clusterObj := cluster.ClusterKubeconfig{}
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		kubeConfigDir := filepath.Join(home, ".kube", "roostconfig")
		var getKubeConfigObj cluster.ClusterKubeconfigResponse
		clusterGetKubeConfig := func(clusterAlias string) {
			kubeConfigPath = filepath.Join(kubeConfigDir, clusterAlias)
			clusterObj.Alias = clusterAlias
			clusterObj.RoostAuthToken = viper.Get("roost_auth_token").(string)
			spinner := spinner.NewSpinner()
			spinner.Start("Getting the kubeconfig of the requested cluster")
			apiEndPoint := "/api/application/cluster/getKubeConfig"
			reqBuff, err := json.Marshal(clusterObj)
			cobra.CheckErr(err)
			body := bytes.NewReader(reqBuff)
			status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
			cobra.CheckErr(err)
			err = json.Unmarshal(resp, &getKubeConfigObj)
			cobra.CheckErr(err)
			if !utils.FileOrFolderExists(kubeConfigDir) {
				err := os.MkdirAll(kubeConfigDir, 0755)
				cobra.CheckErr(err)
			}

			err = os.WriteFile(kubeConfigPath, []byte(getKubeConfigObj.Kubeconfig), 0644)
			cobra.CheckErr(err)
			var apiresp cluster.ClusterApiResponse
			if status == http.StatusCreated {
				spinner.Stop(true)
				fmt.Printf("The kubeconfig file is present in $HOME/.kube/roostconfig/%s.\nUse 'export KUBECONFIG=$HOME/.kube/roostconfig/%s'.\n", clusterObj.Alias, clusterObj.Alias)
			} else {
				spinner.Stop(false)
				json.Unmarshal(resp, &apiresp)
				fmt.Println("Unable to get the kubeconfig of the requested cluster: ", string(apiresp.ClusterRespMessage))
			}
		}

		isSetID := cmd.Flags().Lookup("id").Changed
		if isSetID {
			ClusterIDs, _ := cmd.Flags().GetInt32Slice("id")
			for _, ClusterID := range ClusterIDs {
				clusterinfo, err := cluster.GetClusterDetails(int(ClusterID), "")
				cobra.CheckErr(err)
				clusterGetKubeConfig(clusterinfo.CustomerToken)
			}
		}

		isSetAlias := cmd.Flags().Lookup("alias").Changed
		if isSetAlias {
			ClusterAliasArr, _ := cmd.Flags().GetStringSlice("alias")
			for _, ClusterAlias := range ClusterAliasArr {
				clusterGetKubeConfig(ClusterAlias)
			}
		}

		if !isSetAlias && !isSetID {
			clusterListData := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
			if clusterListData.Count < 1 || len(clusterListData.Clusters) < 1 {
				fmt.Println("No clusters are found")
				return
			}
			var custToken = []string{}
			for _, clusterData := range clusterListData.Clusters {
				if clusterData.IsActive == true {
					custToken = append(custToken, clusterData.CustomerToken)
				}
			}
			if len(custToken) < 1 {
				fmt.Println("No running clusters are found")
				return
			}

			clusterAliasInput := utils.PromptSelectInput(custToken, "Select the cluster you want to get kubeconfig of")

			if clusterAliasInput == "" {
				return
			}
			clusterGetKubeConfig(clusterAliasInput)
		}
	},
	Example: `
	roost cluster get-kubeconfig
	roost cluster get-kubeconfig --id 1
	roost cluster get-kubeconfig --id 1,2,3
	roost cluster get-kubeconfig --alias ExampleAlias
	roost cluster get-kubeconfig --alias ExampleAlias1. ExampleAlias2
	`,
}

var clusterListCmd = &cobra.Command{
	Use:   "list",
	Short: "A command to get the list of Roost cluster",
	Long: `A command to list all the available roost clusters
	Use 'roost cluster list --help' for more info`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}
		clusterListData := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
		if clusterListData.Count > 0 {

			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"ID", "Cluster Alias", "Email", "Public IP", "Nodes", "Cluster Start Time", "Status"})
			t.SetStyle(table.StyleDouble)
			flagRunning, _ := cmd.Flags().GetBool("running")
			flagStopped, _ := cmd.Flags().GetBool("stopped")
			if flagRunning == true {
				for _, clusterData := range clusterListData.Clusters {
					if clusterData.IsActive == true {
						t.AppendRows([]table.Row{
							{clusterData.Id, clusterData.CustomerToken, clusterData.CustomerEmail, clusterData.PublicIP, clusterData.NumNodes, clusterData.CreatedOn, clusterData.StatusMsg},
						})
					}
				}
			} else if flagStopped == true {
				for _, clusterData := range clusterListData.Clusters {
					if clusterData.StatusMsg == "Stopped ..." {
						t.AppendRows([]table.Row{
							{clusterData.Id, clusterData.CustomerToken, clusterData.CustomerEmail, clusterData.PublicIP, clusterData.NumNodes, clusterData.CreatedOn, clusterData.StatusMsg},
						})
					}
				}
			} else {
				for _, clusterData := range clusterListData.Clusters {
					t.AppendRows([]table.Row{
						{clusterData.Id, clusterData.CustomerToken, clusterData.CustomerEmail, clusterData.PublicIP, clusterData.NumNodes, clusterData.CreatedOn, clusterData.StatusMsg},
					})
				}
			}

			fmt.Print("\n")
			t.Render()
			fmt.Print("\n")
		} else {
			fmt.Println("No clusters found. Use 'roost cluster create' command to create a new roost cluster.")
		}
	},
	Example: `
	roost cluster list
	roost cluster list --running
	roost cluster list --stopped
	`,
}

var clusterDetailsCmd = &cobra.Command{
	Use:   "get-details",
	Short: "Get all details of specific roost cluster",
	Long:  `A command to list of the details of a specific cluster selected by the user`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}

		var ClusterDataFormatted []byte
		var err error
		isSetOut := cmd.Flags().Lookup("output").Changed
		isSetID := cmd.Flags().Lookup("id").Changed
		if isSetID {
			ClusterID, _ := cmd.Flags().GetInt32("id")
			clusterInfo, err := cluster.GetClusterDetails(int(ClusterID), "")
			cobra.CheckErr(err)
			ClusterDataFormatted, err = json.MarshalIndent(clusterInfo, "", " ")
			cobra.CheckErr(err)
			if isSetOut {
				path, _ := cmd.Flags().GetString("output")
				jsonPath := filepath.Join(path, fmt.Sprintf("%d", ClusterID))
				if !utils.FileOrFolderExists(path) {
					err := os.MkdirAll(path, 0755)
					cobra.CheckErr(err)
				}
	
				err = os.WriteFile(jsonPath, ClusterDataFormatted, 0644)
				
				fmt.Println("The details JSON file is present in: ", jsonPath)
			}
			fmt.Println(string(ClusterDataFormatted))
		}

		isSetAlias := cmd.Flags().Lookup("alias").Changed
		if isSetAlias {
			clusterAlias, _ := cmd.Flags().GetString("alias")
			clusterInfo, err := cluster.GetClusterDetails(-1, clusterAlias)
			cobra.CheckErr(err)
			ClusterDataFormatted, err = json.MarshalIndent(clusterInfo, "", " ")
			cobra.CheckErr(err)
			if isSetOut {
				path, _ := cmd.Flags().GetString("output")
				jsonPath := filepath.Join(path, clusterAlias)
				if !utils.FileOrFolderExists(path) {
					err := os.MkdirAll(path, 0755)
					cobra.CheckErr(err)
				}
	
				err = os.WriteFile(jsonPath, ClusterDataFormatted, 0644)
				
				fmt.Println("The details JSON file is present in: ", jsonPath)
			}
			fmt.Println(string(ClusterDataFormatted))
		}

		if !isSetAlias && !isSetID {
			clusterListData := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
			if clusterListData.Count < 1 || len(clusterListData.Clusters) < 1 {
				fmt.Println("No clusters are found")
				return
			}
			var custToken = []string{}
			for _, clusterData := range clusterListData.Clusters {
				custToken = append(custToken, clusterData.CustomerToken)
			}

			clusterAliasInput := utils.PromptSelectInput(custToken, "Select the cluster you want to get details")
			if clusterAliasInput != "" {
				for _, clusterData := range clusterListData.Clusters {
					if clusterData.CustomerToken == clusterAliasInput {
						ClusterDataFormatted, err = json.MarshalIndent(clusterData, "", " ")
						cobra.CheckErr(err)
						if isSetOut {
							path, _ := cmd.Flags().GetString("output")
							jsonPath := filepath.Join(path, clusterAliasInput)
							if !utils.FileOrFolderExists(path) {
								err := os.MkdirAll(path, 0755)
								cobra.CheckErr(err)
							}
				
							err = os.WriteFile(jsonPath, ClusterDataFormatted, 0644)
							
							fmt.Println("The details JSON file is present in: ", jsonPath)
						}
						fmt.Println(string(ClusterDataFormatted))
					}
				}
			}
		}
	},
	Example: `
	roost cluster get-details
	roost cluster get-details --id 1
	roost cluster get-details --alias exampleAlias
	`,
}

var clusterUICmd = &cobra.Command{
	Use:   "ui",
	Short: "Connect to roost service fitness for a specific cluster",
	Long:  `A command to display service fitness UI of roost.ai for a specific cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			if args[0] != "help" {
				fmt.Printf("%v is not a valid argument to the command %v\n", args[0], cmd.Name())
			}
			cmd.Help()
			return
		}
		isSetID := cmd.Flags().Lookup("id").Changed
		var err error
		if isSetID {
			ClusterID, _ := cmd.Flags().GetInt32("id")
			clusterInfo, err := cluster.GetClusterDetails(int(ClusterID), "")
			cobra.CheckErr(err)
			clusterIP := clusterInfo.PublicIP
			err = utils.Openbrowser("http://" + clusterIP + ":30070/app")
		}

		isSetAlias := cmd.Flags().Lookup("alias").Changed
		if isSetAlias {
			clusterAlias, _ := cmd.Flags().GetString("alias")
			clusterInfo, err := cluster.GetClusterDetails(-1, clusterAlias)
			cobra.CheckErr(err)
			clusterIP := clusterInfo.PublicIP
			err = utils.Openbrowser("http://" + clusterIP + ":30070/app")
		}

		if !isSetID && !isSetAlias {
			clusterListData := cluster.GetClusterList(viper.Get("roost_auth_token").(string))
			if clusterListData.Count < 1 || len(clusterListData.Clusters) < 1 {
				fmt.Println("No clusters are found")
				return
			}
			var custToken = []string{}
			for _, clusterData := range clusterListData.Clusters {
				if clusterData.IsActive == true && clusterData.ClusterType == "roost" {
					custToken = append(custToken, clusterData.CustomerToken)
				}
			}
			if len(custToken) < 1 {
				fmt.Println("No running clusters to show the Cluster UI interface")
				return
			}

			clusterAliasInput := utils.PromptSelectInput(custToken, "Select the cluster you want to get connect UI")
			if clusterAliasInput != "" {
				for _, clusterData := range clusterListData.Clusters {
					if clusterData.CustomerToken == clusterAliasInput {
						clusterIP := clusterData.PublicIP
						err = utils.Openbrowser("http://" + clusterIP + ":30070/app")
					}
				}
			}
		}
		cobra.CheckErr(err)
	},
	Example: `
	roost cluster ui
	roost cluster ui --id 1
	roost cluster ui --alias exampleAlias
	`,
}


func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterCreateCmd)
	clusterCmd.AddCommand(clusterStopCmd)
	clusterCmd.AddCommand(clusterKubeconfigCmd)
	clusterCmd.AddCommand(clusterDeleteCmd)
	clusterCmd.AddCommand(clusterListCmd)
	clusterCmd.AddCommand(clusterDetailsCmd)
	clusterCmd.AddCommand(clusterUICmd)

	clusterCreateCmd.Flags().String("email", "", "REQUIRED. Customer email.")
	clusterCreateCmd.Flags().String("alias", "", "The Alias of the Cluster to be created")
	clusterCreateCmd.Flags().StringP("namespace", "n", "roostcli", "Namespace in which cluster is to be made (default: roostcli)")
	clusterCreateCmd.Flags().String("ami", "ubuntu jammy jellyfish 22.04", "Ami to be used in the cluster (default: ubuntu jammy jellyfish 22.04)")
	clusterCreateCmd.Flags().String("instance-type", "t3.small", "Instance type of the cluster (default: t3.small)")
	clusterCreateCmd.Flags().String("disk-size", "50GB", "Disk size to use in the cluster (minimum and delfault: 50GB)")
	clusterCreateCmd.Flags().String("region", "ap-south-1", "Region to create the AWS cluster in (default: ap-south-1)")
	clusterCreateCmd.Flags().Int("expiry", 1, "The expiry time of the cluster(in hours) (default: 1)")
	clusterCreateCmd.Flags().String("k8s", "1.22.2", "The k8s version to use in the cluster (default: 1.22.2)")
	clusterCreateCmd.Flags().Int("nodes", 1, "The number of worker nodes in the cluster (default: 1)")

	clusterStopCmd.Flags().Int32Slice("id", []int32{}, "Stop Cluster with ID instead of alias. Provide multiple values separated by commas to stop multiple clusters at once.")
	clusterStopCmd.Flags().StringSlice("alias", []string{}, "Stop Cluster with Alias. Provide multiple values separated by commas to stop multiple clusters at once.")
	clusterStopCmd.MarkFlagsMutuallyExclusive("id", "alias")

	clusterDeleteCmd.Flags().Int32Slice("id", []int32{}, "Delete Cluster with ID instead of alias. Provide multiple values separated by commas to delete multiple clusters at once.")
	clusterDeleteCmd.Flags().StringSlice("alias", []string{}, "Delete Cluster with Alias. Provide multiple values separated by commas to delete multiple clusters at once.")
	clusterDeleteCmd.MarkFlagsMutuallyExclusive("id", "alias")

	clusterListCmd.Flags().Bool("running", false, "Get all running clusters")
	clusterListCmd.Flags().Bool("stopped", false, "Get all stopped clusters")

	clusterKubeconfigCmd.Flags().Int32Slice("id", []int32{}, "Get kubeConfig of a cluster with ID. Provide multiple values separated by commas to get kubeconfig of multiple clusters at once.")
	clusterKubeconfigCmd.Flags().StringSlice("alias", []string{}, "Get kubeConfig of a cluster with Alias. Provide multiple values separated by commas to get kubeconfig of multiple clusters at once.")
	clusterKubeconfigCmd.MarkFlagsMutuallyExclusive("id", "alias")

	clusterDetailsCmd.Flags().Int32("id", -1, "Get the details of a cluster with ID")
	clusterDetailsCmd.Flags().String("alias", "", "Get the details of a cluster with Alias")
	clusterDetailsCmd.Flags().StringP("output", "o", "", "Specify the output path to get the JSON of cluster Details")
	clusterDetailsCmd.MarkFlagsMutuallyExclusive("id", "alias")

	clusterUICmd.Flags().Int32("id", -1, "open the UI of a cluster by using it's ID.")
	clusterUICmd.Flags().String("alias", "", "open the UI of a cluster by using its Alias.")
	clusterUICmd.MarkFlagsMutuallyExclusive("id", "alias")

}
