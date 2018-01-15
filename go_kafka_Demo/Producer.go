package main

import (
	"encoding/json"
	"flag"
	"sync"
	"time"
	// "fmt"
	"log"
	"os"
	"os/signal"
	// "strconv"
	"strings"
	// "sync"

	"strconv"

	"syscall"

	"github.com/Shopify/sarama"
	zk "github.com/samuel/go-zookeeper/zk"
)

//运行命令：./Producer.exe -send=sync -topic=haharsw -tag=x86 -threads=1 -number=10 -msg='aaabbbccc'

var zkhost = flag.String("zk", "192.168.1.175:2181", "")
var topic = flag.String("topic", "Test", "")
var tag = flag.String("tag", "", "")
var msg = flag.String("msg", "31,3,122313472650116.79.194.242116.79.194.108395494415387060454600170394305863GNET,1,11,,15597036883,460017039430586,116.79.194.242,,38659,,147900675,538706045,116.79.194.10,3gnet,,01,,,46001,22,,0971,0971,,,0,,8686120273557078,20161223,134726,50,839549441,,,13505,28481,,,,,0700000GJSJ0099102000201612231350006INM.19.0.sn,000,1,744FC2F2,,2,6,,1,,0,2,01,,1,,002,002,7015101247730062,7015101247828364,89106070,4G00,702003,0,62000,,20100504151102,70,7015101247730062,20161223143736", "")
var send = flag.String("send", "async", "同步(sync),异步(async)")
var threads = flag.Int("threads", 1, "")
var number = flag.Int("number", 200000, "")

func produce(brokers []string, topic, tag, msg string) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGTERM)

	var (
		wg                          sync.WaitGroup
		enqueued, successes, errors int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _ = range producer.Successes() {
			successes++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			errors++
			if errors == 1 {
				log.Println("error:", err)
				signals <- os.Interrupt
			}
		}
	}()

	start := time.Now()
	last := start
	var i int
ProducerLoop:
	for {
		i++
		message := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(strconv.Itoa(i) + ":" + tag + ":" + msg),
		}
		if i%10000 == 0 {
			now := time.Now()
			log.Println(i, now.Sub(last), now.Sub(start))
			last = now
		}
		// time.Sleep(time.Second)
		select {
		case producer.Input() <- message:
			enqueued++

		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}
	wg.Wait()
	s := time.Now().Sub(start).Seconds()
	log.Printf("Successfully produced: %d; errors: %d; 成功 %d 条/秒\n", successes, errors, successes/int(s))
}

func syncProduce(brokers []string, topic, tag, msg string, number int, wg *sync.WaitGroup) {
	defer wg.Done()
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()
	var message *sarama.ProducerMessage
	start := time.Now()
	last := start
	for i := 1; ; i++ {
		message = &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(strconv.Itoa(i) + ":" + tag + ":" + msg),
		}
		// partition, offset, err := producer.SendMessage(message)
		_, _, err := producer.SendMessage(message)
		if err != nil {
			log.Printf("FAILED to send message: %s\n", err)
			break
		} else {
			if i%10000 == 0 {
				now := time.Now()
				log.Println(i, now.Sub(last), now.Sub(start))
				last = now
			}
			if number > 0 && i >= number {
				break
			}
			// log.Printf("> message sent to partition %d at offset %d\n", partition, offset)
		}
		// time.Sleep(time.Second)
	}
}

type topicInfo struct {
	Partitions map[string][]int `json:"partitions"`
}

type brokerInfo struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func checkTopic(zkhost, topic string) ([]string, bool) {
	c, _, err := zk.Connect(strings.Split(zkhost, ","), time.Second*10)
	if err != nil {
		log.Println(err)
	}
	defer c.Close()
	var brokers []string
	paths, _, err := c.Children("/brokers/ids")
	if err != nil {
		log.Println(err)
	} else {
		for _, path := range paths {
			v, _, err := c.Get("/brokers/ids/" + path)
			if err != nil {
				log.Println(err, "/brokers/ids/"+path)
			} else {
				var b brokerInfo
				err = json.Unmarshal(v, &b)
				if err != nil {
					log.Println(err)
				}
				brokers = append(brokers, b.Host+":"+strconv.Itoa(b.Port))
			}
		}
	}
	v, _, err := c.Get("/brokers/topics/" + topic)
	if err != nil {
		log.Println(err)
	}
	var t topicInfo
	// t.Partitions = make(map[string][]int)
	err = json.Unmarshal(v, &t)
	if err != nil {
		log.Println(err)
	}
	if len(t.Partitions) == 0 {
		return brokers, false
	}
	return brokers, true
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()
	brokers, ok := checkTopic(*zkhost, *topic)
	if len(brokers) == 0 || !ok {
		return
	}
	log.Println(brokers)
	log.Println("开始发送消息")
	switch *send {
	case "sync":
		start := time.Now()
		var wg sync.WaitGroup
		for i := 0; i < *threads; i++ {
			wg.Add(1)
			go syncProduce(brokers, *topic, *tag, *msg, *number, &wg)
		}
		wg.Wait()
		d := time.Now().Sub(start)
		c := *threads * *number
		intD := int(d.Seconds())
		if intD==0{
			intD = 1
		}
		log.Println(*threads, c, d, c/intD, "条/秒")
	case "async":
		produce(brokers, *topic, *tag, *msg)
	}
}
