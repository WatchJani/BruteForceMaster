package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"root/cmd"
	"root/server"
	"sync"
	"time"
)

const (
	SingleMod = "single"
	ThreadMod = "thread"
)

type Master struct {
	counterManager int
	sync.RWMutex
	slave []Slave
	time.Time
	net.Conn
}

func NewMaster(path string) (*Master, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	slave := make([]Slave, 2)
	if err := json.Unmarshal(data, &slave); err != nil {
		return nil, err
	}

	for index := range slave {
		conn, err := net.DialTimeout("tcp", slave[index].Addr, 5*time.Minute)
		if err != nil {
			continue
		}

		slave[index].Conn = conn
	}

	return &Master{
		slave: slave,
	}, nil
}

type Slave struct {
	Conn net.Conn
	Addr string `json:"addr"`
	Cors int    `json:"cors"`
}

type State struct {
	Hash string `json:"hash"` //Delete, put in Master
	Mod  string `json:"mod"`  //Delete
}

func (m *Master) Get(c *server.Ctx) {
	payload := &struct {
		Cors  int `json:"cors"`
		State     //better stay in Master
	}{}

	if err := json.Unmarshal([]byte(c.GetHeader()["body"]), payload); err != nil {
		c.ResWriter(err.Error())
		return
	}

	m.Lock()
	point := m.counterManager
	m.counterManager += 10_000_000_000 * m.slave[payload.Cors].Cors
	m.Unlock()

	res := &struct {
		Pointer int `json:"pointer"`
		State
	}{
		Pointer: point,
		State:   payload.State,
	}

	data, err := json.Marshal(res)
	if err != nil {
		c.ResWriter(err.Error())
		return
	}

	format := fmt.Sprintf("cmd: start\n body: %s\n", data)
	if err := c.ResWriter(format); err != nil {
		log.Println(err)
		//add to stack
	}
}

func (m *Master) Start(c *server.Ctx) {
	m.Time = time.Now()

	payload := State{}

	if err := json.Unmarshal([]byte(c.GetHeader()["body"]), &payload); err != nil {
		c.ResWriter(err.Error())
		return
	}

	numberOfNodes := len(m.slave)
	if payload.Mod == SingleMod || payload.Mod == ThreadMod {
		numberOfNodes = 1
	}

	m.Conn = c.Conn

	for index := range m.slave[:numberOfNodes] {
		go func() {
			// conn, err := net.DialTimeout("tcp", m.slave[index].Addr, 5*time.Minute)
			// if err != nil {
			// 	c.ResWriter(err.Error())
			// 	return
			// }

			// go m.Notification()

			m.Lock()
			point := m.counterManager
			m.counterManager += 10_000_000_000 * m.slave[index].Cors
			m.Unlock()

			res := &struct {
				Pointer int `json:"pointer"`
				State
			}{
				Pointer: point,
				State:   payload,
			}

			data, err := json.Marshal(res)
			if err != nil {
				c.ResWriter(err.Error())
				return
			}

			format := fmt.Sprintf("cmd: start\n body: %s\n", data)
			_, err = m.slave[index].Conn.Write([]byte(format))
			if err != nil {
				c.ResWriter(err.Error())
				//add range to stack
			}
		}()
	}
}

func (m *Master) Cancel(c *server.Ctx) {
	fmt.Println("stop process is activated")

	var wg sync.WaitGroup

	wg.Add(len(m.slave))
	for _, conn := range m.slave {
		go func(wg *sync.WaitGroup) {
			timeoutDuration := 10 * time.Second
			if err := conn.Conn.SetDeadline(time.Now().Add(timeoutDuration)); err != nil {
				
			}

			if _, err := conn.Conn.Write([]byte("asd")); err != nil {

			}

			wg.Done()
		}(&wg)
	}

	wg.Wait()
	fmt.Println("all server is finish job")
	fmt.Println("job is canceled")
}

// func (m *Master) Notification() {
// 	ticker := time.NewTicker(15 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		<-ticker.C
// 		m.Conn.Write([]byte(string(m.counterManager)))
// 	}
// }

func main() {
	master, err := NewMaster(cmd.Load())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	mux := server.NewRouter()

	mux.HandleFunc("get", master.Get)
	mux.HandleFunc("start", master.Start)
	mux.HandleFunc("cancel", master.Cancel)

	if err := server.ListenAndServe(":3000", mux); err != nil {
		log.Println(err)
	}
}
