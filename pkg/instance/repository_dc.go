package instance

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"

	"tencentcloud-exporter/pkg/client"
	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/config"
)

func init() {
	registerRepository("QCE/DC", NewDcTcInstanceRepository)
}

type DcTcInstanceRepository struct {
	credential common.CredentialIface
	client     *sdk.Client
	logger     log.Logger
}

func (repo *DcTcInstanceRepository) GetInstanceKey() string {
	return "directConnectId"
}

func (repo *DcTcInstanceRepository) Get(id string) (instance TcInstance, err error) {
	req := sdk.NewDescribeDirectConnectsRequest()
	req.DirectConnectIds = []*string{&id}
	resp, err := repo.client.DescribeDirectConnects(req)
	if err != nil {
		return
	}
	if len(resp.Response.DirectConnectSet) != 1 {
		return nil, fmt.Errorf("Response instanceDetails size != 1, id=%s ", id)
	}
	meta := resp.Response.DirectConnectSet[0]
	instance, err = NewDcTcInstance(id, meta)
	if err != nil {
		return
	}
	return
}

func (repo *DcTcInstanceRepository) ListByIds(id []string) (instances []TcInstance, err error) {
	return
}

func (repo *DcTcInstanceRepository) ListByFilters(filters map[string]string) (instances []TcInstance, err error) {
	req := sdk.NewDescribeDirectConnectsRequest()
	var offset int64 = 0
	var limit int64 = 100
	var total int64 = -1

	req.Offset = &offset
	req.Limit = &limit

getMoreInstances:
	resp, err := repo.client.DescribeDirectConnects(req)
	if err != nil {
		return
	}
	if total == -1 {
		total = *resp.Response.TotalCount
	}
	for _, meta := range resp.Response.DirectConnectSet {
		ins, e := NewDcTcInstance(*meta.DirectConnectId, meta)
		if e != nil {
			level.Error(repo.logger).Log("msg", "Create dc instance fail", "id", *meta.DirectConnectId)
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

func NewDcTcInstanceRepository(cred common.CredentialIface, c *config.TencentConfig, logger log.Logger) (repo TcInstanceRepository, err error) {
	cli, err := client.NewDcClient(cred, c)
	if err != nil {
		return
	}
	repo = &DcTcInstanceRepository{
		credential: cred,
		client:     cli,
		logger:     logger,
	}
	return
}
