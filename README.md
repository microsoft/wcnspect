# Wcnspect

> **W**indows **c**ontainer **n**etworking **s**tack in**spect**or

## Features

Wcnspect features four commands:

* `Capture`: runs a packet capture on Windows nodes, Has the capability to filter on pods, IPs, MACs, ports, protocols, and packet type (all, flow, or drop).
* `Counter`: will retrieve packet counter tables from windows nodes. It only outputs a table on nodes currently running a capture.
* `Vfp-counter`: will retrieve packet counter tables from the specified pod's VFP port. If specified, the counters from the Host vNIC VFP port and External Adapter VFP port.
* `Hns`: will print HNS resources in Windows nodes. Can specify `all`, `endpoints`, `loadbalancers`, `namespaces`, or `networks`. Can request json output.

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

## Wcnspect Server

### Deploying the Wcnspect Server as a DaemonSet (Recommended)

You can apply the [wcnspectserv-daemon.yml](manifest/wcnspectserv-daemon.yml) to deploy the server as a host process container on all Windows nodes. 

The container image is published under `ghcr.io/microsoft/wcnspect:latest`. The [Dockerfile](manifest/Dockerfile) used for the image can be found [here](./manifest/Dockerfile). 

Note that the [manifest](./manifest) directory also contains sample web server deployments for Windows Server 2019 and Windows Server 2022.

## Wcnspect Client
The client needs to be executed as a standalone binary from either a Windows or a Linux VM in the same network (jumpbox).

The Wcnspect client requires access to the [Kubernetes cluster config](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/). 

By default, Wcnspect client will search for a file named `config` in the `$HOME/.kube` directory. Otherwise, it will use the $KUBECONFIG environment variable.

By default, most commands pull information from *all* Windows nodes.
Consequently, when using commands, the user should reference node names and pod names for better filtering of results.

> in-depth documentation and examples are available with the -h flag on any command

```shell
wcnspect -h
wcnspect capture -h
wcnspect hns all -h
```

For commands that accept lists, input should be comma-separated and without spaces. For example, if we want to capture for 10 seconds on nodes named `win1`, `win2`, and `win3`, while also filtering only for TCP packets, we could do the following:

> sample capture command

```shell
wcnspect capture nodes win1,win2,win3 -t TCP -d 10
```

The command will be routed to each node's internal IP on the cluster. It should be noted that if we don't pass a duration, the command will run indefinitely. Additionally, we can terminate the process on the referenced nodes at any time with `Ctrl+C`.

Note that if we pass the `--counters-only` flag to the `capture` command, then packet output won't be displayed and the counter table will only be displayed once the command is finished running.

> sample capture command using --counters-only

```shell
wcnspect capture nodes win1 --counters-only
```

Importantly, while the `vfp-counter` command runs on its own (given a pod), the `counter` command is tied to running instances of the `capture` command. Consequently, in order for it to output a table on any given node, a capture must be run on that node at the same time. The table will output packet counts tied to that capture.

```shell
wcnspect capture nodes win1,win3 -t TCP
wcnspect counter
```

## Assumptions

Currently, this project's code makes the following assumptions:

* The port that Wcnspect server uses is 50051 (this is currently required on the client-side).

## TODO
  * Support for other ports on the Wcnspect client.


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
