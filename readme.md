[![Go Report Card](https://goreportcard.com/badge/github.com/miky4u2/RAserver)](https://goreportcard.com/report/github.com/miky4u2/RAserver)
[![license](https://img.shields.io/github/license/miky4u2/RAserver.svg)](https://github.com/miky4u2/RAserver/blob/master/LICENSE)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/miky4u2/RAserver)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/miky4u2/RAserver)](https://github.com/miky4u2/RAserver/releases/tag/v0.1.0)



# Remote Admin Server

**RAserver is functioning but not tested on Osx at all. It might still contain some bugs I haven't come across yet but feel free to try it, use it, expand on it...**

RAserver is one of two related small and simple tools, particularly useful for system administration of multiple servers and remote task automation. Those tools are the result of my first project in GO so the code design might not be the best but they do what they were designed to do just fine :-).  
Running services allowing remote program execution can represent a high security risk when not properly firewalled, make sure to setup appropriate firewall rules prior to deployment or do not use at all if you don't know what you are doing.

RAserver is of no use by itself, it was designed as an add-on to RAagent. RAserver acts as the **'modules and binaries repository'** for one or several RAagents, making bulk modules deployment and agents update easier. I have done my best to design those tools to run on Linux/Osx/Windows but only used them in production on Linux systems.

Check out the [RAagent repository](https://github.com/miky4u2/RAagent) for the related RAagent code and info.

## Description
- RAserver is a web server with a very simple web interface to remotely update RAagents running on remote servers/computers. It uses Secure REST API to send update requests to the agents which in turn pull their updates from the server. RAserver acts as the 'repository' for the agents making bulk modules synchronization and updates easier. 

- RAserver holds in a single location, the existing agent list their dedicated modules, binary, config file and TLS certificate. 

- From the RAserver web interface you can:
    1. Update agent's modules
    2. Fully update agents (modules, binary, config file & TLS Cert)
    3. Get the status, version and module list of agents
    4. Restart agents
    5. Shutdown agents
    6. Stop the server


## RAserver config file (conf/config.json)

- **bindIP** : IP the server should bind on. Leave blank to bind on all IPs.
- **bindPort** : Port the server should bind on.
- **allowedIPs** : Array of IPV4/IPV6 IPs allowed to access the server's web interface
- **validateAgentTLS** : If the agents use self signed TLS certificates, set this to *false*, otherwise set this to *true*.
- **logToFile** : *true* to log to log file, *false* or blank to log to stdOut  
- **logFile** : When logToFile is *true*, either leave this field blank to use the default log file **log/server.log** or provide a path to a log file. Use windows(\\) or linux(/) path format depending on which platform the server is running on. 


```json
{
    "bindIP": "",
    "bindPort": "8081",
    "allowedIPs": ["127.0.0.1", "::1"],
    "validateAgentTLS": false,
    "logFile":"",
    "logToFile": false
}
```
# Runtime deployment file layout
```
runtime
    |
    +--bin
    |    |
    |    +--server (or server.exe) Server executable
    |
    +--conf
    |     |
    |     +--config.json (Server config file)
    |     +--cert.pem (Server TLS certificate & chains)
    |     +--key.pem (Server TLS certificate key)
    |   
    +--log
    |    |
    |    +--server.log (Default log file. Auto truncated when it reaches 500k)
    |
    +--templates (Web interface GO templates)
    |
    +--agents (Agent files repository)
            |
            +--archives (Internally used by RAserver to store the auto created tar.gz archives)
            |
            +--binaries
            |         |
            |         +--linux   (Store the latest linux **ragent** binary in this folder)
            |         +--osx     (Store the latest osx **ragent** binary in this folder)
            |         +--windows (Store the latest windows **ragent.exe** binary in this folder)
            |         
            +--certs  (Store agent's certificates and keys in this folder)
            |
            +--configs
            |        |
            |        +--local  (Store agent's local configs in this folder)
            |        +--remote (store agent's remote configs in this folder)
            |        
            +--modules
                     |
                     +--agents
                     |       |
                     |       +--[agentID1] (Store agentID1's modules in this folder)
                     |       +--[agentID2] (Store agentID2's modules in this folder)
                     |       +--example.linux (etc..)
                     |
                     +--groups
                             |
                             +--defaultLinux (Linux type sharable modules in this folder)
                             +--defaultOsx (osx type sharable modules in this folder)
                             +--defaultWindows (Windows type sharable modules in this folder)
                             +--someOtherGroup (etc..)
```

## Agent Remote config files
- Those are the copies of the agents config files, see [RAagent repository](https://github.com/miky4u2/RAagent) for more info about those files.
- The files must be named [agentID].json (for example if an agent ID is "example.linux", its config file should be named "example.linux.json") 


## Agent Local config files 
- The files must be named [agentID].json (for example if an agent ID is "example.linux", its config file should be named "example.linux.json")
- Module's path format depends on the OS the server is running on, use / for Linux or double \\ for windows. 
- **agentIP** : Array of IPV4/IPV6. IP of the agent. This will be validated when the agent downloads the updates.
- **agentOS** : This is used to select the correct binary for the agent. Options are *linux*, *osx*, *windows*
- **modules** : map of *module_name* : *module_path* . Module name will be the module file name used on the agent side. The path is pointing to the locally stored module file (see above runtime file layout). See [RAagent repository](https://github.com/miky4u2/RAagent) for more info about modules.
- **TLScert** : The name of the TLS certificate that should be sent to the agent. For example "localhost" will send to the agent the local files "runtime/agents/certs/localhost.cert" and "runtime/agents/certs/localhost.key". The agent will place those files in its "runtime/conf/cert.pem and key.pem".


```json
# Example of an agent running on Linux and server on Windows
{
    "agentIP": ["127.0.0.1","::1"],
    "agentOS":"linux",
    "agentURL":"https://localhost2:8080",
    "modules":{
        "restart_services":"agents\\example.linux\\restart_services",
        "hello":"groups\\defaultLinux\\hello"
    },
    "TLScert":"localhost"
}

```
```json
# Example of an agent running on Windows and server on Linux
{
    "agentIP": ["127.0.0.1","::1"],
    "agentOS":"windows",
    "agentURL":"https://localhost:8080",
    "modules":{
        "start_notepad.cmd":"agents/example.pc/start_notepad.cmd",
        "hello.bat":"groups/defaultWindows/hello.bat"
    },
    "TLScert":"localhost"
}
```

## Building RAserver

The server executable needs to be built and placed in the /bin folder. The *makeosx.sh*, *makewindows.sh*, *makelinux.sh* can be used, alternatively this can be done as below.
```
# Linux
#
GOOS=linux GOARCH=amd64 go build -o ./runtime/bin/server  ./server/server.go

# Windows
#
GOOS=windows GOARCH=amd64 go build -o ./runtime/bin/server.exe  ./server/server.go

# OSX (not tested)
#
GOOS=darwin GOARCH=amd64 go build -o ./runtime/bin/server  ./server/server.go

```

## Starting RAserver

The server can manually be started when required by running bin/server  or bin/server.exe

I am not sure how to start it at boot or as a service on Windows but below is a Linux systemd service file that can be used for linux if you wish to start the server as a service on boot

```
[Unit]
Description = RAserver
After = network.target

[Service]
Type=
ExecStart = /path/to/RAserver/bin/server

[Install]
WantedBy = multi-user.target
```

## RAserver Screenshots

![Menu](https://storage.googleapis.com/githubassets/raserver1.jpg)
  

![Agent Control](https://storage.googleapis.com/githubassets/raserver2.jpg)
  

![Agent Update](https://storage.googleapis.com/githubassets/raserver3.jpg)
  

![Server Control](https://storage.googleapis.com/githubassets/raserver4.jpg)
  


