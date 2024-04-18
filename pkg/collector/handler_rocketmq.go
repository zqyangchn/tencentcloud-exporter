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

/*
	https://cloud.tencent.com/document/product/248/80467
*/

const (
	RocketMQNamespace     = "QCE/ROCKETMQ"
	RocketMQInstanceidKey = "tenant"
)

func init() {
	registerHandler(RocketMQNamespace, defaultHandlerEnabled, NewRocketMQHandler)
}

type rocketMQHandler struct {
	baseProductHandler

	namespaceRepo instance.RocketMQTcInstanceRocketMQNameSpacesRepository
	topicRepo     instance.RocketMQTcInstanceRocketMQTopicsRepository
	groupRepo     instance.RocketMQTcInstanceRocketMQGroupsRepository
}

func (h *rocketMQHandler) GetNamespace() string {
	return RocketMQNamespace
}

func (h *rocketMQHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *rocketMQHandler) IsMetricValid(m *metric.TcmMetric) bool {
	return true
}

func (h *rocketMQHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *rocketMQHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var sList []*metric.TcmSeries
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
		sList = append(sList, sl...)
	}
	return sList, nil
}

func (h *rocketMQHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var sList []*metric.TcmSeries
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
				"metric", m.Meta.MetricName, "instance", ins.GetInstanceId(),
			)
			continue
		}
		sList = append(sList, sl...)
	}
	return sList, nil
}

func (h *rocketMQHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var sList []*metric.TcmSeries
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
		sList = append(sList, sl...)
	}
	return sList, nil
}

func (h *rocketMQHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}

	if util.IsStrInList(dimensions, "namespace") &&
		util.IsStrInList(dimensions, "topic") &&
		util.IsStrInList(dimensions, "group") {
		return h.getNamespaceTopicGroupsSeries(m, ins)
	}

	if util.IsStrInList(dimensions, "namespace") &&
		util.IsStrInList(dimensions, "topic") {
		return h.getNamespaceTopicSeries(m, ins)
	}

	if util.IsStrInList(dimensions, "namespace") &&
		util.IsStrInList(dimensions, "group") {
		return h.getNamespaceGroupsSeries(m, ins)
	}

	return h.getInstanceSeries(m, ins)

}

func (h *rocketMQHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
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

func (h *rocketMQHandler) getNamespaceTopicSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	namespaces, err := h.namespaceRepo.GetRocketMQNamespacesInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces {
		topics, err := h.topicRepo.GetRocketMQTopicsInfo(ins.GetInstanceId(), *namespace.NamespaceId)
		if err != nil {
			return nil, err
		}
		for _, topic := range topics {
			ql := map[string]string{
				"tenant":    ins.GetMonitorQueryKey(),
				"namespace": *namespace.NamespaceId,
				"topic":     *topic.Name,
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

func (h *rocketMQHandler) getNamespaceGroupsSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	namespaces, err := h.namespaceRepo.GetRocketMQNamespacesInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces {
		groups, err := h.groupRepo.GetRocketMQGroupsInfo(ins.GetInstanceId(), *namespace.NamespaceId, "")
		if err != nil {
			return nil, err
		}
		for _, group := range groups {
			ql := map[string]string{
				"tenant":    ins.GetMonitorQueryKey(),
				"namespace": *namespace.NamespaceId,
				"group":     *group.Name,
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

func (h *rocketMQHandler) getNamespaceTopicGroupsSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	namespaces, err := h.namespaceRepo.GetRocketMQNamespacesInfo(ins.GetInstanceId())
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces {
		topics, err := h.topicRepo.GetRocketMQTopicsInfo(ins.GetInstanceId(), *namespace.NamespaceId)
		if err != nil {
			return nil, err
		}
		for _, topic := range topics {
			groups, err := h.groupRepo.GetRocketMQGroupsInfo(
				ins.GetInstanceId(), *namespace.NamespaceId, *topic.Name,
			)
			if err != nil {
				return nil, err
			}
			for _, group := range groups {
				ql := map[string]string{
					"tenant":    ins.GetMonitorQueryKey(),
					"namespace": *namespace.NamespaceId,
					"topic":     *topic.Name,
					"group":     *group.Name,
				}
				s, err := metric.NewTcmSeries(m, ql, ins)
				if err != nil {
					return nil, err
				}
				series = append(series, s)
			}
		}
	}
	return series, nil
}

func NewRocketMQHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	namespaceRepo, err := instance.NewRocketMQTcInstanceRocketMQNameSpacesRepository(cred, c.Conf, logger)
	if err != nil {

		return nil, err
	}
	reloadInterval := time.Duration(c.ProductConf.ReloadIntervalMinutes * int64(time.Minute))
	namespaceRepoCache := instance.NewTcRocketMQInstanceNamespaceCache(namespaceRepo, reloadInterval, logger)

	topicRepo, err := instance.NewRocketMQTcInstanceRocketMQTopicsRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	topicRepoCache := instance.NewTcRocketMQInstanceTopicsCache(topicRepo, reloadInterval, logger)

	groupRepo, err := instance.NewRocketMQTcInstanceRocketMQGroupsRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	groupRepoCache := instance.NewTcRocketMQInstanceGroupsCache(groupRepo, reloadInterval, logger)

	handler = &rocketMQHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: RocketMQInstanceidKey,
			collector:       c,
			logger:          logger,
		},
		namespaceRepo: namespaceRepoCache,
		topicRepo:     topicRepoCache,
		groupRepo:     groupRepoCache,
	}
	return

}
