package nacosClient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type INacosClient interface {
}

type NacosClient struct {
	logger     log.Logger
	httpClient http.Client
}

type ServersInfo struct {
	Servers []struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		State      string `json:"state"`
		ExtendInfo struct {
			LastRefreshTime int64 `json:"lastRefreshTime"`
			RaftMetaData    struct {
				MetaDataMap struct {
					NamingPersistentService struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_persistent_service"`
				} `json:"metaDataMap"`
			} `json:"raftMetaData"`
			RaftPort string `json:"raftPort"`
			Version  string `json:"version"`
		} `json:"extendInfo"`
		Address       string `json:"address"`
		FailAccessCnt int    `json:"failAccessCnt"`
	} `json:"servers"`
}

func (c *NacosClient) GetClusterNodes(ip string) (ServersInfo, error) {
	servers := ServersInfo{}
	//增加支持ipV6 pod状态探测
	var resp *http.Response
	var err error

	if strings.Contains(ip, ":") {
		resp, err = c.httpClient.Get(fmt.Sprintf("http://[%s]:8848/nacos/v1/ns/operator/servers", ip))
	} else {
		resp, err = c.httpClient.Get(fmt.Sprintf("http://%s:8848/nacos/v1/ns/operator/servers", ip))
	}

	if err != nil {
		return servers, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return servers, err
	}

	err = json.Unmarshal(body, &servers)
	if err != nil {
		fmt.Printf("%s\n", body)
		return servers, fmt.Errorf(fmt.Sprintf("instance: %s ; %s ;body: %v", ip, err.Error(), string(body)))
	}
	return servers, nil
}

type ServersInfoV2 struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		IP         string `json:"ip"`
		Port       int    `json:"port"`
		State      string `json:"state"`
		ExtendInfo struct {
			LastRefreshTime int64 `json:"lastRefreshTime"`
			RaftMetaData    struct {
				MetaDataMap struct {
					NamingInstanceMetadata struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_instance_metadata"`
					NamingPersistentService struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_persistent_service"`
					NamingPersistentServiceV2 struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_persistent_service_v2"`
					NamingServiceMetadata struct {
						Leader          string   `json:"leader"`
						RaftGroupMember []string `json:"raftGroupMember"`
						Term            int      `json:"term"`
					} `json:"naming_service_metadata"`
				} `json:"metaDataMap"`
			} `json:"raftMetaData"`
			RaftPort       string `json:"raftPort"`
			ReadyToUpgrade bool   `json:"readyToUpgrade"`
			Version        string `json:"version"`
		} `json:"extendInfo"`
		Address       string `json:"address"`
		FailAccessCnt int    `json:"failAccessCnt"`
		Abilities     struct {
			RemoteAbility struct {
				SupportRemoteConnection bool `json:"supportRemoteConnection"`
				GrpcReportEnabled       bool `json:"grpcReportEnabled"`
			} `json:"remoteAbility"`
			ConfigAbility struct {
				SupportRemoteMetrics bool `json:"supportRemoteMetrics"`
			} `json:"configAbility"`
			NamingAbility struct {
				SupportJraft bool `json:"supportJraft"`
			} `json:"namingAbility"`
		} `json:"abilities"`
	} `json:"data"`
}

func (c *NacosClient) GetClusterNodesV2(ip string) (ServersInfoV2, error) {
	serversInfoV2 := ServersInfoV2{}
	resp, err := c.httpClient.Get(fmt.Sprintf("http://%s:8848/nacos/v2/core/cluster/node/list", ip))
	if err != nil {
		return serversInfoV2, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return serversInfoV2, err
	}

	err = json.Unmarshal(body, &serversInfoV2)
	if err != nil {
		fmt.Printf("%s\n", body)
		return serversInfoV2, fmt.Errorf(fmt.Sprintf("instance: %s ; %s ;body: %v", ip, err.Error(), string(body)))
	}
	return serversInfoV2, nil
}

//func (c *CheckClient) getClusterNodesStaus(ip string) (bool, error) {
//	str, err := c.getClusterNodes(ip)
//	if err != nil {
//		return false, err
//	}
//
//}
