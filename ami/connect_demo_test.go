package ami_test

import (
	"context"
	"fmt"
	"github.com/ezdev128/astgo/ami"
)

func ExampleConnect() {
	boot := make(chan *ami.Message, 1)

	conn, err := ami.Connect(
		"192.168.1.1:5038",
		ami.WithAuth("admin", "admin"), // AMI auth
		// add predefined subscriber
		ami.WithSubscribe(ami.SubscribeFullyBootedChanOnce(boot)),
		ami.WithSubscribe(func(ctx context.Context, msg *ami.Message) bool {
			fmt.Println(msg.Format()) // log everything
			return true               // keep subscribe
		}, ami.SubscribeSend(), // subscribe send message - default recv only
		))
	if err != nil {
		panic(err)
	}
	<-boot
	// AMI now FullyBooted
	_ = conn
}
