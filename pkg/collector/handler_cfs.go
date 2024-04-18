package collector

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/instance"
	"tencentcloud-exporter/pkg/metric"
	"tencentcloud-exporter/pkg/util"
)

const (
	CfsNamespace     = "QCE/CFS"
	CfsInstanceIdKey = "FileSystemId"
)

func init() {
	registerHandler(CfsNamespace, defaultHandlerEnabled, NewCfsHandler)
}

type CfsHandler struct {
	baseProductHandler
	cfsSnapshotsRepo instance.CfsSnapshotsRepository
}

func (h *CfsHandler) IsMetricMetaValid(meta *metric.TcmMeta) bool {
	return true
}

func (h *CfsHandler) GetNamespace() string {
	return CfsNamespace
}

func (h *CfsHandler) IsMetricValid(m *metric.TcmMetric) bool {
	return true
}

func (h *CfsHandler) GetSeries(m *metric.TcmMetric) (slist []*metric.TcmSeries, err error) {
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
func (h *CfsHandler) GetSeriesByOnly(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *CfsHandler) GetSeriesByAll(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *CfsHandler) GetSeriesByCustom(m *metric.TcmMetric) ([]*metric.TcmSeries, error) {
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

func (h *CfsHandler) getSeriesByMetricType(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var dimensions []string
	for _, v := range m.Meta.SupportDimensions {
		dimensions = append(dimensions, v)
	}

	if util.IsStrInList(dimensions, "appid") {
		return h.getCfsSnapshotsSeries(m, ins)
	} else {
		return h.getInstanceSeries(m, ins)
	}
}

func (h *CfsHandler) getInstanceSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
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
func (h *CfsHandler) getCfsSnapshotsSeries(m *metric.TcmMetric, ins instance.TcInstance) ([]*metric.TcmSeries, error) {
	var series []*metric.TcmSeries
	cfsSnapshots, err := h.cfsSnapshotsRepo.GetCfsSnapshotsInfo("")
	if err != nil {
		return nil, err
	}
	for _, cfsSnapshot := range cfsSnapshots.Response.Snapshots {
		ql := map[string]string{
			"appid": strconv.FormatUint(*cfsSnapshot.AppId, 10),
		}
		s, err := metric.NewTcmSeries(m, ql, ins)
		if err != nil {
			return nil, err
		}
		series = append(series, s)
	}
	return series, nil
}

func NewCfsHandler(cred common.CredentialIface, c *TcProductCollector, logger log.Logger) (handler ProductHandler, err error) {
	cfsSnapshotsRepo, err := instance.NewCfsSnapshotsRepositoryRepository(cred, c.Conf, logger)
	if err != nil {
		return nil, err
	}
	reloadInterval := time.Duration(c.ProductConf.ReloadIntervalMinutes * int64(time.Minute))
	cfsSnapshotsRepoCache := instance.NewCfsSnapshotsRepositoryRepositoryCache(cfsSnapshotsRepo, reloadInterval, logger)
	handler = &CfsHandler{
		baseProductHandler: baseProductHandler{
			monitorQueryKey: CfsInstanceIdKey,
			collector:       c,
			logger:          logger,
		},
		cfsSnapshotsRepo: cfsSnapshotsRepoCache,
	}
	return

}
