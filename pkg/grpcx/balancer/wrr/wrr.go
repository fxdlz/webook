package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	maxWeight := 0
	for conn, connInfo := range info.ReadySCs {
		md, _ := connInfo.Address.Metadata.(map[string]any)
		weightVal, _ := md["weight"]
		weight, _ := weightVal.(float64)
		maxWeight += int(weight)
		conns = append(conns, &weightConn{
			SubConn:       conn,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}
	return &Picker{
		conns: conns,
	}
}

type Picker struct {
	conns []*weightConn
	lock  sync.Mutex
}

type weightConn struct {
	balancer.SubConn
	weight        int
	currentWeight int
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	sumWeight := 0
	var maxCC *weightConn
	for _, conn := range p.conns {
		sumWeight += conn.weight
		conn.currentWeight += conn.weight
		if maxCC == nil || conn.currentWeight > maxCC.currentWeight {
			maxCC = conn
		}
	}
	maxCC.currentWeight -= sumWeight
	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			//根据响应是否异常调整权重
			if info.Err != nil {
				maxCC.currentWeight -= maxCC.weight
				if maxCC.currentWeight < -sumWeight {
					maxCC.currentWeight = -sumWeight
				}
			} else {
				maxCC.currentWeight += maxCC.weight
				if maxCC.currentWeight > sumWeight {
					maxCC.currentWeight = sumWeight
				}
			}
		},
	}, nil
}
