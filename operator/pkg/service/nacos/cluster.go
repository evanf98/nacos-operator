package nacosClient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	nacosgroupv1alpha1 "nacos.io/nacos-operator/api/v1alpha1"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

const (
	NacosV2ServersApi = "nacos/v2/core/cluster/node/list"
	NacosV1ServersApi = "nacos/v1/ns/operator/servers"
	NacosLoginApi     = "nacos/v1/auth/login"
	NacosDefaultPass  = "nacos"
)

var cluster_log = ctrl.Log.WithValues("cluster")

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

func (c *NacosClient) GetClusterNodes(ip string, nacos *nacosgroupv1alpha1.Nacos) (ServersInfo, error) {
	servers := ServersInfo{}
	//增加支持ipV6 pod状态探测
	var resp *http.Response
	var req *http.Request
	var err error
	client := c.httpClient

	if nacos.Spec.Certification.Enabled { // 开启认证
		cluster_log.Info("nacos v1 cluster nodes with auth", "ip", ip)
		accessToken, err1 := c.GetNacosAccessToken(ip)
		if err1 != nil {
			return ServersInfo{}, err1
		}
		if strings.Contains(ip, ":") {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://[%s]:8848/%s?accessToken=%s", ip, NacosV1ServersApi, accessToken), nil)
		} else {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://%s:8848/%s?accessToken=%s", ip, NacosV1ServersApi, accessToken), nil)
		}
	} else { // 非认证模式下
		cluster_log.Info("nacos v1 cluster nodes without auth", "ip", ip)
		if strings.Contains(ip, ":") {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://[%s]:8848/%s", ip, NacosV1ServersApi), nil)
		} else {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://%s:8848/%s", ip, NacosV1ServersApi), nil)
		}

	}
	resp, err = client.Do(req)
	if err != nil {
		return servers, err
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

func (c *NacosClient) GetClusterNodesV2(ip string, nacos *nacosgroupv1alpha1.Nacos) (ServersInfoV2, error) {
	serversInfoV2 := ServersInfoV2{}
	var resp *http.Response
	var req *http.Request
	var err error
	client := c.httpClient

	if err != nil {
		return serversInfoV2, err
	}
	// todo: FIXME 默认账号密码，如自定义需获取
	if nacos.Spec.Certification.Enabled { // 开启认证
		cluster_log.Info("nacos v2 cluster nodes with auth", "ip", ip)
		accessToken, err1 := c.GetNacosAccessToken(ip)
		if err1 != nil {
			return ServersInfoV2{}, err1
		}
		if strings.Contains(ip, ":") {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://[%s]:%d/%s?accessToken=%s", ip, 8848, NacosV2ServersApi, accessToken), nil)
		} else {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://%s:%d/%s?accessToken=%s", ip, 8848, NacosV2ServersApi, accessToken), nil)
		}
	} else { // 非认证模式下
		cluster_log.Info("nacos v2 cluster nodes without auth", "ip", ip)
		if strings.Contains(ip, ":") {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://[%s]:%d/%s", ip, 8848, NacosV2ServersApi), nil)
		} else {
			req, err = http.NewRequest("GET", fmt.Sprintf("http://%s:%d/%s", ip, 8848, NacosV2ServersApi), nil)
		}

	}
	resp, err = client.Do(req)
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

type AccessResp struct {
	AccessToken string `json:"accessToken"`
	TokenTTL    int    `json:"tokenTtl"`
	GlobalAdmin bool   `json:"globalAdmin"`
	Username    string `json:"username"`
}

func (c *NacosClient) GetNacosAccessToken(ip string) (string, error) {
	accessResp := AccessResp{}
	accessToken := ""
	payload := strings.NewReader(fmt.Sprintf("username=nacos&password=%s", NacosDefaultPass))
	resp, err := c.httpClient.Post(fmt.Sprintf("http://%s:8848/%s", ip, NacosLoginApi), "application/x-www-form-urlencoded", payload)
	if err != nil {
		return accessToken, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return accessToken, err
	}
	err = json.Unmarshal(body, &accessResp)
	if err != nil {
		fmt.Printf("%s\n", body)
		return accessResp.AccessToken, fmt.Errorf(fmt.Sprintf("instance: %s ; %s ;body: %v", ip, err.Error(), string(body)))
	}
	accessToken = accessResp.AccessToken
	return accessToken, nil
}

//func (c *CheckClient) getClusterNodesStaus(ip string) (bool, error) {
//	str, err := c.getClusterNodes(ip)
//	if err != nil {
//		return false, err
//	}
//
//}
