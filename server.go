package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
	log "github.com/sirupsen/logrus"
	"redis_by_hand/datastructure"
	"redis_by_hand/datastructure/hashtable"
	"redis_by_hand/network/frame"
	"redis_by_hand/network/packet"
	"redis_by_hand/tools"
	"time"
	"unsafe"
)

const PORT = 1234
const MaxMsg = 4096
const (
	ResOk  = int32(0)
	ResErr = int32(1)
	ResNx  = int32(2)
)

var g_data = struct {
	db hashtable.HMap
}{}

type customCodecServer struct {
	*gnet.EventServer
	addr       string
	multicore  bool
	async      bool
	codec      gnet.ICodec
	workerPool *goroutine.Pool
}

func (cs *customCodecServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	return
}

func customCodecServe(addr string, multicore, async bool, codec gnet.ICodec) {
	var err error
	codec = frame.Frame{}
	cs := &customCodecServer{addr: addr, multicore: multicore, async: async, codec: codec, workerPool: goroutine.Default()}
	err = gnet.Serve(cs, addr, gnet.WithMulticore(multicore), gnet.WithTCPKeepAlive(time.Minute*5), gnet.WithCodec(codec))
	if err != nil {
		panic(err)
	}
}

func main() {
	var port int
	var multicore bool
	flag.IntVar(&port, "port", PORT, "server port")
	flag.BoolVar(&multicore, "multicore", false, "multicore")
	flag.Parse()
	addr := fmt.Sprintf("tcp://:%d", port)
	customCodecServe(addr, multicore, false, nil)
}

func (cs *customCodecServer) React(framePayload []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	req := packet.ReqPacket{}
	if err := req.Decode(framePayload); err != nil {
		log.Errorf("React: packet decode error: %v", err)
		action = gnet.Close
		return
	}
	res, err := doRequest(&req)
	if err != nil {
		log.Errorf("react: request error: %v", err)
		action = gnet.Close
		return
	}

	if out, err = res.Encode(); err != nil {
		log.Errorf("react error: %v", err)
		action = gnet.Close
		return
	}
	return
}

func doRequest(req *packet.ReqPacket) (res *packet.ResPacket, err error) {
	cmdCnt := req.StrCnt

	if int(cmdCnt) != len(req.Payload) {
		err = fmt.Errorf("bad request")
	}

	cmd := req.Payload[0].Str
	if cmdCnt == 2 && cmd == "get" {
		res = doGet(req)
	} else if cmdCnt == 3 && cmd == "set" {
		res = doSet(req)
	} else if cmdCnt == 2 && cmd == "del" {
		res = doDel(req)
	} else {
		res = &packet.ResPacket{Status: ResErr, Data: "Unknown cmd"}
	}
	return
}

func doGet(req *packet.ReqPacket) (res *packet.ResPacket) {
	res = &packet.ResPacket{}

	key := datastructure.Entry{}
	key.Key = req.Payload[1].Str

	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))

	node := g_data.db.Lookup(&key.Node, datastructure.EntryEq)
	if node == nil {
		res.Status = ResNx
		return
	}

	val := (*datastructure.Entry)(unsafe.Pointer(node)).Val
	if len(val) > MaxMsg {
		panic("value too long.")
	}
	res.Data = val
	res.Status = ResOk

	return
}

func doSet(req *packet.ReqPacket) (res *packet.ResPacket) {
	res = &packet.ResPacket{}
	key := datastructure.Entry{}
	key.Key = req.Payload[1].Str
	val := req.Payload[2].Str

	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))

	node := g_data.db.Lookup(&key.Node, datastructure.EntryEq)
	if node == nil {
		ent := datastructure.Entry{}
		ent.Key = key.Key
		ent.Node.HCode = key.Node.HCode
		ent.Val = req.Payload[2].Str
		g_data.db.Insert(&ent.Node)
	} else {
		(*datastructure.Entry)(unsafe.Pointer(node)).Val = val
	}

	res.Status = ResOk

	return
}

func doDel(req *packet.ReqPacket) (res *packet.ResPacket) {
	res = &packet.ResPacket{}
	key := datastructure.Entry{}
	key.Key = req.Payload[1].Str
	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))

	g_data.db.Pop(&key.Node, datastructure.EntryEq)

	res.Status = ResOk
	return
}
