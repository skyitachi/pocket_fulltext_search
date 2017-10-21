package collector

import (
  "skyitachi/pocket_fulltext_search/pocket"
  "fmt"
  "time"
  "log"
)

type Agent struct {
  pocketClient *pocket.Client
  es *pocket.ElasticSearch
  Exit chan int // 外部强制退出channel
  Done chan int
  Interval time.Duration // 拉取pocket数据interval
  SyncInterval time.Duration
}

func (agent *Agent) Start() {
  for {
    select {
    case <- agent.Exit:
      fmt.Printf("agent exits\n")
    case <- agent.Done:
      fmt.Println("agent done")
    case <- time.After(agent.Interval):
      ret, err := agent.pocketClient.GetLatestList(time.Now())
      if err != nil {
        log.Println("pull data from pocket failed, ", err)
      }
      log.Println("starting index data to es")
      agent.es.IndexList(ret)
      log.Printf("index data to es successfully with %d items\n", len(ret))
      // start pull pocket
    }
  }
}

func (agent *Agent) Sync() {
  offset, err := agent.pocketClient.ReadOffset()
  if err != nil {
    offset = 0
  }
  for {
    select {
    case <-agent.Exit:
      fmt.Println("agent exits")
    case <- agent.Done:
      fmt.Println("agent done")
    case <- time.After(agent.SyncInterval):
      // use offset get all list
      since := time.Unix(0, 0)
      var ret []pocket.CompleteItem
      var err error
      for {
        ret, err = agent.pocketClient.GetListAfter(10, offset, since)
        if err != nil {
          log.Println("sync data with error: ", err)
          continue
        } else if len(ret) == 0 {
          log.Println("no data to sync")
          agent.pocketClient.WriteOffset(offset)
          break
        }
        agent.es.IndexList(ret)
        fmt.Println("sync and store data successfully")
        offset += len(ret)
      }
    }
  }
}

func NewCollector(client *pocket.Client, es *pocket.ElasticSearch, interval time.Duration, syncInterval time.Duration) *Agent {
  return &Agent{
    pocketClient:client,
    Exit: make(chan int),
    Done: make(chan int),
    Interval: interval,
    SyncInterval: syncInterval,
    es: es,
  }
}

