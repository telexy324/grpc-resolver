package ns

import (
	"context"
	"fmt"

	"google.golang.org/grpc/resolver"
)

// init 将定义好的 NS Builder 注册到 resolver 包中
func init() {
	resolver.Register(NewBuilder())
}

// NewBuilder 构造 nsResolverBuilder
func NewBuilder() resolver.Builder {
	return &nsResolverBuilder{}
}

// nsResolverBuilder 实现了 resolver.Builder 接口，用来构造定义好的 Resolver Bulder
type nsResolverBuilder struct{}

// URI 返回某个服务的统一资源描述符（URI），这个 URI 可以从 nsResolver 中查询实例列表
// URI 设计时可以遵循 RFC-3986(https://tools.ietf.org/html/rfc3986) 规范，
// 比如本例中 ns 格式为：ns://callerService:@calleeService
// 其中 ns 为协议名，callerService 为订阅方服务名（即主调方服务名），calleeService 为发布方服务名（即被调方服务名）
func URI(callerService, calleeService string) string {
	return fmt.Sprintf("ns://%s:@%s", callerService, calleeService)
}

// Build 实现了 resolver.Builder.Build 方法
func (*nsResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &nsResolver{
		target: target,
		cc:     cc,
		ctx:    ctx,
		cancel: cancel,
	}
	// 启动协程，响应指定 Name 服务实例变化
	go r.watcher()
	return r, nil
}

// Scheme 实现了 resolver.Builder.Scheme 方法
// Scheme 方法定义了 ns resolver 的协议名
func (*nsResolverBuilder) Scheme() string {
	return "ns"
}
