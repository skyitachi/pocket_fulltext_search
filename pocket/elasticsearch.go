package pocket
import (
  "gopkg.in/olivere/elastic.v5"
  "os"
  "log"
  "context"
  "errors"
  "skyitachi/pocket_fulltext_search/util"
  "strings"
  "fmt"
  "encoding/json"
)

const IndexName = "pocket"
const TypeName = "item"
const (
  ExcerptPriority = 1
  TitlePriority = 2
  TagPriority = 3
)

type ElasticItem struct {
  Id string `json:"id"`
  Title string `json:"title"`
  Excerpt string `json:"excerpt"`
  Tags []string `json:"tags,omitempty"`
  Create int64 `json:"created,omitempty"`
  Update int64 `json:"updated,omitempty"`
  Url string `json:"source,omitempty"`
  Author string `json:"author,omitempty"`
  Favourite bool `json:"favourite"`
  TagStr string `json:"tag_str,omitempty"`
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
          "type":"text"
        },
				"excerpt":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"tags":{
					"type":"text"
				},
        "tag_str": {
          "type": "text"
        }
        "created": {
          "type": "date",
          "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
        },
        "updated": {
          "type": "date",
          "format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||epoch_millis"
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

func (es *ElasticSearch) RemoveIndex() {
  ret, err := es.client.DeleteIndex(IndexName).Do(context.Background())
  checkError(err)
  if !ret.Acknowledged {
    checkError(errors.New("delete elastic search index failed"))
  }
}

func (es *ElasticSearch) Index(item CompleteItem) (*elastic.IndexResponse, error) {
  esItem, err := transform(item)
  if err != nil {
    return nil, err
  }
  // check if exists
  mQuery := elastic.NewMatchQuery("id", esItem.Id)
  searchRet, err := es.client.Search().Index(IndexName).Type(TypeName).Query(mQuery).Size(1).Do(context.Background())
  if err != nil {
    log.Println(err)
  } else if searchRet.Hits.TotalHits > 0 {
    return nil, errors.New("item :" + esItem.Id + " exists")
  }
  return es.client.Index().Index(IndexName).Type(TypeName).BodyJson(esItem).Do(context.Background())
}

func (es *ElasticSearch) IndexList(itemList []CompleteItem) error {
  for _, item := range itemList {
    _, err := es.Index(item)
    if err != nil {
      log.Println(err)
      continue
    }
  }
  return nil
}

func (es *ElasticSearch) ItemExists(item CompleteItem) bool {
  mQuery := elastic.NewMatchQuery("id", item.Id)
  searchRet, err := es.client.Search().Index(IndexName).Type(TypeName).Query(mQuery).Size(1).Do(context.Background())
  if err != nil {
    return true
  } else if searchRet.Hits.TotalHits > 0 {
    return true
  }
  return false
}

func (es *ElasticSearch) ItemListNotExists(itemList []CompleteItem) []CompleteItem {
  ret := []CompleteItem{}
  for _, item := range itemList {
    if !es.ItemExists(item) {
      ret = append(ret, item)
    }
  }
  return ret
}

func (es *ElasticSearch) SearchByTags(tags []string) {
  bQuery := elastic.NewBoolQuery()
  for _, tag := range tags {
    mQuery := elastic.NewMatchQuery("tags", tag)
    bQuery = bQuery.Must(mQuery)
  }
  searchRet, err := es.client.Search().Index(IndexName).Type(TypeName).Query(bQuery).Do(context.Background())
  if err != nil {
    log.Println("search by tags error: ", err)
    return
  }
  PrettyPrintSearchResult(searchRet)
}

func (es *ElasticSearch) SearchByTitle(title string) {
  mQuery := elastic.NewMatchQuery("title", title)
  searchRet, err := es.client.Search().Index(IndexName).Type(TypeName).Query(mQuery).Do(context.Background())
  if err != nil {
    log.Println("search by title error: ", err)
  }
  PrettyPrintSearchResult(searchRet)
}

// exact term match
func (es *ElasticSearch) Search(text string) {
  bQuery := elastic.NewBoolQuery()
  mTagQuery := elastic.NewMatchQuery("tags", text).Boost(TagPriority)
  bQuery = bQuery.Should(mTagQuery)
  mTitleQuery := elastic.NewMatchQuery("title", text).Boost(TitlePriority)
  bQuery = bQuery.Should(mTitleQuery)
  mExcerptQuery := elastic.NewMatchQuery("excerpt", text).Boost(ExcerptPriority)
  bQuery = bQuery.Should(mExcerptQuery)
  searchRet, err :=
    es.client.Search().Index(IndexName).Type(TypeName).Query(bQuery).Do(context.Background())
  if err != nil {
    log.Println("search by text error: ", err)
  }
  PrettyPrintSearchResult(searchRet)
}

// wildcard match
func (es *ElasticSearch) SearchFull(text string) {
  wildStr := util.BuildWildCardString(text)
  bQuery := elastic.NewBoolQuery()
  mTagWQuery := elastic.NewWildcardQuery("tags", wildStr).Boost(TagPriority)
  bQuery = bQuery.Should(mTagWQuery)
  mTitleWQuery := elastic.NewWildcardQuery("title", wildStr).Boost(TitlePriority)
  bQuery = bQuery.Should(mTitleWQuery)
  mExcerptMQuery := elastic.NewWildcardQuery("excerpt", wildStr).Boost(ExcerptPriority)
  bQuery = bQuery.Should(mExcerptMQuery)
  searchRet, err :=
    es.client.Search().Index(IndexName).Type(TypeName).Query(bQuery).Do(context.Background())
  if err != nil {
    log.Println("search by full text error: ", err)
  }
  PrettyPrintSearchResult(searchRet)
}

func PrettyPrintSearchResult(ret *elastic.SearchResult) {
  fmt.Printf("match %d items\n", ret.Hits.TotalHits)
  for _, hit := range ret.Hits.Hits {
    var esItem ElasticItem
    err := json.Unmarshal(*hit.Source, &esItem)
    if err != nil {
      log.Println(err)
    } else {
      fmt.Printf("{\n  title: \"%s\", \n  source: \"%s\"\n}\n", esItem.Title, esItem.Url)
    }
  }
}

func transform(item CompleteItem) (ElasticItem, error) {
  if item.ResolvedId == "" {
    return ElasticItem{}, errors.New("resolved_id is empty")
  }
  cTime, err := util.Str2Time(item.Create)
  if err != nil {
    return ElasticItem{}, err
  }
  uTime, err := util.Str2Time(item.Update)
  if err != nil {
    return ElasticItem{}, err
  }
  favourite := false
  if item.Favorite == "1" {
    favourite = true
  }
  tagsList := GetTagList(item.Tags)
  tagStr := strings.Join(tagsList, ",")
  esItem := ElasticItem {
    Id: item.ResolvedId,
    Title: item.Title,
    Excerpt: item.Excerpt,
    Tags: tagsList,
    TagStr: tagStr,
    Create: cTime.Unix() * 1000,
    Update: uTime.Unix() * 1000,
    Url: item.Url,
    Favourite: favourite,
  }
  return esItem, nil
}
