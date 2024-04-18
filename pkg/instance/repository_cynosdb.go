package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cynosdb/v20190107"

	"tencentcloud-exporter/pkg/client"
	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/CYNOSDB_MYSQL", NewCynosdbTcInstanceRepository)
}

var dbType = "MYSQL"
var status = "running"

type CynosdbTcInstanceRepository struct {
	client *sdk.Client
	logger log.Logger
}

func (repo *CynosdbTcInstanceRepository) GetInstanceKey() string {
	return "InstanceId"
}

func (repo *CynosdbTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	level.Info(repo.logger).Log("start CynosdbTcInstanceRepository")
	req := sdk.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&id}
	req.DbType = &dbType
	req.Status = &status
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if len(resp.Response.InstanceSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.InstanceSet[0]
	instance, err = NewCynosdbTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *CynosdbTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *CynosdbTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeInstancesRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit
	req.DbType = &dbType
	req.Status = &status
getMoreInstances:
	resp, err := repo.client.DescribeInstances(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = int64(*resp.Response.TotalCount)
	}
	for _, meta := range resp.Response.InstanceSet {
		ins, e := NewCynosdbTcInstance(*meta.InstanceId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create Cynosdb instance fail", "id", *meta.InstanceId)
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

func NewCynosdbTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewCynosdbClient(cred, c)
	if err != nil {
		return
	}
	repo = &CynosdbTcInstanceRepository{
		client: cli,
		logger: logger,
	}
	return
}
