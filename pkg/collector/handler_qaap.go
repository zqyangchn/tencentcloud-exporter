package collector

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/instance"
	"tencentcloud-exporter/pkg/metric"
	"tencentcloud-exporter/pkg/util"
)

const (
	QaapNamespace     = "QCE/QAAP"
	QaapInstanceidKey = "channelId"
)

var (
	QaapDetail2GroupidMetricNames = []string{
		"GroupInFlow", "GroupOutFlow", "GroupInbandwidth", "GroupOutbandwidth",
	}
	QaapIpDetailMetricNames = []string{
		"IpConnum", "IpInbandwidth", "IpInpacket", "IpLatency", "IpOutbandwidth", "IpOutpacket", "IpPacketLoss",
	}
	QaapListenerStatMetricNames = []string{
		"ListenerConnum", "ListenerOutbandwidth", "ListenerInpacket", "ListenerOutpacket", "ListenerInbandwidth",
	}
	QaapListenerRsMetricNames = []string{
		"ListenerRsStatus",
	}
	QaapRuleRsMetricNames = []string{
		"RuleRsStatus",
	}
)

func init() {
	registerHandler(QaapNamespace, defaultHandlerEnabled, NewQaapHandler)
}

type QaapHandler struct {
	baseProductHandler
	commonQaapInstanceInfoRepo instance.CommonQaapTcInstanceRepository
	qaapInstanceInfoRepo       instance.QaapTcInstanceInfoRepository
}

func (h *QaapHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *QaapHandler) GetNamespace() string {
	return QaapNamespace
}

