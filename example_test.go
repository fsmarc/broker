package broker

import (
	"fmt"
	"time"
)

//ExampleSimpleBroker creates a Broker, connects a client and sends a message.
func ExampleBroker() {
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
}
