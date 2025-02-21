//lint:file-ignore U1000 Not done yet

package lime

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"sync"
	"time"
)

type clientBuilder struct {
	addr       string
	bufferSize int

	tcpConfig *TCPConfig
}

func NewClientBuilder() ClientBuilder {
	return &clientBuilder{}
}

type ClientBuilder interface {
	Build(ctx context.Context) (*ClientChannel, error)
}

func (b *clientBuilder) Build(ctx context.Context) (*ClientChannel, error) {

	panic("Not implemented")
}

type Client interface {
	io.Closer
	MessageSender
	NotificationSender
	CommandSender
	CommandProcessor
}

type client struct {
	channel *ClientChannel
	builder ClientBuilder
	buildMu sync.Mutex
	mux     *EnvelopeMux
}

func (c *client) Close() error {
	c.buildMu.Lock()
	defer c.buildMu.Unlock()

	if c.channel == nil {
		return nil
	}

	if c.channel.Established() {
		// Try to close the session gracefully
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFunc()
		_, err := c.channel.FinishSession(ctx)
		c.channel = nil
		return err
	}

	err := c.channel.transport.Close()
	c.channel = nil
	return err
}

func (c *client) SendMessage(ctx context.Context, msg *Message) error {
	channel, err := c.getOrBuildChannel(ctx)
	if err != nil {
		return err
	}

	return channel.SendMessage(ctx, msg)
}

func (c *client) SendNotification(ctx context.Context, not *Notification) error {
	//TODO implement me
	panic("implement me")
}

func (c *client) SendCommand(ctx context.Context, cmd *Command) error {
	//TODO implement me
	panic("implement me")
}

func (c *client) ProcessCommand(ctx context.Context, cmd *Command) (*Command, error) {
	//TODO implement me
	panic("implement me")
}

func (c *client) channelOK() bool {
	return c.channel != nil && c.channel.Established()
}

func (c *client) getOrBuildChannel(ctx context.Context) (*ClientChannel, error) {
	if c.channelOK() {
		return c.channel, nil
	}

	c.buildMu.Lock()
	defer c.buildMu.Unlock()
	if c.channelOK() {
		return c.channel, nil
	}

	count := 0.0

	for ctx.Err() == nil {
		channel, err := c.buildChannel(ctx)
		if channel != nil {
			return channel, nil
		}

		interval := time.Duration(math.Pow(count, 2) * 100)
		log.Printf("build channel error on attempt %v, sleeping %v ms: %v", count, interval, err)
		time.Sleep(interval * time.Millisecond)
		count++
	}

	return nil, fmt.Errorf("getOrBuildChannel: %w", ctx.Err())
}

func (c *client) buildChannel(ctx context.Context) (*ClientChannel, error) {
	channel, err := c.builder.Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("getOrBuildChannel: %w", err)
	}

	c.channel = channel
	return c.channel, nil
}
