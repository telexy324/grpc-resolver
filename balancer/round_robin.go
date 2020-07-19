package balancer

import (
	"github.com/telexy324/grpc-resolver/common"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	"math/rand"
	"sync"
)

const RoundRobin = "round_robin"

// newRoundRobinBuilder creates a new roundrobin balancer builder.
func newRoundRobinBuilder() balancer.Builder {
	return base.NewBalancerBuilder(RoundRobin, &roundRobinPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newRoundRobinBuilder())
}

type roundRobinPickerBuilder struct{}

//func (*roundRobinPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
//	grpclog.Infof("roundrobinPicker: newPicker called with readySCs: %v", readySCs)
//
//	if len(readySCs) == 0 {
//		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
//	}
//	var scs []balancer.SubConn
//	for addr, sc := range readySCs {
//		weight := common.GetWeight(addr)
//		for i := 0; i < weight; i++ {
//			scs = append(scs, sc)
//		}
//	}
//
//	return &roundRobinPicker{
//		subConns: scs,
//		next:     rand.Intn(len(scs)),
//	}
//}

func (roundRobinPickerBuilder) Build(pickerBuildInfo base.PickerBuildInfo) balancer.Picker {
	grpclog.Infof("roundrobinPicker: newPicker called with readySCs: %v", pickerBuildInfo.ReadySCs)

	if len(pickerBuildInfo.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	var scs []balancer.SubConn
	for sc, scInfo := range pickerBuildInfo.ReadySCs {
		weight := common.GetWeight(scInfo.Address)
		for i := 0; i < weight; i++ {
			scs = append(scs, sc)
		}
	}

	return &roundRobinPicker{
		subConns: scs,
		mu:       sync.Mutex{},
		next:     rand.Intn(len(scs)),
	}
}

type roundRobinPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
	next     int
}

func (p *roundRobinPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	sc := p.subConns[p.next]
	p.next = (p.next + 1) % len(p.subConns)
	p.mu.Unlock()
	return balancer.PickResult{
		SubConn: sc,
	}, nil
}
