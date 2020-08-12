# Broker - A simple message broker in golang

Broker is a simple message broker for golang. It dispatches byte slices, aka messages, to connected clients.
It features a simple one-way, fire-and-forget message pipeline:
- Clients can be registered to a Broker and wait for recieving messages
- The Broker can be used to dispatch messages to all connected client

The message-dispatching is non-blocking. If a client can't recieve a message it will be skipped.
So there is no growing buffer, caused by blocking clients. 

> Critical messages, that must be recieved from every client
> should not be distribiuted!

## Installation
```
go get https://github.com/fsmarc/broker
```

## Example
```
//1. initialize broker
	br := New()

	//2. start the broker
	go br.Start()
	defer br.Stop() //8. close the broker

	//3. connect a client
	cl, _ := br.Connect()

	//4. wait for a message
	go func() {
		for {
			msg, err := cl.Listen()
			if err != nil {
				//9. recieve broker closed!
				fmt.Println("Broker closed!")
				//Output:
				////Broker closed!
				return
			}
			//6. Recieve message
			fmt.Printf("Message: %s", msg)
			//Output:
			//Message: Hello World!
		}
	}()

	//5. wait until everything is set up, then send a message
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintln(br, "Hello World!")
```