# Winspect - A Network Diagnostics Tool

> A network inspector for the Windows container networking stack on Kubernetes

## Installation

After cloning this repo and installing its dependencies with Go, refer to the following:

### Building the Client and Server

All executables will be placed in `./out/bin`.

> to build the client and server

```shell
make all
```

> to clean your repo of executables

```shell
make clean
```

> the server and client can also be built individually

```shell
make client
make server
```

### Building and Deploying the Server as a Container

Currently, deploying the server executable on a cluster entails creating an image for it and passing this to the daemonset `.yml` file in `manifest`, then applying it. The most convient way to do this, as of now, is for the user to create their own Azure Container Registry (ACR) and upload the image through there.

The steps for using ACR are outlined <a href="https://docs.microsoft.com/en-us/azure/container-registry/container-registry-tutorial-quick-task" target="_blank">in these docs.</a>
The Dockerfile used for the original image is also contained in `manifest`.

It should be noted that the port on which the server runs can be changed by modifying the command input to the daemonset file in `manifest`. For example:

> running on the default port of 50051

```shell
./winspectserv
```

> running on a chosen port of 43058

```shell
./winspectserv -p 43058
```

## Features

Currently, the winspect tool features four commands -

* Capture: runs a packet capture on windows nodes, Has the capability to filter on pods, IPs, MACs, ports, protocols, and packet type (all, flow, or drop).
* Counter: will retrieve packet counter tables from windows nodes. It only outputs a table on nodes currently running a capture.
* Vfp-counter: will retrieve packet counter tables from the specified pod's Port VFP. If specified, the pod's Host vNIC Port VFP and External Adapter Port VFP.
* Hns: will retrieve HNS logs from windows nodes. Can specify all, endpoints, loadbalancers, namespaces, or networks. Can request json output. 

## Example

If you decide not to deploy the server as a container and manually download it to a node, then it must be run from a CLS with Admin permissions:

> running on the default port of 50051

```shell
./winspectserv
```

The winspect tool pulls on the user's `.kube` config file and the `default` namespace. By default, most commands pull information from all windows nodes.
Consequently, when using commands, the user should reference node names and pod names for better filtering of results.

> in-depth documentation and examples are available with the -h flag on any command

```shell
./winspect -h
./winspect capture -h
./winspect hns all -h
```

For commands that accept lists, input should be comma-separated and without spaces. For example, if we want to capture for 10 seconds on nodes named `win1`, `win2`, and `win3`, while also filtering only for TCP packets, we could do the following:

> sample capture command

```shell
./winspect capture nodes win1,win2,win3 -t TCP -d 10
```

The command will be routed to each node's internal IP on the cluster. It should be noted that if we don't pass a duration, the command will run indefinitely. Additionally, we can terminate the process on the referenced nodes at any time with `Ctrl+C`.

Note that if we pass the `--counters-only` flag to the `capture` command, then packet output won't be displayed and the counter table will only be displayed once the command is finished running.

> sample capture command using --counters-only

```shell
./winspect capture nodes win1 --counters-only
```

Importantly, while the `vfp-counter` command runs on its own (given a pod), the `counter` command is tied to running instances of the `capture` command. Consequently, in order for it to output a table on any given node, a capture must be run on that node at the same time. The table will output packet counts tied to that capture.

```shell
./winspect capture nodes win1,win3 -t TCP
./winspect counter
```

## Assumptions

Currently, this project's code makes the following assumptions:

* The port that winspect server nodes use is 50051 (assumed on client side).
* When applying `winspectserv-daemon.yml`, the user has access to the ACR referenced in the file.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
