# Winspect - A Network Diagnostics Tool

> A network inspector for the Windows container networking stack on Kubernetes

## Features

Currently, the winspect tool features four commands -

* Capture: runs a packet capture on Windows nodes, Has the capability to filter on pods, IPs, MACs, ports, protocols, and packet type (all, flow, or drop).
* Counter: will retrieve packet counter tables from windows nodes. It only outputs a table on nodes currently running a capture.
* Vfp-counter: will retrieve packet counter tables from the specified pod's VFP port. If specified, the counters from the Host vNIC VFP port and External Adapter VFP port.
* Hns: will retrieve HNS logs from Windows nodes. Can specify `all`, `endpoints`, `loadbalancers`, `namespaces`, or `networks`. Can request json output.

### Building

This project requires `Go 1.18`. After cloning this repo and installing its dependencies with Go, refer to the following:

All executables will be placed in `./out/bin`. Upon making the client, two executables will be built: one for Windows and one for Linux.

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

It should be noted that while the client is cross-platform, the server can only run on Windows.

## Winspect Server

### Deploying the Winspect Server as a DaemonSet (Recommended)

You can apply the [winspectserv-daemon.yml](manifest/winspectserv-daemon.yml) to deploy the server as a host process container on all Windows nodes. The only way to do this, as of now, is for the user to create their own Azure Container Registry (ACR) and upload the image through there.

The steps for using ACR are outlined <a href="https://docs.microsoft.com/en-us/azure/container-registry/container-registry-tutorial-quick-task" target="_blank">in these docs.</a>
The [Dockerfile](manifest/Dockerfile) used for the original image is also contained in [manifest](./manifest). 

Note that the [manifest](./manifest) directory also contains sample web server deployments for Windows Server 2019 and Windows Server 2022.

 
### Running Winspect Server directly

If you decide not to deploy the server as a container and manually download it to a node, then it must be run with elevated Administrator permissions. You can run it as follows:

> running on the default port of 50051

```shell
./winspectserv
```

> running on a chosen port of 43058

```shell
./winspectserv -p 43058
```
> **NOTE: The Winspect client only supports connecting to port 50051 currently**

## Winspect Client
The client needs to be executed as a standalone binary from either a Windows or a Linux VM in the same network (jumpbox).

The winspect client reads in the user's `.kube` config file and uses the `default` namespace. 

By default, most commands pull information from *all* Windows nodes.
Consequently, when using commands, the user should reference node names and pod names for better filtering of results.

> in-depth documentation and examples are available with the -h flag on any command

```shell
winspect -h
winspect capture -h
winspect hns all -h
```

For commands that accept lists, input should be comma-separated and without spaces. For example, if we want to capture for 10 seconds on nodes named `win1`, `win2`, and `win3`, while also filtering only for TCP packets, we could do the following:

> sample capture command

```shell
winspect capture nodes win1,win2,win3 -t TCP -d 10
```

The command will be routed to each node's internal IP on the cluster. It should be noted that if we don't pass a duration, the command will run indefinitely. Additionally, we can terminate the process on the referenced nodes at any time with `Ctrl+C`.

Note that if we pass the `--counters-only` flag to the `capture` command, then packet output won't be displayed and the counter table will only be displayed once the command is finished running.

> sample capture command using --counters-only

```shell
winspect capture nodes win1 --counters-only
```

Importantly, while the `vfp-counter` command runs on its own (given a pod), the `counter` command is tied to running instances of the `capture` command. Consequently, in order for it to output a table on any given node, a capture must be run on that node at the same time. The table will output packet counts tied to that capture.

```shell
winspect capture nodes win1,win3 -t TCP
winspect counter
```

## Assumptions

Currently, this project's code makes the following assumptions:

* The port that winspect server nodes use is 50051 (this is currently required on the client-side).
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
