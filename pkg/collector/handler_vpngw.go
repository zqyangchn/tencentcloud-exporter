package collector

import (
	"github.com/go-kit/log"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/metric"
)

const (
	VpngwNamespace     = "QCE/VPNGW"
	VpngwInstanceidKey = "vpnGwId"
)

func init() {
	registerHandler(VpngwNamespace, defaultHandlerEnabled, NewVpngwHandler)
}

type VpngwHandler struct {
	baseProductHandler
}

func (h *VpngwHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *VpngwHandler) GetNamespace() string {
	return VpngwNamespace
}

func (h *VpngwHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func NewVpngwHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	handler = &VpngwHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: VpngwInstanceidKey,
			collector:       c,
			logger:          logger,
		},
	}
	return

}
