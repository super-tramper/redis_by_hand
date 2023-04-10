package network

import (
	"flag"
	"fmt"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"
	log "github.com/sirupsen/logrus"
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
	var req packet.ReqPacket
	if err := req.Decode(framePayload); err != nil {
		log.Errorf("react packet decode error: %v", err)
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

func doRequest(req *packet.ReqPacket) (*packet.ResPacket, error) {

}
