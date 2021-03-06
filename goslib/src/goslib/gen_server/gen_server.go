package gen_server

import (
	"errors"
	"fmt"
	"goslib/cmap"
	"goslib/logger"
)

var SIGN_STOP int = 1
var ServerRegisterMap = cmap.NewCMap()

type Packet struct {
	method string
	args   []interface{}
}

type SignPacket struct {
	signal           int
	reason           string
	response_channel chan *ResponsePacket
}

type ResponsePacket struct {
	result interface{}
	err    error
}

type GenServer struct {
	name         string
	callback     GenServerBehavior
	cast_channel chan []interface{}
	call_channel chan []interface{}
	sign_channel chan SignPacket
}

type GenServerBehavior interface {
	Init(args []interface{}) (err error)
	HandleCast(args []interface{})
	HandleCall(args []interface{}) (interface{}, error)
	Terminate(reason string) (err error)
}

func setGenServer(name string, instance *GenServer) {
	ServerRegisterMap.Set(name, instance)
}

func GetGenServer(name string) (*GenServer, bool) {
	v := ServerRegisterMap.Get(name)
	if v == nil {
		return nil, false
	} else {
		return v.(*GenServer), true
	}
}

func Exists(name string) bool {
	return ServerRegisterMap.Check(name)
}

func delGenServer(name string) {
	ServerRegisterMap.Delete(name)
}

func Start(server_name string, module GenServerBehavior, args ...interface{}) (gen_server *GenServer) {
	gen_server, exists := GetGenServer(server_name)
	if !exists {
		cast_channel := make(chan []interface{}, 1024)
		call_channel := make(chan []interface{})
		sign_channel := make(chan SignPacket)

		gen_server = &GenServer{
			name:         server_name,
			callback:     module,
			cast_channel: cast_channel,
			call_channel: call_channel,
			sign_channel: sign_channel}

		gen_server.callback.Init(args)

		go loop(gen_server) // Enter infinity loop

		setGenServer(server_name, gen_server)
	} else {
		fmt.Println(server_name, " is already exists!")
	}
	return gen_server
}

func Stop(server_name, reason string) error {
	if gen_server, exists := GetGenServer(server_name); exists {
		response_channel := make(chan *ResponsePacket)
		gen_server.sign_channel <- SignPacket{
			signal:           SIGN_STOP,
			reason:           reason,
			response_channel: response_channel,
		}
		response := <-response_channel
		return response.err
	} else {
		fmt.Println(server_name, " not found!")
		return nil
	}
}

func Call(server_name string, args ...interface{}) (interface{}, error) {
	if gen_server, exists := GetGenServer(server_name); exists {
		response_channel := make(chan *ResponsePacket)
		defer func() {
			close(response_channel)
		}()
		args = append(args, response_channel)
		gen_server.call_channel <- args
		packet := <-response_channel
		return packet.result, packet.err
	} else {
		errMsg := fmt.Sprintf("GenServer call failed: ", server_name, " server not found!")
		logger.ERR(errMsg)
		return nil, errors.New(errMsg)
	}
}

func Cast(server_name string, args ...interface{}) error {
	if gen_server, exists := GetGenServer(server_name); exists {
		gen_server.cast_channel <- args
		return nil
	} else {
		errMsg := fmt.Sprintf("GenServer cast failed: ", server_name, " server not found!")
		logger.ERR(errMsg)
		return errors.New(errMsg)
	}
}

func loop(gen_server *GenServer) {
	defer func() {
		terminate(gen_server)
	}()

	for {
		select {
		case args, ok := <-gen_server.cast_channel:
			if ok {
				gen_server.callback.HandleCast(args)
			}
		case args, ok := <-gen_server.call_channel:
			if ok {
				size := len(args)
				response_channel := args[size-1]
				result, err := gen_server.callback.HandleCall(args[0 : size-1])
				response_channel.(chan *ResponsePacket) <- &ResponsePacket{
					result: result,
					err:    err,
				}
			}
		case sign_packet, ok := <-gen_server.sign_channel:
			if ok {
				switch sign_packet.signal {
				case SIGN_STOP:
					if err := gen_server.callback.Terminate(sign_packet.reason); err != nil {
						logger.ERR("GenServer stop failed: ", err)
						sign_packet.response_channel <- &ResponsePacket{
							err: err,
						}
					} else {
						sign_packet.response_channel <- &ResponsePacket{
							err: nil,
						}
						return
					}
				}
			}
		}
	}
}

func terminate(gen_server *GenServer) {
	close(gen_server.cast_channel)
	close(gen_server.call_channel)
	close(gen_server.sign_channel)
	delGenServer(gen_server.name)
}
