package wrr

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"log"
	"sync"
	"time"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

var dLock sync.Mutex

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
			available:     true,
			addr:          connInfo.Address.Addr,
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
	available     bool
	addr          string
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
		dLock.Lock()
		available := conn.available
		dLock.Unlock()
		if available {
			sumWeight += conn.weight
			conn.currentWeight += conn.weight
			if maxCC == nil || conn.currentWeight > maxCC.currentWeight {
				maxCC = conn
			}
		}
	}
	maxCC.currentWeight -= sumWeight
	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			if status.Code(info.Err) == codes.ResourceExhausted { //触发限流
				maxCC.currentWeight -= maxCC.weight
				if maxCC.currentWeight < -sumWeight {
					maxCC.currentWeight = -sumWeight
				}
			} else if status.Code(info.Err) == codes.Unavailable { //触发熔断
				dLock.Lock() //加锁保证同一服务端熔断只启动一个协程进行健康检查
				if !maxCC.available {
					return
				}
				maxCC.available = false
				dLock.Unlock()
				go func() {
					// 连接到 gRPC 服务器
					conn, err := grpc.Dial(maxCC.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
					if err != nil {
						log.Fatalf("无法连接到 gRPC 服务器：%v", err)
					}
					defer conn.Close()

					// 创建 gRPC 健康检查客户端
					healthClient := grpc_health_v1.NewHealthClient(conn)

					// 定义健康检查请求
					req := &grpc_health_v1.HealthCheckRequest{
						Service: "", // 如果服务端未实现多个服务健康检查，可以为空字符串
					}

					for {
						// 发送健康检查请求
						resp, err := healthClient.Check(context.Background(), req)
						if err != nil {
							log.Fatalf("健康检查请求失败：%v", err)
						}

						// 检查服务状态
						if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
							maxCC.currentWeight = 0
							maxCC.available = true
							return
						} else {
							fmt.Println("服务端不可用")
						}

						// 等待一段时间后再次进行健康检查
						time.Sleep(5 * time.Second)
					}
				}()
			}

		},
	}, nil
}