func (h *QaapHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func (h *QaapHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	if m.Conf.IsIncludeOnlyInstance() {
		return h.GetSeriesByOnly(m)
	}

	if m.Conf.IsIncludeAllInstance() {
		return h.GetSeriesByAll(m)
	}

	if m.Conf.IsCustomQueryDimensions() {
		return h.GetSeriesByCustom(m)
	}
	return nil, fmt.Errorf("must config all_instances or only_include_instances or custom_query_dimensions")
}

func (h *QaapHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, insId := range m.Conf.OnlyIncludeInstances {
		ins, err := h.collector.InstanceRepo.Get(insId)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "id", insId)
			continue
		}
		sl, err := h.getSeriesByMetricType(m, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *QaapHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	insList, err := h.collector.InstanceRepo.ListByFilters(m.Conf.InstanceFilters)
	if err != nil {
		return nil, err
	}
	for _, ins := range insList {
		if len(m.Conf.ExcludeInstances) != 0 && util.IsStrInList(m.Conf.ExcludeInstances, ins.GetInstanceId()) {
			continue
		}
		sl, err := h.getSeriesByMetricType(m, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *QaapHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var slist []*metric.TcmSeries
	for _, ql := range m.Conf.CustomQueryDimensions {
		v, ok := ql[h.monitorQueryKey]
		if !ok {
			level.Error(h.logger).Log(
				"msg", fmt.Sprintf("not found %s in queryDimensions", h.monitorQueryKey),
				"ql", fmt.Sprintf("%v", ql))
			continue
		}
		ins, err := h.collector.InstanceRepo.Get(v)
		if err != nil {
			level.Error(h.logger).Log("msg", "Instance not found", "err", err, "id", v)
			continue
		}

		sl, err := h.getSeriesByMetricType(m, ins)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail",
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId())
			continue
		}
		slist = append(slist, sl...)
	}
	return slist, nil
}

func (h *QaapHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(QaapDetail2GroupidMetricNames, m.Meta.MetricName) {
		return h.getQaapDetail2GroupidSeries(m, ins)
	} else if util.IsStrInList(QaapIpDetailMetricNames, m.Meta.MetricName) {
		return h.getQaapIpDetailSeries(m, ins)
	} else if util.IsStrInList(QaapListenerStatMetricNames, m.Meta.MetricName) {
		return h.getQaapListenerStatSeries(m, ins)
	} else if util.IsStrInList(QaapListenerRsMetricNames, m.Meta.MetricName) {
		return h.getQaapListenerRsSeries(m, ins)
	} else if util.IsStrInList(QaapRuleRsMetricNames, m.Meta.MetricName) {
		return h.getRuleRsSeries(m, ins)
	} else {
		return h.getInstanceSeries(m, ins)
	}
}

func (h *QaapHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {

	var series []*metric.TcmSeries
	ql := map[string]string{
		h.monitorQueryKey: ins.GetMonitorQueryKey(),
	}
	s, err := metric.NewTcmSeries(m, ql, ins)
	if err != nil {
		return nil, err
	}
	series = append(series, s)

	return series, nil
}

func (h *QaapHandler) getQaapListenerRsSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	var tcpSeries, udpSeries []*metric.TcmSeries
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		tcpSeries, _ = h.getTcpSeries(m, ins)
		series = append(series, tcpSeries...)
	}()
	go func() {
		defer wg.Done()
		udpSeries, _ = h.getUdpSeries(m, ins)
		series = append(series, udpSeries...)
	}()
	wg.Wait()

	// tcpListenersInfos, err := h.qaapInstanceInfoRepo.GetTCPListenersInfo(ins.GetInstanceId())
	// if err != nil {
	// 	return nil, err
	// }
	// for _, tcpListenersInfo := range tcpListenersInfos.Response.ListenerSet {
	// 	for _, realServerSet := range tcpListenersInfo.RealServerSet {
	// 		ql := map[string]string{
	// 			h.monitorQueryKey:  ins.GetMonitorQueryKey(),
	// 			"listenerId":       *tcpListenersInfo.ListenerId,
	// 			"originServerInfo": *realServerSet.RealServerIP,
	// 			"protocol":         *tcpListenersInfo.Protocol,
	// 			"listenerName":     *tcpListenersInfo.ListenerName,
	// 		}
	// 		s, err := metric.NewTcmSeries(m, ql, ins)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		series = append(series, s)
	// 	}
	// }
	// udpListenersInfos, err := h.qaapInstanceInfoRepo.GetUDPListenersInfo(ins.GetInstanceId())
	// if err != nil {
	// 	return nil, err
	// }
	// for _, udpListenersInfo := range udpListenersInfos.Response.ListenerSet {
	// 	for _, realServerSet := range udpListenersInfo.RealServerSet {
	// 		ql := map[string]string{
	// 			h.monitorQueryKey:  ins.GetMonitorQueryKey(),
	// 			"listenerId":       *udpListenersInfo.ListenerId,
	// 			"originServerInfo": *realServerSet.RealServerIP,
	// 			"protocol":         *udpListenersInfo.Protocol,
	// 			"listenerName":     *udpListenersInfo.ListenerName,
	// 		}
	// 		s, err := metric.NewTcmSeries(m, ql, ins)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		series = append(series, s)
	// 	}
	// }
	return series, nil
}
func (h *QaapHandler) getTcpSeries(m *metric.TcmMetric, ins instance.TcInstance) (series []*metric.TcmSeries, err error) {
	tcpListenersInfos, err := h.qaapInstanceInfoRepo.GetTCPListenersInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, tcpListenersInfo := range tcpListenersInfos.Response.ListenerSet {
		for _, realServerSet := range tcpListenersInfo.RealServerSet {
			ql := map[string]string{
				h.monitorQueryKey:  ins.GetMonitorQueryKey(),
				"listenerId":       *tcpListenersInfo.ListenerId,
				"originServerInfo": *realServerSet.RealServerIP,
				"protocol":         *tcpListenersInfo.Protocol,
				"listenerName":     *tcpListenersInfo.ListenerName,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				return nil, err
			}
			series = append(series, s)
		}
	}
	return series, nil
}
func (h *QaapHandler) getUdpSeries(m *metric.TcmMetric, ins instance.TcInstance) (series []*metric.TcmSeries, err error) {
	udpListenersInfos, err := h.qaapInstanceInfoRepo.GetUDPListenersInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, udpListenersInfo := range udpListenersInfos.Response.ListenerSet {
		for _, realServerSet := range udpListenersInfo.RealServerSet {
			ql := map[string]string{
				h.monitorQueryKey:  ins.GetMonitorQueryKey(),
				"listenerId":       *udpListenersInfo.ListenerId,
				"originServerInfo": *realServerSet.RealServerIP,
				"protocol":         *udpListenersInfo.Protocol,
				"listenerName":     *udpListenersInfo.ListenerName,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				return nil, err
			}
			series = append(series, s)
		}
	}
	return series, nil
}

