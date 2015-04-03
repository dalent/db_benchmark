package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"runtime"
	"strconv"
	"time"
)

var threads int
var passwd string
var url string
var db string
var mode string
var perSess = 10000
var data = struct {
	Table string
	Key   string
	Value string
}{"test", "key", "value"}

func init() {
	flag.IntVar(&threads, "t", 1, "multiple threads")
	flag.StringVar(&passwd, "p", "", "redis password")
	flag.StringVar(&url, "u", "", "db url")
	flag.StringVar(&db, "d", "", "db ")
}

func ReadRedis(re redis.Conn, val int) {
	defer re.Close()
	for i := perSess * val; i < perSess*(val+1); i++ {
		_, err := re.Do("HGET", data.Table, i)
		if err != nil {
			panic(err)
		}
	}
	finish <- 1
}
func WriteRedis(re redis.Conn, val int) {
	defer re.Close()
	for i := perSess * val; i < perSess*(val+1); i++ {
		_, err := re.Do("HSET", data.Table, i, data.Value)
		if err != nil {
			panic(err)
		}
	}
	finish <- 1
}
func RunRedis() {

	for i := 0; i < threads; i++ {
		re, err := redis.Dial("tcp", url)
		if err != nil {
			panic(err)
		}
		if passwd != "" {
			re.Do("AUTH", passwd)
		}
		switch mode {
		case "r":
			ReadRedis(re, i)
		case "w":
			WriteRedis(re, i)
		}
	}
}

func ReadMongo(sess *mgo.Session, val int) {
	defer sess.Close()
	col := sess.DB("test").C(data.Table)
	for i := perSess * val; i < perSess*(val+1); i++ {
		var test interface{}
		err := col.Find(bson.M{data.Key: strconv.Itoa(i)}).One(&test)
		if err != nil {
			panic(err)
		}
	}
	finish <- 1
}
func WriteMongo(sess *mgo.Session, val int) {
	defer sess.Close()
	col := sess.DB("test").C(data.Table)
	for i := perSess * val; i < perSess*(val+1); i++ {
		err := col.Insert(bson.M{data.Key: strconv.Itoa(i)})
		if err != nil {
			panic(err)
		}
	}
	finish <- 1
}
func RunMongo() {
	session, err := mgo.Dial(url)
	session.SetSafe(&mgo.Safe{W: 1})
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < threads; i++ {
		switch mode {
		case "r":
			ReadMongo(session.Copy(), i)
		case "w":
			WriteMongo(session.Copy(), i)
		}
	}

	session.Close()
}

var finish chan int

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(threads)
	if url == "" {
		fmt.Println(
			`-d db  //mongo redis
-u dburl  //db url 
-t  threads   //threads concurrency
-p  password //redis auth]`)
		return
	}

	mode = "w"
	finish = make(chan int, threads)
	for i := 0; i < 2; i++ {
		begin := time.Now()
		switch db {
		case "mongo":
			RunMongo()
		case "redis":
			RunRedis()
		}

		for i := 0; i < threads; i++ {
			<-finish
		}

		end := time.Since(begin)
		fmt.Printf("mode %s\nthreads: %d\nper thread: %d\ntotal time:%v\n\n", mode, threads, perSess, end)
		mode = "r"
	}

}
