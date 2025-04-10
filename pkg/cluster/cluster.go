package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ZB-io/internal/roostcli/pkg/spinner"
	"github.com/ZB-io/internal/roostcli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateClusterRequest can be used to accept data from promptUI. If prompt tag is not used, field name would apper in UI.
// Prompted attributes must be an exported field with supported data type int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, ,float32, float64, string
type CreateClusterRequest struct {
	Alias          string `json:"alias" prompt:"Cluster Alias"`
	Email          string `json:"customer_email"`
	Namespace      string `json:"namespace"`
	Ami            string `json:"ami"`
	InstanceType   string `json:"instance_type"`
	DiskSize       string `json:"disk_size"`
	Region         string `json:"region"`
	ClusterExpiry  int    `json:"cluster_expires_in_hours"`
	K8sVersion     string `json:"k8s_version"`
	WorkerNodes    int    `json:"num_workers"`
	RoostAuthToken string `json:"roost_auth_token"`
}

type ClusterKubeconfig struct {
	Alias          string `json:"cluster_alias"`
	RoostAuthToken string `json:"app_user_id"`
}

type ClusterKubeconfigResponse struct {
	ClusterID  int    `json:"cluster_id"`
	Kubeconfig string `json:"kubeconfig"`
	PublicIP   string `json:"public_ip"`
}

type ClusterStopObj struct {
	Alias          string `json:"alias"`
	RoostAuthToken string `json:"roost_auth_token"`
}

type ClusterListObj struct {
	RoostAuthToken string `json:"app_user_id"`
}

type ClusterListResponse struct {
	Clusters []ClusterList `json:"clusters"`
	Count    int           `json:"count"`
}

type ClusterList struct {
	Id             int    `json:"id"`
	Alias          string `json:"alias"`
	CustomerEmail  string `json:"customer_email"`
	CustomerToken  string `json:"customer_token"`
	CreatedOn      string `json:"created_on"`
	RunningOn      string `json:"running_on"`
	StoppedOn      string `json:"stopped_on"`
	IsActive       bool   `json:"is_active"`
	PublicIP       string `json:"public_ip"`
	NumNodes       int    `json:"num_nodes"`
	ClusterType    string `json:"cluster_type"`
	EnvType        string `json:"env_type"`
	FailureMsg     string `json:"failure_message"`
	FailureDetails string `json:"failure_details"`
	StatusMsg      string `json:"status_message"`
}

type ClusterApiResponse struct {
	ClusterRespMessage string `json:"message"`
}

func GetClusterDetails(clusterid int, alias string) (ClusterList, error) {

	clusterListObj := ClusterListObj{}
	var ClusterInfo ClusterListResponse
	clusterListObj.RoostAuthToken = viper.Get("roost_auth_token").(string)
	spinner := spinner.NewSpinner()
	spinner.Start("Fetching the cluster list")
	apiEndPoint := "/api/application/getAppUserClusters"
	reqBuff, err := json.Marshal(clusterListObj)
	if err != nil {
		spinner.Stop(false)
		fmt.Println(err)
		os.Exit(0)
	}
	body := bytes.NewReader(reqBuff)

	_, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
	if err != nil {
		spinner.Stop(false)
		fmt.Println(err)
		os.Exit(0)
	}

	err = json.Unmarshal(resp, &ClusterInfo)
	if err != nil {
		spinner.Stop(false)
		fmt.Println(err)
		os.Exit(0)
	}

	if alias != "" {
		for _, clusterData := range ClusterInfo.Clusters {
			if clusterData.CustomerToken == alias {
				spinner.Stop(true)
				return clusterData, nil
			}
		}
	} else {
		for _, clusterData := range ClusterInfo.Clusters {
			if clusterData.Id == clusterid {
				spinner.Stop(true)
				return clusterData, nil
			}
		}
	}
	spinner.Stop(false)
	err = fmt.Errorf("Invalid Cluster ID or alias provided")
	return ClusterList{}, err
}

//ClusterList is used get cluster details and list,To be used in teams section also
func GetClusterList(authToken string) (list ClusterListResponse) {
	clusterListObj := ClusterListObj{}
	var getClusterList ClusterListResponse
	clusterListObj.RoostAuthToken = authToken
	apiEndPoint := "/api/application/getAppUserClusters"
	reqBuff, err := json.Marshal(clusterListObj)
	cobra.CheckErr(err)
	body := bytes.NewReader(reqBuff)

	status, resp, err := utils.HTTPClientRequest(http.MethodPost, apiEndPoint, "", body)
	cobra.CheckErr(err)
	var apiresp ClusterApiResponse
	json.Unmarshal(resp, &apiresp)

	if status != http.StatusCreated {
		fmt.Println("Unable to fetch the cluster list associated with the IP", status, string(apiresp.ClusterRespMessage))
	}
	err = json.Unmarshal(resp, &getClusterList)
	cobra.CheckErr(err)
	return getClusterList
}


