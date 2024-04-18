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
	DTSNamespace     = "QCE/DTS"
	DTSInstanceIdKey = "SubscribeId"
)

func init() {
	registerHandler(DTSNamespace, defaultHandlerEnabled, NewDTSHandler)
}

type dtsHandler struct {
	baseProductHandler
	replicationRepo  instance.DtsTcInstanceReplicationsRepository
	migrateInfosRepo instance.DtsTcInstanceMigrateInfosRepository
}

func (h *dtsHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *dtsHandler) GetNamespace() string {
	return DTSNamespace
}

func (h *dtsHandler) IsMetricValid(m *metric.TcmMetric) bool {
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

func (h *dtsHandler) GetSeries(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *dtsHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *dtsHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}
	var slist []*metric.TcmSeries
	if util.IsStrInList(dimensions, "replicationjobid") || util.IsStrInList(dimensions, "migratejobid") {
		sl, err := h.getSeriesByMetricType(m, nil)
		if err != nil {
			level.Error(h.logger).Log("msg", "Create metric series fail", "metric", m.Meta.MetricName)
		}
		slist = append(slist, sl...)
	} else {
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
	}

	return slist, nil
}

func (h *dtsHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *dtsHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}

	if util.IsStrInList(dimensions, "replicationjobid") {
		return h.getReplicationSeries(m, ins)
	} else if util.IsStrInList(dimensions, "migratejobid") {
		return h.getMigrateInfoSeries(m, ins)
	} else {
		return h.getInstanceSeries(m, ins)
	}
}

func (h *dtsHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	subscribeName, err := ins.GetFieldValueByName("SubscribeName")
	ql := map[string]string{
		h.monitorQueryKey: ins.GetMonitorQueryKey(),
		"subscribe_name":  subscribeName,
	}
	s, err := metric.NewTcmSeries(m, ql, ins)
	if err != nil {
		return nil, err
	}
	series = append(series, s)

	return series, nil
}

func (h *dtsHandler) getReplicationSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	replications, err := h.replicationRepo.GetReplicationsInfo("")
	if err != nil {
		return nil, err
	}
	for _, replication := range replications.Response.JobList {
		ql := map[string]string{
			"replicationjobid":    *replication.JobId,
			"replicationjob_name": *replication.JobName,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}
	return series, nil
}
func (h *dtsHandler) getMigrateInfoSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	migrateInfos, err := h.migrateInfosRepo.GetMigrateInfos("")
	if err != nil {
		return nil, err
	}
	for _, migrateInfo := range migrateInfos.Response.JobList {
		ql := map[string]string{
			"migratejob_id":   *migrateInfo.JobId,
			"migratejob_name": *migrateInfo.JobName,
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}

	return series, nil
}

func NewDTSHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	migrateInfosRepo, err := instance.NewDtsTcInstanceMigrateInfosRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	reloadInterval := time.Duration(c.ProductConf.ReloadIntervalMinutes * int64(time.Minute))
	migrateInfosRepoCache := instance.NewTcDtsInstanceMigrateInfosCache(migrateInfosRepo, reloadInterval, logger)

	replicationRepo, err := instance.NewDtsTcInstanceReplicationsRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	replicationRepoCache := instance.NewTcDtsInstanceReplicationsInfosCache(replicationRepo, reloadInterval, logger)

	handler = &dtsHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: DTSInstanceIdKey,
			collector:       c,
			logger:          logger,
		},
		migrateInfosRepo: migrateInfosRepoCache,
		replicationRepo:  replicationRepoCache,
	}
	return

}
