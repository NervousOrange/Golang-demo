package discover

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// 服务实例结构体
type InstanceInfo struct {
	ID                string                     `json:"ID"`
	Name              string                     `json:"Name"`
	Service           string                     `json:"Service,omitempty"` // 服务发现时返回的服务名
	Tags              []string                   `json:"tags,omitempty"`    // 标签，可用于服务过滤
	Address           string                     `json:"Address"`           // 服务实例 host
	Port              int                        `json:"Port"`
	Meta              map[string]string          `json:"Meta,omitempty"`    // 元数据
	EnableTagOverride bool                       `json:"EnableTagOverride"` // 是否允许标签覆盖
	Check             `json:"Check,omitempty"`   // 健康检查相关配置
	Weights           `json:"Weights,omitempty"` // 权重
}

type Check struct {
	DeregisterCriticalServiceAfter string   `json:"DeregisterCriticalServiceAfter"` // 多久之后注销服务
	Args                           []string `json:"Args,omitempty"`                 // 请求参数
	HTTP                           string   `json:"HTTP"`                           // 健康检查地址
	Interval                       string   `json:"Interval,omitempty"`             // Consul 主动检查间隔
	TTL                            string   `json:"TTL,omitempty"`                  // 服务实例主动维持心跳间隔，与Interval只存其一
}

// consul 支持 consul 主动调用服务的健康检查接口 / 实例主动提交健康数据（二选一）来维持心跳

type Weights struct {
	Passing int `json:"Passing"`
	Warning int `json:"Warning"`
}

type MyDiscoverClient struct {
	Host string // consul 的地址
	Port int
}

func NewMyDiscoverClient(consulHost string, consulPort int) (DiscoveryClient, error) {
	return &MyDiscoverClient{
		Host: consulHost,
		Port: consulPort,
	}, nil
}

func (consulClient *MyDiscoverClient) Register(serviceName, instanceId, healthCheckUrl, instanceHost string,
	instancePort int, meta map[string]string, logger *log.Logger) bool {
	// 1. 封装服务实例的元数据
	instanceInfo := &InstanceInfo{
		ID:                instanceId,
		Name:              serviceName,
		Address:           instanceHost,
		Port:              instancePort,
		Meta:              meta,
		EnableTagOverride: false,
		Check: Check{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + instanceHost + ":" + strconv.Itoa(instancePort) + healthCheckUrl,
			Interval:                       "15s",
		},
		Weights: Weights{
			Passing: 10,
			Warning: 1,
		},
	}
	byteData, _ := json.Marshal(instanceInfo)

	// 2. 向 consul 发送服务注册的请求
	req, err := http.NewRequest("PUT", "http://"+consulClient.Host+":"+strconv.Itoa(consulClient.Port)+"/v1/agent/service/register",
		bytes.NewReader(byteData))
	if err == nil {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		client := http.Client{}
		resp, err := client.Do(req)

		// 3. 检查注册结果
		if err != nil {
			log.Println("Register Service Error!")
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Println("Register Service Success!")
				return true
			} else {
				log.Printf("status code is: %d \n", resp.StatusCode)
				response := make([]byte, 1024)
				_, err := resp.Body.Read(response)
				if err != nil {
					log.Printf("read response error, err: %v \n", err)
				} else {
					log.Println(response)
				}
				log.Println("Register Service Error! Status code not equal 200")
			}
		}
	}
	return false
}

func (consulClient *MyDiscoverClient) DeRegister(instanceId string, logger *log.Logger) bool {
	req, _ := http.NewRequest("PUT",
		"http://"+consulClient.Host+":"+strconv.Itoa(consulClient.Port)+"/v1/agent/service/deregister/"+instanceId, nil)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Deregister Service Error!")
	} else {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			log.Println("Deregister Service Success!")
			return true
		} else {
			log.Println("Deregister Service Error!")
		}
	}
	return false
}

func (consulClient *MyDiscoverClient) DiscoverServices(serviceName string, logger *log.Logger) []interface{} {
	req, _ := http.NewRequest("GET",
		"http://"+consulClient.Host+":"+strconv.Itoa(consulClient.Port)+"/v1/health/service/"+serviceName, nil)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Discover Service Error!")
	} else if resp.StatusCode == 200 {
		var serviceList []struct {
			Service InstanceInfo `json:"Service"`
		}
		err = json.NewDecoder(resp.Body).Decode(&serviceList)
		err := resp.Body.Close()
		if err != nil {
			return nil
		}
		if err == nil {
			instances := make([]interface{}, len(serviceList))
			for i := 0; i < len(serviceList); i++ {
				instances[i] = serviceList[i].Service
			}
			return instances
		}
	}
	return nil
}
