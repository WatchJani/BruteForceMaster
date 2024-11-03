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
}

func NewMaster(path string) (*Master, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	slave := make([]Slave, 2)
	json.Unmarshal(data, &slave)

	return &Master{
		slave: slave,
	}, nil
}

type Slave struct {
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

	for index := range m.slave[:numberOfNodes] {
		go func() {
			conn, err := net.DialTimeout("tcp", m.slave[index].Addr, 5*time.Minute)
			if err != nil {
				c.ResWriter(err.Error())
				return
			}

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
			_, err = conn.Write([]byte(format))
			if err != nil {
				c.ResWriter(err.Error())
				//add range to stack
			}
		}()
	}
}

func main() {
	master, err := NewMaster(cmd.Load())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	mux := server.NewRouter()

	mux.HandleFunc("start", master.Start)

	server.ListenAndServe(":3000", mux)
}
