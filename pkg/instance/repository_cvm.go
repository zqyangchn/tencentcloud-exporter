package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"tencentcloud-exporter/pkg/client"
	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CVM", NewCvmTcInstanceRepository)
}

type CvmTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *CvmTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *CvmTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&id}
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.InstanceSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.InstanceSet[0]
	instance, err = NewCvmTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CvmTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CvmTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.InstanceSet {
		ins, e := NewCvmTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create cvm instance fail", "id", *meta.InstanceId)
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

func NewCvmTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCvmClient(cred, c)
	if err != nil {
		return
	}
	repo = &CvmTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
