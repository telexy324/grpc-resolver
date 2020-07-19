package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mypkg/internal/logz"        // 私有日志包，基于 uber 开源的 zap 实现
	sdk "mypkg/internal/soa-sdk" // 私有 ns sdk 包，封装了内部 soa 平台进行服务发现的 sdk

	_ "google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

const (
	// syncNSInterval 定义了从 NS 服务同步实例列表的周期
	syncNSInterval = 1 * time.Second
)

// nsResolver 实现了 resolver.Resolver 接口
type nsResolver struct {
	target    resolver.Target
	cc        resolver.ClientConn
	ctx       context.Context
	cancel    context.CancelFunc
	...
}

// watcher 轮询并更新指定 CalleeService 服务的实例变化
func (r *nsResolver) watcher() {
	r.updateCC()
	ticker := time.NewTicker(syncNSInterval)
	for {
		select {
		// 当* nsResolver Close 时退出监听
		case <-r.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			// 调用* nsResolver.updagteCC() 方法，更新实例地址
			r.updateCC()
		}
	}
}

// updateCC 更新 resolver.Resolver.ClientConn 配置
func (r *nsResolver) updateCC() {
	// 从 NS 服务获取指定 target 的实例列表
	instances, err := r.getInstances(r.target)
	// 如果获取实例列表失败，或者实例列表为空，则不更新 resolver 中实例列表
	if err != nil || len(instances.CalleeIns) == 0 {
		logz.Warn("[mis] error retrieving instances from Mis", logz.Any("target", r.target), logz.Error(err))
		return
	}
	...

	// 组装实例列表 []resolver.Address
	// resolver.Address 结构体表示 grpc server 端实例地址
	var newAddrs []resolver.Address
	for k := range instances.CalleeIns {
		newAddrs = append(newAddrs, instances.CalleeIns)
	}
	...

	// 更新实例列表
	// grpc 底层 LB 组件对每个服务端实例创建一个 subConnection。并根据设定的 LB 策略，选择合适的 subConnection 处理某次 RPC 请求。
	// 此处代码比较复杂，后续在 LB 相关原理文章中再做概述
	r.cc.UpdateState(resolver.State{Addresses: newAddrs})
}

// ResolveNow 实现了 resolver.Resolver.ResolveNow 方法
func (*nsResolver) ResolveNow(o resolver.ResolveNowOption) {}

// Close 实现了 resolver.Resolver.Close 方法
func (r *nsResolver) Close() {
	r.cancel()
}

// instances 包含调用方服务名、被调方服务名、被调方实例列表等数据
type instances struct {
	callerService string
	calleeService string
	calleeIns     []string
}

// getInstances 获取指定服务所有可用的实例列表
func (r *nsResolver) getInstances(target resolver.Target) (s *instances, e error) {
	auths := strings.Split(target.Authority, "@")
	// auths[0] 为 callerService 名，target.Endpoint 为 calleeService 名
	// 通过自定义 sdk 从内部 NameServer 查询指定 calleeService 对应的实例列表
	ins, e := sdk.GetInstances(auths[0], target.Endpoint)
	if e != nil {
		return nil, e
	}
	return &instances{
		callerService: auths[0],
		calleeService: target.Endpoint,
		calleeIns:     ins.Instances,
	}, nil
}
