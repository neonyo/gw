package router

import "github.com/neonyo/gw/pkg/ratelimiter"

type Endpoint struct {
	Addr               string           //服务端地址
	Name               string           //服务端名称
	BlackIps           []string         //黑名单
	RateLimiterEnabled bool             //是否开启限流
	ReteLimiterRule    ratelimiter.Rule //限流规则
	RateLimiterMsg     string           //限流后返回的json数据
	Router             []*Router        //具体路由
}

type Router struct {
	ID                     string           //id编码，需要保持唯一
	Name                   string           //路由名称
	ReqMethod              string           //请求方法
	ReqPath                string           //请求地址
	ProxyPath              string           //代理地址
	IsAuth                 int              //是否需要鉴权 0:不鉴权 1:验证token&鉴权 2:验证token不鉴权
	RateLimiterEnabled     bool             //是否开启限流
	ReteLimiterRule        ratelimiter.Rule //限流规则
	RateLimiterMsg         string           //限流后返回的json数据
	CircuitBreakerEnabled  bool             //是否开启自动熔断
	CircuitBreakerRequest  int              //最大并发数
	CircuitBreakerPercent  int              //请求出错比
	CircuitBreakerTimeout  int              // 超时时间定义
	CircuitVolumeThreshold int              // 跳闸的最小请求数 只有滑动窗口时间内的请求数量超过该值，断路器才会执行对应的判断逻辑
	CircuitSleepWindow     int              // 跳闸之后可以重试的时间
	CircuitBreakerForce    bool             //开启手动熔断
	CircuitBreakerMsg      string           //手动熔断返回json数据
}
