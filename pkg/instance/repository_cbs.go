package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"tencentcloud-exporter/pkg/client"
	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/BLOCK_STORAGE", NewCbsTcInstanceRepository)
}

type CbsTcInstanceRepository struct {
	credential common.CredentialIface
	c          *config.TencentConfig
	client     *sdk.Client
	logger     log.Logger
}

func (repo *CbsTcInstanceRepository) GetInstanceKey() string {
	return "DiskId"
}

func (repo *CbsTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDisksRequest()
	req.DiskIds = []*string{&id}
	resp, err := repo.client.DescribeDisks(req)
	if err != nil {
		return
	}
	if len(resp.Response.DiskSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.DiskSet[0]
	instance, err = NewCbsTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CbsTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CbsTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDisksRequest()
	var offset uint64 = 0
	var limit uint64 = 100
	var total uint64 = 0

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDisks(req)
	if err != nil {
		return
	}
	if total == 0 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.DiskSet {
		ins, e := NewCbsTcInstance(*meta.DiskId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cbs instance fail", "id", *meta.DiskId)
			continue
		}
		instances = append(instances, ins)
	}
	offset += limit
	if offset < total {
		req.Offset = &offset
		goto getMoreInstances
	}
	return
}

func NewCbsTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCbsClient(cred, c)
	if err != nil {
		return
	}
	repo = &CbsTcInstanceRepository{
		credential: cred,
		c:          c,
		client:     cli,
		logger:     logger,
	}
	return
}

// cvm instance
type CbsTcInstanceInfosRepository interface {
	Get(id string) ([]string, error)
	GetInstanceInfosInfoByFilters([]string) (*cvm.DescribeInstancesResponse, error)
}

type CbsTcInstanceInfosRepositoryImpl struct {
	client *cvm.Client
	logger log.Logger
}

func (repo *CbsTcInstanceInfosRepositoryImpl) Get(id string) (instanceIds []string, err error) {
	req := cvm.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&id}
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	for _, instanceInfo := range resp.Response.InstanceSet {
		instanceIds = append(instanceIds, *instanceInfo.InstanceId)
	}
	return
}

func (repo *CbsTcInstanceInfosRepositoryImpl) GetInstanceInfosInfoByFilters(ids []string) (instances *cvm.DescribeInstancesResponse, err error) {
	req := cvm.NewDescribeInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100

	req.Offset = &offset
	req.Limit = &limit
	if ids != nil {
		for _, id := range ids {
			req.InstanceIds = []*string{&id}
		}
	}
	return repo.client.DescribeInstances(req)
}

func NewCbsTcInstanceInfosRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (CbsTcInstanceInfosRepository, error) {
	cli, err := client.NewCvmClient(cred, c)
	if err != nil {
		return nil, err
	}
	repo := &CbsTcInstanceInfosRepositoryImpl{
		client: cli,
		logger: logger,
	}
	return repo, nil
}
