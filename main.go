package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Db *sqlx.DB

type Reader interface {
	Read(rc chan []byte)
}
type Writer interface {
	Write(wc chan *Message)
}

type ReadFromFlie struct {
	path string
}

type WriteToMySQL struct {
	MySQLDsn string
}
type LogProcess struct {
	rc    chan []byte
	wc    chan *Message
	read  Reader
	write Writer
}

type Message struct {
	ids      int
	ctime    string
	password string
}

func (r *ReadFromFlie) Read(rc chan []byte) {
	f, err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error:%s", err.Error()))
	}
	f.Seek(0, 2)
	rd := bufio.NewReader(f)
	for {
		lines, err := rd.ReadBytes('\n')
		if err == io.EOF {
			time.Sleep(500 * time.Millisecond)
			continue
		} else if err != nil {
			panic(fmt.Sprintf("ReadBytes error:%s", err.Error()))
		}
		rc <- lines[:len(lines)-1]
	}

}
func (w *WriteToMySQL) Write(wc chan *Message) {
	for {
		//fmt.Println(<-wc)
		a := <-wc
		fmt.Println(a)
		insert(a.ids, a.ctime, a.password)
	}
}

func (l *LogProcess) Process() {

	for v := range l.rc {
		//fmt.Println(string(v))
		//reg := regexp.MustCompile("\\d{4}-\\d{2}-\\d{2}_\\d{2}:\\d{2}:\\d{2}")
		reg := regexp.MustCompile(`([\d]+)\s+([\d]+-[\d]+-[\d]+_[\d]+:[\d]+:[\d]+)\s+([\w]+)`)
		//l.wc <- string(reg.Find(v))
		//<-l.rc
		ret := reg.FindStringSubmatch(string(v))
		if len(ret) != 4 {
			fmt.Println("length not eq 4")
			fmt.Println(len(ret))
			continue
		}
		message := &Message{}
		message.ids, _ = strconv.Atoi(ret[1])
		message.ctime = ret[2]
		message.password = ret[3]
		l.wc <- message

	}
}

func init() {
	database, err := sqlx.Open("mysql", "root:123456@tcp(192.168.0.170:3306)/gotest")
	if err != nil {
		fmt.Println("open mysql failed,", err)
		return
	}
	Db = database
}

func insert(ids int, ctime, password string) {
	r, err := Db.Exec("insert into logs_dbtest(logsid,ctime,password) values(?,?,?)", ids, ctime, password)
	if err != nil {
		fmt.Println("exec failed", err)
		return
	}
	id, err := r.LastInsertId()
	if err != nil {
		fmt.Println("find id faild", err)
		return
	}
	fmt.Println("insert succ:", id)
}
func main() {

	r := &ReadFromFlie{
		path: "D:\\binlog.log",
	}
	w := &WriteToMySQL{
		MySQLDsn: "root@123456tcp(192.168.0.170:3306)/gotest",
	}
	lp := &LogProcess{
		rc:    make(chan []byte),
		wc:    make(chan *Message),
		read:  r,
		write: w,
	}

	go lp.read.Read(lp.rc)
	go lp.Process()
	go lp.write.Write(lp.wc)
	//time.Sleep(100 * time.Second)
	http.ListenAndServe(":9193", nil)
}
