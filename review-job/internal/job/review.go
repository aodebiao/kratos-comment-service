package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
	"review-job/internal/conf"
)

// 评价数据流处理
// 1.从kafka中获取MYSQL中数据变更消息
// 2.将数据写入ES

// JobWorker 自定义执行job的结构体，实现transport.Server
type JobWorker struct {
	kafkaReader *kafka.Reader
	esClient    *ESClient
	log         *log.Helper
}

func NewJobWorker(kafkaReader *kafka.Reader, esClient *ESClient, logger log.Logger) *JobWorker {
	return &JobWorker{
		kafkaReader: kafkaReader,
		esClient:    esClient,
		log:         log.NewHelper(logger),
	}
}

func NewKafkaReader(cfg *conf.Kafka) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupId,
		Topic:   cfg.Topic,
	})
}

func NewESClient(cfg *conf.Elasticsearch) (*ESClient, error) {
	cf := elasticsearch.Config{Addresses: cfg.Addresses}
	client, err := elasticsearch.NewTypedClient(cf)
	if err != nil {
		return nil, err
	}
	return &ESClient{
		index:       cfg.Index,
		TypedClient: client,
	}, nil
}

type ESClient struct {
	*elasticsearch.TypedClient
	index string
}

type Msg struct {
	Type     string                   `json:"type"`
	Database string                   `json:"database"`
	Table    string                   `json:"table"`
	IsDdl    bool                     `json:"isddl"`
	Data     []map[string]interface{} `json:"data"`
}

// Start 程序启动后干活的
func (job *JobWorker) Start(ctx context.Context) error {
	// 1.从kafka中获取MySQL中的数据变更消息
	job.log.Debugf("start job worker.....")
	for {
		m, err := job.kafkaReader.ReadMessage(ctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		if err != nil {
			job.log.Errorf("failed to read message: %v", err)
			break
		}
		fmt.Printf("message at topic/partition/offset %v/%v/%v:%s = %s \n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		msg := new(Msg)
		err = json.Unmarshal(m.Value, msg)
		if err != nil {
			job.log.Errorf("failed to unmarshal message: %v", err)
			continue
		}

		if msg.Type == "INSERT" {
			// 新增文档到Es
			for idx := range msg.Data {
				job.IndexDocument(msg.Data[idx])
			}
		} else {
			// 更新 文档
			for idx := range msg.Data {
				job.updateDocument(msg.Data[idx])

			}
		}
	}
	// 2.将完整的评价数据写入ES
	return nil
}

func (job *JobWorker) IndexDocument(doc map[string]interface{}) {
	reviewID := doc["review_id"].(string)

	resp, err := job.esClient.Index(job.esClient.index).Id(reviewID).Document(doc).Do(context.Background())
	if err != nil {
		job.log.Errorf("failed to index document: %v", err)
		return
	}
	job.log.Debugf("document indexed: %v", resp)
}

func (job *JobWorker) updateDocument(doc map[string]interface{}) {
	reviewID := doc["review_id"].(string)
	resp, err := job.esClient.Update(job.esClient.index, reviewID).Doc(doc).Do(context.Background())
	if err != nil {
		job.log.Errorf("failed to update document: %v", err)
		return
	}
	job.log.Debugf("document updated to %v\n", resp.Result)
}

// Stop kratos结束后调用的
func (j *JobWorker) Stop(ctx context.Context) error {
	j.log.Debugf("stopping job worker")
	return j.kafkaReader.Close()
}
