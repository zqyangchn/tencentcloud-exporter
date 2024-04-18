package collector

import (
	"github.com/go-kit/log"
	"strings"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/metric"
	"tencentcloud-exporter/pkg/util"
)

const (
	ClbNamespace     = "QCE/LB_PUBLIC"
	ClbInstanceidKey = "vip"
)

var (
	ClbPublicExcludeMetrics = []string{
		"RspMin",
	}
)

func init() {
	registerHandler(ClbNamespace, defaultHandlerEnabled, NewClbHandler)
}

type clbHandler struct {
	baseProductHandler
}

func (h *clbHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	if !util.IsStrInList(meta.SupportDimensions, ClbInstanceidKey) {
		meta.SupportDimensions = append(meta.SupportDimensions, ClbInstanceidKey)
	}

	return true
}

func (h *clbHandler) GetNamespace() string {
	return ClbNamespace
}

func (h *clbHandler) IsMetricValid(m *metric.TcmMetric) bool {
	if util.IsStrInList(ClbPublicExcludeMetrics, strings.ToLower(m.Meta.MetricName)) {
		return false
	}

	return true
}

func NewClbHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &clbHandler{
		baseProductHandler{
			monitorQueryKey: ClbInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return
}
