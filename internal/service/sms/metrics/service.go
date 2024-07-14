package metrics

import (
	"context"
	"github.com/dadaxiaoxiao/user/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service,
	namespace string,
	subsystem string,
	name string,
	instanceId string,
) sms.Service {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		ConstLabels: map[string]string{
			"instance_id": instanceId,
		},
		Help: "统计 SMS 服务的性能数据",
	}, []string{"tpl"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (p PrometheusDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(tplId).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, tplId, args, numbers...)
}

