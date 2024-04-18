package instance

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	rocketmq "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
)

/*
	RocketMQ Instance Cache
*/

// TcRocketMQInstanceNamespaceCache Namespace
type TcRocketMQInstanceNamespaceCache struct {
	Raw            RocketMQTcInstanceRocketMQNameSpacesRepository
	cache          map[string][]*rocketmq.RocketMQNamespace
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcRocketMQInstanceNamespaceCache) GetRocketMQNamespacesInfo(
	instanceId string,
) ([]*rocketmq.RocketMQNamespace, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		namespace, ok := c.cache[instanceId]
		if ok {
			return namespace, nil
		}
	}

	namespace, err := c.Raw.GetRocketMQNamespacesInfo(instanceId)
	if err != nil {
		return nil, err
	}
	c.cache[instanceId] = namespace
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return namespace, nil
}

func NewTcRocketMQInstanceNamespaceCache(
	repo RocketMQTcInstanceRocketMQNameSpacesRepository, reloadInterval time.Duration, logger log.Logger,
) RocketMQTcInstanceRocketMQNameSpacesRepository {
	cache := &TcRocketMQInstanceNamespaceCache{
		Raw:            repo,
		cache:          make(map[string][]*rocketmq.RocketMQNamespace),
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

// TcRocketMQInstanceTopicsCache Topic
type TcRocketMQInstanceTopicsCache struct {
	Raw            RocketMQTcInstanceRocketMQTopicsRepository
	cache          map[string][]*rocketmq.RocketMQTopic
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcRocketMQInstanceTopicsCache) GetRocketMQTopicsInfo(
	instanceId string, namespaceId string,
) ([]*rocketmq.RocketMQTopic, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		topic, ok := c.cache[instanceId]
		if ok {
			return topic, nil
		}
	}

	topic, err := c.Raw.GetRocketMQTopicsInfo(instanceId, namespaceId)
	if err != nil {
		return nil, err
	}
	instanceIdNamspace := fmt.Sprintf("%v-%v", instanceId, namespaceId)
	c.cache[instanceIdNamspace] = topic
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Namespaces info from api", "instanceId", instanceId)
	return topic, nil
}

func NewTcRocketMQInstanceTopicsCache(
	repo RocketMQTcInstanceRocketMQTopicsRepository, reloadInterval time.Duration, logger log.Logger,
) RocketMQTcInstanceRocketMQTopicsRepository {
	cache := &TcRocketMQInstanceTopicsCache{
		Raw:            repo,
		cache:          make(map[string][]*rocketmq.RocketMQTopic),
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}

// TcRocketMQInstanceGroupsCache Group
type TcRocketMQInstanceGroupsCache struct {
	Raw            RocketMQTcInstanceRocketMQGroupsRepository
	cache          map[string][]*rocketmq.RocketMQGroup
	lastReloadTime map[string]time.Time
	reloadInterval time.Duration
	mu             sync.Mutex

	logger log.Logger
}

func (c *TcRocketMQInstanceGroupsCache) GetRocketMQGroupsInfo(
	instanceId string, namespaceId string, topic string,
) ([]*rocketmq.RocketMQGroup, error) {
	lrtime, exists := c.lastReloadTime[instanceId]
	if exists && time.Now().Sub(lrtime) < c.reloadInterval {
		group, ok := c.cache[instanceId]
		if ok {
			return group, nil
		}
	}

	group, err := c.Raw.GetRocketMQGroupsInfo(instanceId, namespaceId, topic)
	if err != nil {
		return nil, err
	}
	instanceIdNamspace := fmt.Sprintf("%v-%v", instanceId, namespaceId)
	c.cache[instanceIdNamspace] = group
	c.lastReloadTime[instanceId] = time.Now()
	level.Debug(c.logger).Log("msg", "Get RocketMQ Group info from api", "instanceId", instanceId)

	return group, nil
}

func NewTcRocketMQInstanceGroupsCache(
	repo RocketMQTcInstanceRocketMQGroupsRepository, reloadInterval time.Duration, logger log.Logger,
) RocketMQTcInstanceRocketMQGroupsRepository {
	cache := &TcRocketMQInstanceGroupsCache{
		Raw:            repo,
		cache:          make(map[string][]*rocketmq.RocketMQGroup),
		lastReloadTime: map[string]time.Time{},
		reloadInterval: reloadInterval,
		logger:         logger,
	}
	return cache
}
