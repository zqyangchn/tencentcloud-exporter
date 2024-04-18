package instance

import (
	"fmt"

	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	rocketmq "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"tencentcloud-exporter/pkg/client"
	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/ROCKETMQ", NewRocketMQTcInstanceRepository)
}

var includeVip = "includeVip"
var includeVipTrue = "true"

type RocketMQTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *RocketMQTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *RocketMQTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeRocketMQClustersRequest()
	req.Filters = []*sdk.Filter{{
		Name:   &includeVip,
		Values: []*string{&includeVipTrue},
	}}
	req.ClusterIdList = []*string{&id}
	resp, err := repo.client.DescribeRocketMQClusters(req)
	if err != nil {
		return
	}
	if len(resp.Response.ClusterList) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.ClusterList[0]
	instance, err = NewRocketMQTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *RocketMQTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *RocketMQTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeRocketMQClustersRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit
	req.Filters = []*sdk.Filter{{
		Name:   &includeVip,
		Values: []*string{&includeVipTrue},
	}}
getMoreInstances:
	resp, err := repo.client.DescribeRocketMQClusters(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.ClusterList {
		ins, e := NewRocketMQTcInstance(*meta.Info.ClusterId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create rocketMQ instance fail", "id", *meta.Info.ClusterId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < uint64(total) {
		req.Offset = &offset
		goto getMoreInstances
	}

	return
}

// RocketMQNamespaces
type RocketMQTcInstanceRocketMQNameSpacesRepository interface {
	GetRocketMQNamespacesInfo(instanceId string) ([]*rocketmq.RocketMQNamespace, error)
}

type RocketMQTcInstanceRocketMQNameSpacesRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *RocketMQTcInstanceRocketMQNameSpacesRepositoryImpl) GetRocketMQNamespacesInfo(
	instanceId string,
) ([]*rocketmq.RocketMQNamespace, error) {
	req := sdk.NewDescribeRocketMQNamespacesRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Limit = &limit
	req.Offset = &offset
	req.ClusterId = tccommon.StringPtr(instanceId)

	namespaces := make([]*rocketmq.RocketMQNamespace, 0)

getMores:
	resp, err := repo.client.DescribeRocketMQNamespaces(req)
	if err != nil {
		return nil, err
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)

	}
	namespaces = append(namespaces, resp.Response.Namespaces...)

	offset += limit
	if offset < uint64(total) {
		req.Offset = &offset
		goto getMores
	}

	return namespaces, nil
}

func NewRocketMQTcInstanceRocketMQNameSpacesRepository(
	cred common.CredentialIface, c *config.TencentConfig, logger log.Logger,
) (RocketMQTcInstanceRocketMQNameSpacesRepository, error) {
	cli, err := client.NewRocketMQClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &RocketMQTcInstanceRocketMQNameSpacesRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// RocketMQTopics
type RocketMQTcInstanceRocketMQTopicsRepository interface {
	GetRocketMQTopicsInfo(instanceId string, namespaceId string,
	) ([]*rocketmq.RocketMQTopic, error)
}

type RocketMQTcInstanceRocketMQTopicsRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *RocketMQTcInstanceRocketMQTopicsRepositoryImpl) GetRocketMQTopicsInfo(
	instanceId string, namespaceId string,
) ([]*rocketmq.RocketMQTopic, error) {
	req := sdk.NewDescribeRocketMQTopicsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Limit = &limit
	req.Offset = &offset
	req.ClusterId = tccommon.StringPtr(instanceId)
	req.NamespaceId = tccommon.StringPtr(namespaceId)

	topics := make([]*rocketmq.RocketMQTopic, 0)

getMores:
	resp, err := repo.client.DescribeRocketMQTopics(req)
	if err != nil {
		return nil, err
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	topics = append(topics, resp.Response.Topics...)

	offset += limit
	if offset < uint64(total) {
		req.Offset = &offset
		goto getMores
	}

	return topics, nil
}

func NewRocketMQTcInstanceRocketMQTopicsRepository(
	cred common.CredentialIface, c *config.TencentConfig, logger log.Logger,
) (RocketMQTcInstanceRocketMQTopicsRepository, error) {
	cli, err := client.NewRocketMQClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &RocketMQTcInstanceRocketMQTopicsRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

// RocketMQTcInstanceRocketMQGroupsRepository RocketMQ Groups
type RocketMQTcInstanceRocketMQGroupsRepository interface {
	GetRocketMQGroupsInfo(instanceId string, namespaceId string, topic string,
	) ([]*rocketmq.RocketMQGroup, error)
}

type RocketMQTcInstanceRocketMQGroupsRepositoryImpl struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *RocketMQTcInstanceRocketMQGroupsRepositoryImpl) GetRocketMQGroupsInfo(
	instanceId string, namespaceId string, topic string,
) ([]*rocketmq.RocketMQGroup, error) {
	req := sdk.NewDescribeRocketMQGroupsRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total int64 = -1

	req.Limit = &limit
	req.Offset = &offset
	req.ClusterId = tccommon.StringPtr(instanceId)
	req.NamespaceId = tccommon.StringPtr(namespaceId)
	if topic != "" {
		req.FilterTopic = tccommon.StringPtr(topic)
	}

	groups := make([]*rocketmq.RocketMQGroup, 0)

getMores:
	resp, err := repo.client.DescribeRocketMQGroups(req)
	if err != nil {
		return nil, err
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	groups = append(groups, resp.Response.Groups...)

	offset += limit
	if offset < uint64(total) {
		req.Offset = &offset
		goto getMores
	}

	return groups, nil
}

func NewRocketMQTcInstanceRocketMQGroupsRepository(
	cred common.CredentialIface, c *config.TencentConfig, logger log.Logger,
) (RocketMQTcInstanceRocketMQGroupsRepository, error) {
	cli, err := client.NewRocketMQClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &RocketMQTcInstanceRocketMQGroupsRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}

func NewRocketMQTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewRocketMQClient(cred, c)
	if err != nil {
		return
	}
	repo = &RocketMQTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
