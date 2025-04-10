# RoostCLI

This documentation enables us to interact with Roost CLI wizard, a CLI wizard to interact with roost clusters, teams and environments.

## Intial Setup to Interact with the CLI.
Note:In this example we will take app.roost.io as an ent Server for all the purposes
We currently are supporting google OATH to authenticate with Roost.<br />

After the setup of roostcli in your local system, to use roost-cli, first run the 'roost login' command and log-in to the roost CLI wizard after providing the roost enterprise server and selecting login using google, this command will open up a browser window, where you will need to complete your log-in process.
Please enter the command, <br />
Command:- roost login <br />

It will ask for the Ent server in which your roost site is hosted, After the submission of ENT server the prompt will asked for OATH. <br />
![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/examples/roost_login.gif) <br />

Alternatively, you can run the 'roost configure' command to manually enter your enterprise server, roost-auth_token, and roost_jwt_token. This information will be stored in a config file present in .roost/configzbio
To do so, Please enter the command, <br />
Command:- roost congfigure
    
## Interaction with Roost Clusters.    
The 'roost cluster' command includes an array of sub-commands that will allow interaction with roost clusters. <br />
Subcommands:
## Creation of Roost Clusters
- create <br />
Usage:
    The cluster create subcommand will allow creation of roost clusters, to create a cluster simply run the 'roost cluster create' command and provide the customer email in the prompt that appears, you can also set the default values in the prompt to create a cluster that fits your usage. <br />
Example: <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/create.gif) <br />
Flags:
    You can also use flags to provide the cluster specifications for the cluster to be created, to check a list of the supported flags, run 'roost cluster create help', 'roost cluster create -h' or 'roost cluster create --help' <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/create_flag.gif) <br />

## List Roost Clusters
- list <br />
Usage:
    The cluster list subcommand will display a list of all the available clusters, their status, aliases, etc.
Example: <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/list.gif) <br />
Flags:
    You can also use flags such as --running or --stopped to filter to display only the clusters with the specified status. <br />   
## Get details of a specific Roost Cluster
- get-details <br />
Usage:
    The get-details subcommand will allow you to get the details of a specific cluster, on running the command, a list of all the currently available clusters will be displayed, the required cluster can be selected from the given list.
Example: <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/get-details.gif) <br />
Flags:
    You can also get the details of a specific cluster by providing its ID by using the --id flag or it's alias by using the --alias flag. <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/get-details_flag.gif) <br />
## Get the KubeConfig of a specific Roost Cluster
- get-kubeconfig <br /> 
Usage:
    The get-kubeconfig subcommand will download the kubeconfig of the requested cluster and put it under .kube/roostconfig/ with the alias of the requested cluster as the file name, if successful, it will give you the command to connect with the cluster.
Example: <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/get-kubeconfig.gif) <br />
Flags:
    You can also get the kubeconfig of a specific cluster by providing its ID by using the --id flag or it's alias by using the --alias flag. <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/get-kubeconfig_flag.gif) <br />
## Open the service-fitness view of a Roost Cluster    
- ui <br />
Usage:
    The ui subcommand will open the ui or the service-fitness view of a requested cluster in your browser window, on running the command, it will show you a list of all the running clusters, and the requested cluster can be selected from the given list.
Example:<br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/ui.gif) <br />
Flags:
    You can also stop a specific cluster by providing its ID by using the --id flag or it's alias by using the --alias flag. <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/ui_flag.gif) <br />
## Stop a Roost Cluster
 - stop <br />
Usage:
    The cluster stop subcommand will stop a requested or running cluster, on running the command, a list of all the currently runnning/requested clusters will be shown and the target cluster can be selected from the list.
Example: <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/stop.gif) <br />
Flags:
    You can also stop a specific cluster by providing its ID by using the --id flag or it's alias by using the --alias flag. <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/stop_flag.gif) <br />
## Delete a Roost Cluster    
- delete <br />
Usage:
    The cluster delete subcommand can be used to delete a roost cluster, upon running the command, a list of all the currently available clusters will be displayed and the required cluster can be selected from the list.
    Note: It's always recommended that you delete your old/stopped clusters that you are not using.
Example: <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/delete.gif) <br />
Flags:
    You can also delete a specific cluster by providing its ID by using the --id flag or it's alias by using the --alias flag. <br />
    ![](https://github.com/ZB-io/internal/blob/RoostCLI/roostcli/gifs/cluster/delete_flag.gif) <br />
