package broker

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestExtTermination(t *testing.T) {
	ctx, end := context.WithCancel(context.Background())
	b := CtxNew(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := b.Start()
		if !errors.Is(err, ErrBrokerTerminated) {
			t.Errorf("Expected '%s'; got '%s'", ErrBrokerTerminated, err)
		}
	}()
	end()
	wg.Wait()

}

func TestIntTermination(t *testing.T) {
	b := New()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := b.Start()
		if !errors.Is(err, ErrBrokerTerminated) {
			t.Errorf("Expected '%s'; got '%s'", ErrBrokerTerminated, err)
		}
	}()
	b.Stop()
	wg.Wait()
}

func TestAddClient(t *testing.T) {
	b := New()
	go b.Start()
	cli := make(clientCh)
	b.entering <- cli
	b.Stop()
	if len(b.clients) != 1 {
		t.Errorf("Expextet n Clients: %d, got %d", 1, len(b.clients))
	}
}

func TestRemoveClient(t *testing.T) {
	b := New()
	cli := make(clientCh)
	b.clients[cli] = true
	go b.Start()
	b.leaving <- cli
	//b.Stop()

	if len(b.clients) != 0 {
		t.Errorf("Expextet n Clients: %d, got %d", 0, len(b.clients))
	}
}

func TestWrite(t *testing.T) {
	cli := make(clientCh)
	b := New()
	go b.Start()

	msg := "HelloWorld!"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		recieved := string(<-cli)
		if recieved != msg {
			t.Errorf("Expected '%s'; got ''%s", msg, recieved)
		}
	}()
	b.entering <- cli
	n, err := b.Write([]byte(msg))
	wg.Wait()
	b.Stop()
	if n != len(msg) {
		t.Errorf("Should send %d; got %d", len(msg), n)
	}
	if err != nil {
		t.Errorf("Got error: %s", err)
	}
}

func TestWriteClosed(t *testing.T) {
	b := New()
	b.Stop()

	_, err := b.Write([]byte("Test"))
	if !errors.Is(ErrBrokerTerminated, err) {
		t.Errorf("Expected %s, got %s", ErrBrokerTerminated, err)
	}
}

func TestNoneBlocking(t *testing.T) {
	b := New()
	go b.Start()

	b.entering <- make(clientCh)

	b.Write([]byte("test"))
	b.Write([]byte("test"))

}

func TestConnect(t *testing.T) {
	b := New()
	go b.Start()
	defer b.Stop()

	b.Connect()
	if len(b.clients) != 1 {
		t.Errorf("Expextet n Clients: %d, got %d", 1, len(b.clients))
	}
}

func TestConnectClosed(t *testing.T) {
	b := New()
	b.Stop()

	_, err := b.Connect()
	if !errors.Is(ErrBrokerTerminated, err) {
		t.Errorf("Expected %s, got %s", ErrBrokerTerminated, err)
	}
}

func TestInfo(t *testing.T) {
	b := New()
	go b.Start()
	defer b.Stop()

	info := b.Info()

	if info.ClientCount != 0 {
		t.Errorf("Expected n Clients: %d, Got %d", 0, info.ClientCount)
	}
	if info.Terminated {
		t.Errorf("Expected %t; got %t", false, info.Terminated)
	}
}

func TestIsTerminated(t *testing.T) {
	b := New()
	if b.isTerminated() {
		t.Error("Should return false")
	}
	b.Stop()
	if !b.isTerminated() {
		t.Error("Should return true")
	}

}
