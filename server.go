package main

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
	log "github.com/sirupsen/logrus"
	"redis_by_hand/config"
	"redis_by_hand/constants"
	"redis_by_hand/datastructure"
	"redis_by_hand/datastructure/hashtable"
	"redis_by_hand/network/frame"
	"redis_by_hand/network/packet"
	"redis_by_hand/serialization"
	"redis_by_hand/tools"
	"time"
	"unsafe"
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
	flag.IntVar(&port, "port", config.PORT, "server port")
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
	res = &packet.ResPacket{}
	if cmdCnt == 2 && cmd == "get" {
		doGet(req, res)
	} else if cmdCnt == 3 && cmd == "set" {
		doSet(req, res)
	} else if cmdCnt == 2 && cmd == "del" {
		doDel(req, res)
	} else if cmdCnt == 1 && cmd == "keys" {
		doKeys(res)
	} else {
		res.Status = constants.ResErr
		msg := "Unknown cmd"
		serialization.SerializeErr(&res.Data, constants.ErrUnknown, &msg)
	}
	return
}

func doGet(req *packet.ReqPacket, res *packet.ResPacket) {
	key := datastructure.Entry{}
	key.Key = req.Payload[1].Str

	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))

	node := g_data.db.Lookup(&key.Node, datastructure.EntryEq)
	if node == nil {
		res.Status = constants.ResNx
		serialization.SerializeNil(&res.Data)
		return
	}

	val := (*datastructure.Entry)(unsafe.Pointer(node)).Val
	if len(val) > config.MaxMsg {
		msg := "too big"
		serialization.SerializeErr(&res.Data, constants.Err2Big, &msg)
	}
	serialization.SerializeStr(&res.Data, &val)
	res.Status = constants.ResOk

	return
}

func doSet(req *packet.ReqPacket, res *packet.ResPacket) {
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

	res.Status = constants.ResOk
	serialization.SerializeNil(&res.Data)

	return
}

func doDel(req *packet.ReqPacket, res *packet.ResPacket) {
	key := datastructure.Entry{}
	key.Key = req.Payload[1].Str
	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))

	node := g_data.db.Pop(&key.Node, datastructure.EntryEq)
	out := 0
	if node != nil {
		out = 1
	}

	res.Status = constants.ResOk
	serialization.SerializeInt(&res.Data, int64(out))
	return
}

func doKeys(res *packet.ResPacket) {
	serialization.SerializeArr(&res.Data, uint32(g_data.db.Size()))
	g_data.db.T1.Scan(datastructure.EntryKey, &res.Data)
	g_data.db.T2.Scan(datastructure.EntryKey, &res.Data)
}
