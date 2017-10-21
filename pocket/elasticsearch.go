package pocket
import (
  "gopkg.in/olivere/elastic.v5"
  "os"
  "log"
  "context"
  "errors"
)

const IndexName = "pocket"
const TypeName = "item"

type ElasticItem struct {
  Id string `json:"id"`
  Title string `json:"title"`
  Excerpt string `json:"excerpt"`
  Tags []string `json:"tags,omitempty"`
}

type ElasticSearch struct {
  client *elastic.Client
}

func NewElasticClient() (*ElasticSearch, error) {
  errorlog := log.New(os.Stdout, "APP ", log.LstdFlags)

  // Obtain a client. You can also provide your own HTTP client here.
  client, err := elastic.NewClient(elastic.SetErrorLog(errorlog))
  if err != nil {
    return nil, err
  }
  info, code, err := client.Ping("http://127.0.0.1:9200").Do(context.Background())
  if err != nil {
    return nil, err
  }
  log.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
  return &ElasticSearch{
    client: client,
  },  nil
}

// create elastic search index and so on
func (es *ElasticSearch) Init() {
  exists, err := es.client.IndexExists(IndexName).Do(context.Background())
  checkError(err)
  if !exists {
    mapping := `
{
	"settings":{
		"number_of_shards":1,
		"number_of_replicas":0
	},
	"mappings":{
		"_default_": {
			"_all": {
				"enabled": true
			}
		},
		"pocket":{
			"properties":{
				"id":{
					"type":"keyword"
				},
        "title": {
          "type":"keyword"
        },
				"excerpt":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"tags":{
					"type":"keyword"
				}
			}
		}
	}
}
    `
    createIndex, err := es.client.CreateIndex(IndexName).Body(mapping).Do(context.Background())
    checkError(err)
    if !createIndex.Acknowledged {
      // Not acknowledged
      checkError(errors.New("creates elastic search index failed"))
    }
  }
}

func (es *ElasticSearch) Index(rawItem CompleteItem) (*elastic.IndexResponse, error) {
  esItem := ElasticItem{
    Id: rawItem.ResolvedId,
    Title: rawItem.Title,
    Excerpt: rawItem.Excerpt,
    Tags: GetTagList(rawItem.Tags),
  }
  return es.client.Index().Index(IndexName).Type(TypeName).BodyJson(esItem).Do(context.Background())
}

func (es *ElasticSearch) IndexList(itemList []CompleteItem) error {
  for _, item := range itemList {
    esItem := ElasticItem{
      Id: item.ResolvedId,
      Title: item.Title,
      Excerpt: item.Excerpt,
      Tags: GetTagList(item.Tags),
    }
    _, err := es.client.Index().Index(IndexName).Type(TypeName).BodyJson(esItem).Do(context.Background())
    if err != nil {
      return err
    }
  }
  return nil
}