func (h *QaapHandler) getQaapDetail2GroupidSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	proxyGroupLists, err := h.qaapInstanceInfoRepo.GetProxyGroupList(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, proxyGroupList := range proxyGroupLists.Response.ProxyGroupList {
		ql := map[string]string{
			"GroupId": *proxyGroupList.GroupId,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}
	return series, nil
}

func (h *QaapHandler) getQaapIpDetailSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	noneBgpIpLists, err := h.commonQaapInstanceInfoRepo.GetCommonQaapNoneBgpIpList(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, instanceSet := range noneBgpIpLists.Response.InstanceSet {
		ql := map[string]string{
			"ip":      instanceSet.IP,
			"proxyid": instanceSet.ProxyId,
			"groupid": instanceSet.GroupId,
			"isp":     strings.ToLower(instanceSet.Isp),
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}
	return series, nil
}

func (h *QaapHandler) getQaapListenerStatSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	ProxyInstances, err := h.commonQaapInstanceInfoRepo.GetCommonQaapProxyInstances(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, proxySet := range ProxyInstances.Response.ProxySet {
		for _, l4Listener := range proxySet.L4ListenerSet {
			ql := map[string]string{
				"instanceid": proxySet.ProxyId,
				"listenerid": l4Listener.ListenerId,
				"protocol":   l4Listener.Protocol,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				return nil, err
			}
			series = append(series, s)
		}
		for _, l7Listener := range proxySet.L7ListenerSet {
			ql := map[string]string{
				"instanceid": proxySet.ProxyId,
				"listenerid": l7Listener.ListenerId,
				"protocol":   l7Listener.ForwardProtocol,
			}
			s, err := metric.NewTcmSeries(m, ql, ins)
			if err != nil {
				return nil, err
			}
			series = append(series, s)
		}

	}
	return series, nil
}
func (h *QaapHandler) getRuleRsSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	ProxyInstances, err := h.commonQaapInstanceInfoRepo.GetCommonQaapProxyInstances(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, proxySet := range ProxyInstances.Response.ProxySet {
		for _, l7Listener := range proxySet.L7ListenerSet {
			for _, rule := range l7Listener.RuleSet {
				for _, rs := range rule.RsSet {
					ql := map[string]string{
						"instanceid": proxySet.ProxyId,
						"listenerid": l7Listener.ListenerId,
						"ruleid":     rule.RuleId,
						"rs_ip":      rs.RsInfo,
					}
					s, err := metric.NewTcmSeries(m, ql, ins)
					if err != nil {
						return nil, err
					}
					series = append(series, s)
				}

			}

		}

	}
	return series, nil
}

func NewQaapHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	qaapInstanceInfo, err := instance.NewQaapTcInstanceInfoRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	reloadInterval := time.Duration(c.ProductConf.ReloadIntervalMinutes * int64(time.Minute))
	qaapInstanceInfoCache := instance.NewTcGaapInstanceInfosCache(qaapInstanceInfo, reloadInterval, logger)

	commonQaapInstanceInfoRepo, err := instance.NewCommonQaapTcInstanceRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	commonQaapInstanceInfoCache := instance.NewTcCommonGaapInstanceInfosCache(commonQaapInstanceInfoRepo, reloadInterval, logger)

	handler = &QaapHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: QaapInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		commonQaapInstanceInfoRepo: commonQaapInstanceInfoCache,
		qaapInstanceInfoRepo:       qaapInstanceInfoCache,
	}
	return

}
