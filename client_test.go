package broker

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestStop(t *testing.T) {
	b := New()
	go b.Start()

	cli, _ := b.Connect()

	if n := len(b.clients); n != 1 {
		t.Errorf("Expected n clients: %d, got %d", 1, n)
	}

	err := cli.Stop()
	if err != nil {
		t.Errorf("Got error: %s", err)
	}

	if n := b.Info().ClientCount; n != 0 {
		t.Errorf("Expected n clients: %d, got %d", 0, n)
	}

	_, err = cli.Listen()
	if !errors.Is(err, ErrBrokerTerminated) {
		t.Errorf("Expected err: %s, got %s", ErrBrokerTerminated, err)
	}
}

func TestListen(t *testing.T) {
	b := New()
	go b.Start()
	defer b.Stop()

	cli, _ := b.Connect()

	var wg sync.WaitGroup
	msg := "Test"

	go func() {
		wg.Add(1)
		recieved, err := cli.Listen()
		if err != nil {
			t.Errorf("Got error: %s", err)
		}
		if string(recieved) != msg {
			t.Errorf("Expected: %s, got: %s", msg, recieved)
		}
		wg.Done()
	}()

	b.Write([]byte(msg))
	wg.Wait()
}

func TestCtxListen(t *testing.T) {
	b := New()
	go b.Start()
	defer b.Stop()

	cli, _ := b.Connect()

	ctx, done := context.WithCancel(context.Background())

	done()
	_, err := cli.CtxListen(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected %s, Got %s", context.Canceled, err)
	}
}

func TestManualSuccess(t *testing.T) {
	b := New()
	go b.Start()
	//defer b.Stop()

	cli, _ := b.Connect()
	msg := "Test"

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()
		select {
		case recieved, ok := <-cli.Message():
			if ok && string(recieved) != msg {
				t.Errorf("Expected: %s; got: %s", msg, recieved)
			}
			return
		case <-cli.Done():
			t.Error("cli should not be terminated")
			return
		}
	}()
	b.Write([]byte(msg))
	wg.Wait()
}

func TestManualFail(t *testing.T) {
	b := New()
	go b.Start()

	cli, _ := b.Connect()
	b.Stop()

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		defer wg.Done()
		select {
		case _, ok := <-cli.Message():
			if ok {
				t.Errorf("Should not get a Message")
			}
		case <-cli.Done():
			return
		}
	}()
	wg.Wait()
}

func TestTerminated(t *testing.T) {
	b := New()
	go b.Start()

	cli, _ := b.Connect()
	b.Stop()

	_, err := cli.Listen()
	if !errors.Is(err, ErrBrokerTerminated) {
		t.Errorf("Expexted error: %s; got %s", ErrBrokerTerminated, err)
	}

	_, err = cli.CtxListen(context.Background())
	if !errors.Is(err, ErrBrokerTerminated) {
		t.Errorf("Expexted error: %s; got %s", ErrBrokerTerminated, err)
	}

	err = cli.Stop()
	if !errors.Is(err, ErrBrokerTerminated) {
		t.Errorf("Expexted error: %s; got %s", ErrBrokerTerminated, err)
	}
}
