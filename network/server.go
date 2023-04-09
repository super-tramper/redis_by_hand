package network

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
	"redis_by_hand/network/frame"
	"redis_by_hand/network/packet"
	"time"
)

const PORT = 1234

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

func RunServer() {
	var port int
	var multicore bool
	flag.IntVar(&port, "port", PORT, "server port")
	flag.BoolVar(&multicore, "multicore", false, "multicore")
	flag.Parse()
	addr := fmt.Sprintf("tcp://:%d", port)
	customCodecServe(addr, multicore, false, nil)
}

func (cs *customCodecServer) React(framePayload []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	var p packet.Packet
	err := p.Decode(framePayload)
	if err != nil {
		fmt.Println("react: packet decode error:", err)
		action = gnet.Close // close the connection
		return
	}
	out = []byte{'h', 'i'}
	return
}
