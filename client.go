package broker

import "context"

//Client represents a connection to a Broker. Clients can recieve messages from a Broker.
//They are read-only, to write a Message the Broker's write method must be used.
type Client struct {
	ch chan []byte
	br *Broker
}

//Stop detaches the client from the Broker. If Stop is called while the Broker is terminated ErrBrokerTerminated will be returned.
func (c *Client) Stop() error {
	select {
	case c.br.leaving <- c.ch:
		return nil
	case <-c.br.ctx.Done():
		return ErrBrokerTerminated
	}
}

//Listen returns a message from the broker. It blocks until a message is recieved. If the Broker is terminated ErrBrokerTerminated will be returned.
func (c *Client) Listen() ([]byte, error) {
	for msg := range c.ch {
		return msg, nil
	}
	return nil, ErrBrokerTerminated
}

//CtxListen returns a message from the broker. It blocks until a message is recieved or the provided context is terminated.
func (c *Client) CtxListen(ctx context.Context) ([]byte, error) {
	select {
	case msg, ok := <-c.ch:
		if !ok {
			return nil, ErrBrokerTerminated
		}
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

//Message returns the clients recieving message channel.
func (c *Client) Message() <-chan []byte {
	return c.ch
}

//Done returnes the Broker's termination channel. This channel will be closed when the Broker is terminated.
func (c *Client) Done() <-chan struct{} {
	return c.br.ctx.Done()
}
