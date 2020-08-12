//Package broker provides a simple one-way message-broker. Messages can be dispatched tho clients, that are connected to the Broker.
//clients are read-only and can't send messages to the Broker. Only the Broker can dispatch messages to the clients.
package broker

import (
	"context"
	"errors"
)

var (
	//ErrBrokerTerminated is returned from Broker or Client functions if they are called when the Broker was terminated.
	ErrBrokerTerminated = errors.New("Broker: is Terminated")
)

type clientCh chan []byte

//Broker is a simple Message Broker. Messages emmited to the Broker via its write method will be dispatched to all connected clients.
//NOTE: Clients recieve their messages via channels. Client-Channels that are blocked will be skiped from message despatchment.
type Broker struct {
	messages chan []byte
	entering chan clientCh
	leaving  chan clientCh
	nClients chan int

	clients map[clientCh]bool

	ctx context.Context
	end context.CancelFunc
}

//Info holds current information of the Broker
type Info struct {
	ClientCount int  //Number of clients currently connected
	Terminated  bool //Indicates if broker is terminated
}

//CtxNew returnes a new Broker.
//An initialized Broker must be started by calling the start method and can be terminated by calling the stop method.
//If the provided context is terminated the Broker will also terminate.
func CtxNew(ctx context.Context) *Broker {
	nCtx, end := context.WithCancel(ctx)
	return &Broker{
		messages: make(chan []byte),
		entering: make(chan clientCh),
		leaving:  make(chan clientCh),
		nClients: make(chan int),
		clients:  make(map[clientCh]bool),
		ctx:      nCtx,
		end:      end,
	}
}

//New wraps CtxNew with a Background context.
func New() *Broker {
	return CtxNew(context.Background())
}

//Start starts the broker. Start have to be only called once.
func (br *Broker) Start() error {
	for {
		select {
		case msg := <-br.messages:
			for cli := range br.clients {
				select {
				case cli <- msg:
				default:
					//blocked channels will be skipped
				}
			}
		case cli := <-br.entering:
			br.clients[cli] = true
		case cli := <-br.leaving:
			delete(br.clients, cli)
			close(cli)
		case br.nClients <- len(br.clients):
		case <-br.ctx.Done():
			for cli := range br.clients {
				delete(br.clients, cli)
				close(cli)
			}
			return ErrBrokerTerminated
		}
	}
}

//Stop terminates a running broker. It closes all client connections.
func (br *Broker) Stop() {
	br.end()
}

//Write sends a message b to the broker and returns the number of bits n that are written.
//The broker dispatches the message to all clients that are not blocked.
//If Write is called on a terminated broker ErrBrokerTerminated will be returned.
func (br *Broker) Write(b []byte) (int, error) {
	select {
	case br.messages <- b:
		return len(b), nil
	case <-br.ctx.Done():
		return 0, ErrBrokerTerminated
	}
}

//Connect registers a new client on the broker and returns a Client instance.
//If Connect is called on a terminated broker ErrBrokerTerminated will be returned.
func (br *Broker) Connect() (*Client, error) {
	ch := make(chan []byte)
	select {
	case br.entering <- ch:
		return &Client{
			ch: ch,
			br: br,
		}, nil
	case <-br.ctx.Done():
		return nil, ErrBrokerTerminated
	}
}

//Info returns an Info-Object containing the current number of connected clients and the status of the broker.
func (br *Broker) Info() Info {
	return Info{
		ClientCount: <-br.nClients,
		Terminated:  br.isTerminated(),
	}
}

func (br *Broker) isTerminated() bool {
	select {
	case <-br.ctx.Done():
		return true
	default:
		return false
	}
}
