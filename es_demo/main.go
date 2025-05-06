package main

import (
	"context"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"strconv"
	"time"
)

func main() {
	cfg := elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	}

	client, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		fmt.Printf("NewTypedClient failed,err:%v\n", err)
		return
	}
	fmt.Println("NewTypedClient success")
	fmt.Println(client)
	// 创建index
	//createIndex(client)

	// 创建文档
	//indexDocument2(client)

	// 查询文档
	//getDocumentByID(client, "2")

	// 搜索文档
	//searchDocument2(client)

	//searchDocumentAll(client)

	aggregationDemo(client)
}

// aggregationDemo 聚合
func aggregationDemo(client *elasticsearch.TypedClient) {
	avgScoreAgg, err := client.Search().
		Index("my-review-1").
		Request(
			&search.Request{
				Size: some.Int(0),
				Aggregations: map[string]types.Aggregations{
					"avg_score": { // 将所有文档的 score 的平均值聚合为 avg_score
						Avg: &types.AverageAggregation{
							Field: some.String("score"),
						},
					},
				},
			},
		).Do(context.Background())
	if err != nil {
		fmt.Printf("aggregation failed, err:%v\n", err)
		return
	}
	v, ok := avgScoreAgg.Aggregations["avg_score"]
	if !ok {
		fmt.Printf("aggregation1 failed, err:%v\n", err)
		return
	}
	a, ok := v.(*types.AvgAggregate)
	if !ok {
		fmt.Printf("aggregation2 failed, err:%v\n", err)
		return
	}
	fmt.Printf("avgScore:%#v\n", a.Value)
}

// searchDocument 搜索所有文档
func searchDocumentAll(client *elasticsearch.TypedClient) {
	// 搜索文档
	resp, err := client.Search().
		Index("my-review-1").
		Query(&types.Query{
			MatchAll: &types.MatchAllQuery{},
		}).
		Do(context.Background())
	if err != nil {
		fmt.Printf("search document failed, err:%v\n", err)
		return
	}
	fmt.Printf("total: %d\n", resp.Hits.Total.Value)
	// 遍历所有结果
	for _, hit := range resp.Hits.Hits {
		fmt.Printf("%s\n", hit.Source_)
	}
}

// searchDocument2 指定条件搜索文档
func searchDocument2(client *elasticsearch.TypedClient) {
	// 搜索content中包含好评的文档
	resp, err := client.Search().
		Index("my-review-1").
		Query(&types.Query{
			MatchPhrase: map[string]types.MatchPhraseQuery{
				"content": {Query: "好评"},
			},
		}).
		Do(context.Background())
	if err != nil {
		fmt.Printf("search document failed, err:%v\n", err)
		return
	}
	fmt.Printf("total: %d\n", resp.Hits.Total.Value)
	// 遍历所有结果
	for _, hit := range resp.Hits.Hits {
		fmt.Printf("%s\n", hit.Source_)
	}
}

func getDocumentByID(client *elasticsearch.TypedClient, id string) {
	resp, err := client.Get("my-review-1", id).Do(context.Background())
	if err != nil {
		fmt.Printf("Get failed,err:%v\n", err)
		return
	}
	fmt.Printf("resp:%s\n", resp.Source_)
}

func indexDocument(client *elasticsearch.TypedClient) {
	d1 := Review{ID: 1,
		UserID:  1237112,
		Score:   5,
		Content: "这是一个好评",
		Tags: []Tag{
			{10000, "好评"},
			{1100, "物超所值"},
			{9000, "有图"},
		},
		PublishTime: time.Now(),
	}
	// 添加文档
	resp, err := client.Index("my-review-1").Id(strconv.FormatInt(d1.ID, 10)).Document(d1).Do(context.Background())
	if err != nil {
		fmt.Printf("client.Index failed,err:%v\n", err)
		return
	}
	fmt.Printf("resp:%v\n", resp.Result)
}

func indexDocument2(client *elasticsearch.TypedClient) {
	d1 := Review{ID: 2,
		UserID:  1237112,
		Score:   1,
		Content: "这是一个差评",
		Tags: []Tag{
			{20000, "差评"},
		},
		PublishTime: time.Now(),
	}
	// 添加文档
	resp, err := client.Index("my-review-1").Id(strconv.FormatInt(d1.ID, 10)).Document(d1).Do(context.Background())
	if err != nil {
		fmt.Printf("client.Index failed,err:%v\n", err)
		return
	}
	fmt.Printf("resp:%v\n", resp.Result)
}

type Review struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Score       uint8     `json:"score"`
	Content     string    `json:"content"`
	Tags        []Tag     `json:"tags"`
	PublishTime time.Time `json:"publishTime"`
}

type Tag struct {
	Code  int    `json:"code"`
	Title string `json:"title"`
}

func createIndex(client *elasticsearch.TypedClient) {
	create, err := client.Indices.Create("my-review-2").Do(context.Background())
	if err != nil {
		fmt.Printf("CreateIndex failed,err:%v\n", err)
		return
	}
	fmt.Println("CreateIndex success")
	fmt.Printf("ack index %v", create.Acknowledged)
}
