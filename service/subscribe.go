package service

import (
	"github.com/coreos/go-etcd/etcd"
	"path"
)

func Subscribe(service string) <-chan *etcd.Response {
	stop := make(chan bool)
	responses := make(chan *etcd.Response)
	go func() {
		_, err := client.Watch("/services/"+service, 0, true, responses, stop)
		if err != nil {
			logger.Println("fail to watch", service, err)
			close(responses)
			close(stop)
		}
	}()
	return responses
}

func SubscribeDown(service string) chan string {
	expirations := make(chan string)
	go func() {
		responses := Subscribe(service)
		for response := range responses {
			if response.Action == "expire" {
				expirations <- path.Base(response.Node.Key)
			}
		}
	}()
	return expirations
}

func SubscribeNew(service string) chan *Host {
	hosts := make(chan *Host)
	go func() {
		responses := Subscribe(service)
		for response := range responses {
			if response.Node.Dir && response.Action == "create" {
				res, err := client.Get(response.Node.Key, false, true)
				if err != nil {
					logger.Println("Fail to get", response.Node.Key, ":", err)
					continue
				}
				hosts <- buildHostFromNode(res.Node)
			}
		}
	}()
	return hosts
}
