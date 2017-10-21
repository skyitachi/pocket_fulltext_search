package collector

import (
  "skyitachi/pocket_fulltext_search/pocket"
  "fmt"
  "time"
  "log"
  "skyitachi/pocket_fulltext_search/util"
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
      // start pull pocket
    }
  }
}

func (agent *Agent) Sync() {
  fetched := false
  since := time.Now()
  for {
    select {
    case <-agent.Exit:
      fmt.Println("agent exits")
    case <- agent.Done:
      fmt.Println("agent done")
    case <- time.After(agent.SyncInterval):
      var ret []pocket.CompleteItem
      var err error
      if fetched {
        ret, err = agent.pocketClient.GetListAfter(10,0, since)
        if err != nil {
          log.Println("sync data with error: ", err)
          goto currentLoopEnd
        } else if len(ret) == 0 {
          log.Println("no data to sync")
          goto currentLoopEnd
        }
        agent.es.IndexList(ret)
        fmt.Println("sync and store data successfully")
      } else {
        ret, err = agent.pocketClient.GetAllList(1, 0)
        if err != nil || len(ret) == 0 {
          log.Println("no data to sync")
          goto currentLoopEnd
        }
        agent.es.IndexList(ret)
        fmt.Println("sync and store data successfully")
        fetched = true
      }
      if len(ret) == 0 {
        goto currentLoopEnd
      }
      parsed, err := util.Str2Time(ret[len(ret) - 1].Create)
      if err != nil {
        fmt.Println("parsed date error, ", err)
        log.Println("sync data with unexpect timestamp: ", ret[len(ret) - 1].Create)
        goto currentLoopEnd
      }
      fmt.Println("parased date is ", parsed)
      since = parsed
    }
    currentLoopEnd:
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

