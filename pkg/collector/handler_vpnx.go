package collector

import (
	"github.com/go-kit/log"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/metric"
)

const (
	VpnxNamespace     = "QCE/VPNX"
	VpnxInstanceidKey = "vpnConnId"
)

func init() {
	registerHandler(VpnxNamespace, defaultHandlerEnabled, NewVpnxHandler)
}

type VpnxHandler struct {
	baseProductHandler
}

func (h *VpnxHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *VpnxHandler) GetNamespace() string {
	return VpnxNamespace
}

func (h *VpnxHandler) IsMetricValid(m *metric.TcmMetric) bool {
	_, ok := excludeMetricName[m.Meta.MetricName]
	if ok {
		return false
	}
	p, err := m.Meta.GetPeriod(m.Conf.StatPeriodSeconds)
	if err != nil {
		return false
	}
	if p != m.Conf.StatPeriodSeconds {
		return false
	}
	return true
}

func NewVpnxHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &VpnxHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: VpnxInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
