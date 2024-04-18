package collector

import (
	"fmt"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/instance"
	"tencentcloud-exporter/pkg/metric"
	"tencentcloud-exporter/pkg/util"
)

const (
	ZookeeperNamespace     = "TSE/ZOOKEEPER"
	ZookeeperInstanceidKey = "InstanceId"
)

func init() {
	registerHandler(ZookeeperNamespace, defaultHandlerEnabled, NewZookeeperHandler)
}

type ZookeeperHandler struct {
	baseProductHandler
	podRepo       instance.ZookeeperTcInstancePodRepository
	interfaceRepo instance.ZookeeperTcInstanceInterfaceRepository
}

func (h *ZookeeperHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *ZookeeperHandler) GetNamespace() string {
	return ZookeeperNamespace
}

func (h *ZookeeperHandler) IsMetricValid(m *metric.TcmMetric) bool {
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
func (h *ZookeeperHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ZookeeperHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ZookeeperHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ZookeeperHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *ZookeeperHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	if util.IsStrInList(dimensions, "Interface") {
		return h.getInterfaceSeries(m, ins)
	} else {
		return h.getPodSeries(m, ins)
	}
}

func (h *ZookeeperHandler) getPodSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	podInfoResp, err := h.podRepo.GetZookeeperPodInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, podInfo := range podInfoResp.Response.Replicas {

		ql := map[string]string{
			"InstanceId": ins.GetMonitorQueryKey(),
			"PodName":    *podInfo.Name,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)

	}

	return series, nil
}

func (h *ZookeeperHandler) getInterfaceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	interfaceInfoResp, err := h.interfaceRepo.GetZookeeperInterfaceInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	podInfoResp, err := h.podRepo.GetZookeeperPodInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, podInfo := range podInfoResp.Response.Replicas {
		for _, interfaceInfo := range interfaceInfoResp.Response.Content {
			ql := map[string]string{
				"InstanceId": ins.GetMonitorQueryKey(),
				"PodName":    *podInfo.Name,
				"Interface":  *interfaceInfo.Interface,
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

func NewZookeeperHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	podRepo, err := instance.NewZookeeperTcInstancePodRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	reloadInterval := time.Duration(c.ProductConf.ReloadIntervalMinutes * int64(time.Minute))
	podRepoCache := instance.NewTcZookeeperInstancePodCache(podRepo, reloadInterval, logger)

	interfaceRepo, err := instance.NewZookeeperTcInstanceInterfaceRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	interfaceRepoCache := instance.NewTcZookeeperInstanceInterfaceCache(interfaceRepo, reloadInterval, logger)

	handler = &ZookeeperHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: ZookeeperInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		podRepo:       podRepoCache,
		interfaceRepo: interfaceRepoCache,
	}
	return
}
