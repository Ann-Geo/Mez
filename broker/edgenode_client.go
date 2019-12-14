// Edge node is a client of edge server registering with edge server

package broker

import (
	"context"
	"errors"
	"time"

	"github.com/Ann-Geo/Mez/api/edgeserver"
)

type EdgeNodeClient struct {
	Auth Authentication
}

func NewEdgeNodeClient(login, password string) *EdgeNodeClient {
	return &EdgeNodeClient{Auth: Authentication{login: login, password: password}}
}

func (cc *EdgeNodeClient) Register(ipaddr string, camid string, client edgeserver.PubSubClient) error {

	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
	defer cancel()

	nodepars := &edgeserver.NodeInfo{Ipaddr: ipaddr, Camid: camid}
	status, _ := client.Register(ctx, nodepars)
	if status.GetStatus() == false {
		return errors.New("Could not register with edge server")
	}
	return nil

}
