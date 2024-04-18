package client

import (
	"net"
	"net/http"
	"net/url"
	"time"

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	cdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdb/v20170320"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	cfs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"
	kafka "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ckafka/v20190819"
	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	cmq "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cmq/v20190304"
	tccommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcregions "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	cynosdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cynosdb/v20190107"
	dc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dc/v20180410"
	dcdb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dcdb/v20180411"
	dts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20180330"
	dtsNew "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dts/v20211206"
	es "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/es/v20180416"
	gaap "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/gaap/v20180529"
	lh "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	mariadb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mariadb/v20170312"
	memcached "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/memcached/v20190318"
	mongodb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/mongodb/v20190725"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	pg "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/postgres/v20170312"
	redis "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/redis/v20180412"
	sqlserver "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sqlserver/v20180328"
	rocketmq "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tdmq/v20200217"
	tse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tse/v20201207"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
	waf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/waf/v20180125"
	cos "github.com/tencentyun/cos-go-sdk-v5"

	"tencentcloud-exporter/pkg/common"
	"tencentcloud-exporter/pkg/config"
)

func NewMonitorClient(cred common.CredentialIface, conf *config.TencentConfig, region string) (*monitor.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "monitor.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "monitor.tencentcloudapi.com"
	}
	return newClient(cred, region, cpf)
}

func newClient(
	credential common.CredentialIface,
	region string,
	clientProfile *tcprofile.ClientProfile,
) (client *monitor.Client, err error) {
	client = &monitor.Client{}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 5 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          0,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	clientProfile.HttpProfile.ReqTimeout = 5
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile).WithHttpTransport(transport)
	return
}

func NewMongodbClient(cred common.CredentialIface, conf *config.TencentConfig) (*mongodb.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "mongodb.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "mongodb.tencentcloudapi.com"
	}
	return mongodb.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCdbClient(cred common.CredentialIface, conf *config.TencentConfig) (*cdb.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cdb.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cdb.tencentcloudapi.com"
	}
	return cdb.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCvmClient(cred common.CredentialIface, conf *config.TencentConfig) (*cvm.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cvm.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"
	}
	return cvm.NewClient(cred, conf.Credential.Region, cpf)
}

func NewRedisClient(cred common.CredentialIface, conf *config.TencentConfig) (*redis.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "redis.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "redis.tencentcloudapi.com"
	}
	return redis.NewClient(cred, conf.Credential.Region, cpf)
}

func NewDcClient(cred common.CredentialIface, conf *config.TencentConfig) (*dc.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "dc.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "dc.tencentcloudapi.com"
	}
	return dc.NewClient(cred, conf.Credential.Region, cpf)
}

func NewClbClient(cred common.CredentialIface, conf *config.TencentConfig) (*clb.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "clb.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "clb.tencentcloudapi.com"
	}
	return clb.NewClient(cred, conf.Credential.Region, cpf)
}

func NewVpvClient(cred common.CredentialIface, conf *config.TencentConfig) (*vpc.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "vpc.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	}
	return vpc.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCbsClient(cred common.CredentialIface, conf *config.TencentConfig) (*cbs.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cbs.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cbs.tencentcloudapi.com"
	}
	return cbs.NewClient(cred, conf.Credential.Region, cpf)
}

func NewSqlServerClient(cred common.CredentialIface, conf *config.TencentConfig) (*sqlserver.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "sqlserver.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "sqlserver.tencentcloudapi.com"
	}
	return sqlserver.NewClient(cred, conf.Credential.Region, cpf)
}

func NewMariaDBClient(cred common.CredentialIface, conf *config.TencentConfig) (*mariadb.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "mariadb.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "mariadb.tencentcloudapi.com"
	}
	return mariadb.NewClient(cred, conf.Credential.Region, cpf)
}

func NewESClient(cred common.CredentialIface, conf *config.TencentConfig) (*es.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "es.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "es.tencentcloudapi.com"
	}
	return es.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCMQClient(cred common.CredentialIface, conf *config.TencentConfig) (*cmq.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cmq.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cmq.tencentcloudapi.com"
	}
	return cmq.NewClient(cred, conf.Credential.Region, cpf)
}

func NewPGClient(cred common.CredentialIface, conf *config.TencentConfig) (*pg.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "postgres.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "postgres.tencentcloudapi.com"
	}
	return pg.NewClient(cred, conf.Credential.Region, cpf)
}

