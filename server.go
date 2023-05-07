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
	"redis_by_hand/datastructure/zset"
	"redis_by_hand/network/frame"
	"redis_by_hand/network/packet"
	"redis_by_hand/serialization"
	"redis_by_hand/tools"
	"strconv"
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
	} else if cmdCnt == 4 && cmd == "zadd" {
		doZAdd(req, res)
	} else if cmdCnt == 3 && cmd == "zrem" {
		doZRem(req, res)
	} else if cmdCnt == 3 && cmd == "zscore" {
		doZScore(req, res)
	} else if cmdCnt == 6 && cmd == "zquery" {
		doZQuery(req, res)
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

	ent := (*datastructure.Entry)(unsafe.Pointer(node))
	if ent.Type_ != constants.TStr {
		res.Status = constants.ResErr
		msg := "expect string type"
		serialization.SerializeErr(&res.Data, constants.ErrTyp, &msg)
	}
	val := &ent.Val
	if len(*val) > config.MaxMsg {
		msg := "too big"
		serialization.SerializeErr(&res.Data, constants.Err2Big, &msg)
	}
	serialization.SerializeStr(&res.Data, val)
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
		ent := (*datastructure.Entry)(unsafe.Pointer(node))
		if ent.Type_ != constants.TStr {
			res.Status = constants.ResErr
			msg := "expect string type"
			serialization.SerializeErr(&res.Data, constants.ErrTyp, &msg)
			return
		}
		ent.Val = val
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

func doZAdd(req *packet.ReqPacket, res *packet.ResPacket) {
	var score float64 = 0
	score, err := strconv.ParseFloat(req.Payload[2].Str, 64)
	if err != nil {
		msg := "expect fp number"
		res.Status = constants.ResErr
		serialization.SerializeErr(&res.Data, constants.ErrArg, &msg)
		return
	}

	var key datastructure.Entry
	key.Key = req.Payload[1].Str
	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))
	hNode := g_data.db.Lookup(&key.Node, datastructure.EntryEq)

	var ent *datastructure.Entry
	if hNode == nil {
		ent = &datastructure.Entry{}
		ent.Key = key.Key
		ent.Node.HCode = key.Node.HCode
		ent.Type_ = constants.TZSet
		ent.ZSet = &zset.ZSet{}
		g_data.db.Insert(&ent.Node)
	} else {
		ent = (*datastructure.Entry)(unsafe.Pointer(hNode))
		if ent.Type_ != constants.TZSet {
			res.Status = constants.ResErr
			msg := "expected zset"
			serialization.SerializeErr(&res.Data, constants.ErrTyp, &msg)
			return
		}
	}

	name := req.Payload[3].Str
	added := ent.ZSet.Add(&name, uint32(len(name)), score)
	res.Status = constants.ResOk
	serialization.SerializeInt(&res.Data, int64(tools.BToI(added)))
	return
}

func expectZSet(out *packet.ResPacket, s *string, ent **datastructure.Entry) bool {
	var key datastructure.Entry
	key.Key = *s
	key.Node.HCode = tools.StrHash([]byte(key.Key), uint64(len(key.Key)))
	hNode := g_data.db.Lookup(&key.Node, datastructure.EntryEq)
	if hNode == nil {
		serialization.SerializeNil(&out.Data)
		return false
	}

	*ent = (*datastructure.Entry)(unsafe.Pointer(hNode))
	if (*ent).Type_ != constants.TZSet {
		msg := "expected zset"
		serialization.SerializeErr(&out.Data, constants.ErrTyp, &msg)
		return false
	}
	return true
}

func doZRem(req *packet.ReqPacket, res *packet.ResPacket) {
	var ent *datastructure.Entry
	if !expectZSet(res, &req.Payload[1].Str, &ent) {
		return
	}

	name := &req.Payload[2].Str
	zNode := ent.ZSet.Pop(name, uint32(len(*name)))
	status := int64(0)
	if zNode != nil {
		status = 1
	}
	serialization.SerializeInt(&res.Data, status)
	return
}

func doZScore(req *packet.ReqPacket, res *packet.ResPacket) {
	var ent *datastructure.Entry
	if !expectZSet(res, &req.Payload[1].Str, &ent) {
		return
	}

	name := &req.Payload[2].Str
	zNode := ent.ZSet.Lookup(name, uint32(len(*name)))
	if zNode != nil {
		serialization.SerializeDbl(&res.Data, zNode.Score)
		return
	}
	serialization.SerializeNil(&res.Data)
	return
}

func doZQuery(req *packet.ReqPacket, res *packet.ResPacket) {
	score := float64(0)
	score, err := strconv.ParseFloat(req.Payload[2].Str, 64)
	if err != nil {
		return
	}

	name := &req.Payload[3].Str
	offset := int64(0)
	limit := int64(0)

	if offset, err = strconv.ParseInt(req.Payload[4].Str, 10, 64); err != nil {
		msg := "expect int"
		serialization.SerializeErr(&res.Data, constants.ErrArg, &msg)
	}
	if limit, err = strconv.ParseInt(req.Payload[5].Str, 10, 64); err != nil {
		msg := "expect int"
		serialization.SerializeErr(&res.Data, constants.ErrArg, &msg)
	}

	var ent *datastructure.Entry
	if !expectZSet(res, &req.Payload[1].Str, &ent) {
		if serialization.DeserializeSerType(&res.Data) == constants.SerNil {
			res.Data = []byte{}
			serialization.SerializeArr(&res.Data, uint32(0))
		}
		return
	}

	if limit <= 0 {
		serialization.SerializeArr(&res.Data, 0)
	}
	zNode := ent.ZSet.Query(score, name, uint32(len(*name)), offset)
	serialization.SerializeArr(&res.Data, 0)
	n := uint32(0)
	for zNode != nil && int64(n) < limit {
		serialization.SerializeStr(&res.Data, zNode.Name)
		serialization.SerializeDbl(&res.Data, zNode.Score)
		n += 2
	}
	res.Status = constants.ResOk
	serialization.SerializeUpdateArr(&res.Data, n)
	return
}
