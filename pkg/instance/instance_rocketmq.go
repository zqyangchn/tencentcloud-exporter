package instance

import (
	"fmt"
	"reflect"

	sdk "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
)

type RocketMQTcInstance struct {
	baseTcInstance
	meta *sdk.RocketMQClusterDetail
}

func (ins *RocketMQTcInstance) GetFieldValuesByName(name string) (map[string][]string, error) {
	switch name {
	case "ClusterName":
		return map[string][]string{"ClusterName": {*ins.meta.Info.ClusterName}}, nil
	}

	return nil, fmt.Errorf("not found field name %s", name)
}

func (ins *RocketMQTcInstance) GetMeta() interface{} {
	return ins.meta
}

func NewRocketMQTcInstance(instanceId string, meta *sdk.RocketMQClusterDetail) (ins *RocketMQTcInstance, err error) {
	if instanceId == "" {
		return nil, fmt.Errorf("instanceId is empty ")
	}
	if meta == nil {
		return nil, fmt.Errorf("meta is empty ")
	}
	ins = &RocketMQTcInstance{
		baseTcInstance: baseTcInstance{
			instanceId: instanceId,
			value:      reflect.ValueOf(*meta),
		},
		meta: meta,
	}
	return
}
