package util

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/creasty/defaults"
	yaml "gopkg.in/yaml.v3"
)

var (
	CommonConfig *Common
)

// Struct for parsing config files

type NodeTracker struct {
	Hostname string `yaml:"hostname"`
	RpcHost  string `yaml:"rpc_host"`
	RpcPort  int    `yaml:"rpc_port"`
}
type EtcdServer struct {
	Hostname         string   `yaml:"hostname"`
	ListenClientUrls []string `yaml:"listen_client_urls"`
}

type StorageServer struct {
	Hostname         string `yaml:"hostname"`
	RpcPort          int    `yaml:"rpc_port"`
	PlasmaSocket     string `yaml:"plasma_socket"`
	PlasmaMemoryByte int    `yaml:"plasma_memory_byte"`
}

type StorageClient struct {
	Hostname      string `yaml:"hostname"`
	StorageServer string `yaml:"storage_server"`
}

type Common struct {
	DialTimeout               time.Duration `yaml:"dial_timeout"          default:"10s"`
	RequestTimeout            time.Duration `yaml:"request_timeout"       default:"60s"`
	FetchTaskRetryMax         int           `yaml:"fetch_task_retry_max"  default:"5"`
	FetchSrcVirtualNodeNumber int           `yaml:"fetch_src_virtual_node_number" default:"3"`
	FetchSrcVirtualNodeFanout int           `yaml:"fetch_src_virtual_node_fanout" default:"1" `
	BackgroundTaskLimit       int           `yaml:"background_task_limit" default:"20"`
	TaskQueueCap              int           `yaml:"task_queue_cap"        default:"40"`
	GrpcMaxCallRecvMsgSize    int           `yaml:"grpc_max_call_recv_msg_size" default:"1073741824"`
}

type Config struct {
	configFilePath string
	Common         Common          `yaml:"common"`
	EtcdServers    []EtcdServer    `yaml:"etcd_servers"`
	NodeTrackers   []NodeTracker   `yaml:"node_trackers"`
	StorageServers []StorageServer `yaml:"storage_servers"`
	StorageClients []StorageClient `yaml:"storage_clients"`
}

func ReadConfig(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	defaults.Set(&conf.Common)
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		return nil, err
	}
	conf.configFilePath = path
	CommonConfig = &conf.Common
	return conf, nil
}

func (conf *Config) NodeTracker(hostname string) (*NodeTracker, error) {
	for _, tracker := range conf.NodeTrackers {
		if tracker.Hostname == hostname {
			return &tracker, nil
		}
	}
	return nil, fmt.Errorf("node tracker (with hostname %s) is not found in configure file %s", hostname, conf.configFilePath)
}

func (conf *Config) StorageServer(hostname string) (*StorageServer, error) {
	var defaultServer *StorageServer
	for _, server := range conf.StorageServers {
		switch server.Hostname {
		case hostname:
			return &server, nil
		case "*":
			defaultServer = &server
		}
	}
	if defaultServer != nil {
		return defaultServer, nil
	} else {
		return nil, fmt.Errorf("storage server (with hostname %s) is not found in configure file %s", hostname, conf.configFilePath)
	}
}

func (conf *Config) EtcdServerUrls() ([]string, error) {
	var l []string
	for _, server := range conf.EtcdServers {
		l = append(l, server.ListenClientUrls...)
	}
	if len(l) == 0 {
		return nil, fmt.Errorf("etcd_servers.listen_client_urls is not found in configure file %s", conf.configFilePath)
	}
	return l, nil
}