func NewMemcacheClient(cred common.CredentialIface, conf *config.TencentConfig) (*memcached.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "memcached.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "memcached.tencentcloudapi.com"
	}
	return memcached.NewClient(cred, conf.Credential.Region, cpf)
}

func NewLighthouseClient(cred common.CredentialIface, conf *config.TencentConfig) (*lh.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "lighthouse.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	}
	return lh.NewClient(cred, conf.Credential.Region, cpf)
}

func NewKafkaClient(cred common.CredentialIface, conf *config.TencentConfig) (*kafka.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "ckafka.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "ckafka.tencentcloudapi.com"
	}
	return kafka.NewClient(cred, conf.Credential.Region, cpf)
}

func NewDCDBClient(cred common.CredentialIface, conf *config.TencentConfig) (*dcdb.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "dcdb.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "dcdb.tencentcloudapi.com"
	}
	return dcdb.NewClient(cred, conf.Credential.Region, cpf)
}

func NewRocketMQClient(cred common.CredentialIface, conf *config.TencentConfig) (*rocketmq.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "tdmq.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "tdmq.tencentcloudapi.com"
	}
	return rocketmq.NewClient(cred, conf.Credential.Region, cpf)
}

func NewTseClient(cred common.CredentialIface, conf *config.TencentConfig) (*tse.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "tse.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "tse.tencentcloudapi.com"
	}
	return tse.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCynosdbClient(cred common.CredentialIface, conf *config.TencentConfig) (*cynosdb.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cynosdb.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cynosdb.tencentcloudapi.com"
	}
	return cynosdb.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCdnClient(cred common.CredentialIface, conf *config.TencentConfig) (*cdn.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cdn.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cdn.tencentcloudapi.com"
	}
	return cdn.NewClient(cred, "", cpf)
}

func NewCosClient(cred common.CredentialIface, conf *config.TencentConfig) (*cos.Client, error) {
	// 用于Get Service 查询, service域名暂时只支持外网
	su, _ := url.Parse("http://cos." + conf.Credential.Region + ".myqcloud.com")
	b := &cos.BaseURL{BucketURL: nil, ServiceURL: su}
	client := &cos.Client{}
	if conf.Credential.Role == "" {
		client = cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  conf.Credential.AccessKey,
				SecretKey: conf.Credential.SecretKey,
			},
		})
	} else {
		client = cos.NewClient(b, &http.Client{
			Transport: common.NewCredentialTransport(cred.GetRole()),
		})
	}

	return client, nil
}

func NewDTSClient(cred common.CredentialIface, conf *config.TencentConfig) (*dts.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "dts.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "dts.tencentcloudapi.com"
	}
	return dts.NewClient(cred, conf.Credential.Region, cpf)
}
func NewDTSNewClient(cred common.CredentialIface, conf *config.TencentConfig) (*dtsNew.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "dts.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "dts.tencentcloudapi.com"
	}
	return dtsNew.NewClient(cred, conf.Credential.Region, cpf)
}

func NewGAAPClient(cred common.CredentialIface, conf *config.TencentConfig) (*gaap.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "gaap.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "gaap.tencentcloudapi.com"
	}
	return gaap.NewClient(cred, conf.Credential.Region, cpf)
}

func NewGAAPCommonClient(cred common.CredentialIface, conf *config.TencentConfig) *tccommon.Client {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "gaap.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "gaap.tencentcloudapi.com"
	}
	cpf.HttpProfile.ReqMethod = "POST"
	return tccommon.NewCommonClient(cred, tcregions.Guangzhou, cpf)
}

func NewWafClient(cred common.CredentialIface, conf *config.TencentConfig) (*waf.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "waf.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "waf.tencentcloudapi.com"
	}
	return waf.NewClient(cred, conf.Credential.Region, cpf)
}

func NewCfsClient(cred common.CredentialIface, conf *config.TencentConfig) (*cfs.Client, error) {
	cpf := tcprofile.NewClientProfile()
	if conf.Credential.IsInternal == true {
		cpf.HttpProfile.Endpoint = "cfs.internal.tencentcloudapi.com"
	} else {
		cpf.HttpProfile.Endpoint = "cfs.tencentcloudapi.com"
	}
	return cfs.NewClient(cred, conf.Credential.Region, cpf)
}
