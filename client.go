package rcon

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	rcon "github.com/jankyjames/james4krcon"
	"github.com/sirupsen/logrus"
)

type Client interface {
	Do(command string, options ...opt) (string, error)
}

type client struct {
	console   *rcon.RemoteConsole
	lock      sync.Mutex
	connected bool
	ticker    *time.Ticker
	requests  chan request
}

type request struct {
	command  string
	search   *string
	response chan string
}

func New(host, password string) Client {
	// TODO: With options to customize connection
	c := &client{
		lock:     sync.Mutex{},
		requests: make(chan request),
		ticker:   time.NewTicker(time.Millisecond * 200),
	}

	logrus.WithFields(logrus.Fields{
		"host":     host,
		"password": password,
	}).Debug("creating RCON poller")

	go c.connectAndPoll(host, password)
	return c
}

func (c *client) connectAndPoll(host string, password string) {
	err := c.connect(host, password)
	if err != nil {
		logrus.WithError(err).Error("failed to connect to RCON")
		panic(err.Error()) // TODO: Handle this better?
	}

	err = c.poll()
	if err != nil {
		logrus.WithError(err).Error("polling quit with error")
		panic(err.Error()) // TODO: Handle this better?
	}

	c.connectAndPoll(host, password)
}

func (c *client) connect(host string, password string) error {
	for {
		dial, err := rcon.Dial(host, password)
		if err != nil {
			err = handleDialError(err)
			if err != nil {
				return err
			}
			time.Sleep(time.Second * 5)
			continue
		}
		c.lock.Lock()
		// TODO Check for other threads
		if c.console != nil {
			c.console.Close()
		}
		c.console = dial
		c.connected = true
		c.lock.Unlock()
		logrus.Info("RCON service connected!")
		return nil
	}
}

func (c *client) poll() error {
	defer func() {
		c.lock.Lock()
		err := c.console.Close()
		if err != nil {
			logrus.WithError(err).Error("error closing RCON console")
		}
		c.console = nil
		c.lock.Unlock()
	}()

	for {
		select {
		case req := <-c.requests:
			fields := logrus.Fields{
				"rcon_request": req.command,
			}
			if req.search != nil {
				fields["search"] = *req.search
			}

			requestId, err := c.console.Write(req.command)
			if err != nil {
				if isAborted(err) {
					c.requests <- req
					return nil
				}
				logrus.WithError(err).WithFields(fields).Error("failed to write an RCON command")
				return err
			}
			for {
				response, id, err := c.console.Read()
				if err != nil {
					if isAborted(err) {
						req.response <- "error: " + err.Error()
						return nil
					}
					logrus.WithError(err).WithFields(fields).Error("failed to read an RCON message")
					return err
				}

				logrus.WithFields(fields).WithField("rcon_message", response).
					Debug("rcon message")

				if (req.search != nil && strings.Contains(response, *req.search)) || id == requestId {
					req.response <- response
					break
				}
			}
		case <-c.ticker.C:
			err := c.read()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
	}
}

func (c *client) read() error {
	response, _, err := c.console.Read()
	if err != nil {
		return err
	}
	logrus.WithField("rcon_message", response).Debug("rcon polling to keepalive")
	return nil
}

func handleDialError(err error) error {
	switch {
	case isRefused(err):
		logrus.Debug("RCON service not listening")
	case os.IsTimeout(err):
		logrus.Debug("RCON service timing out")
	default:
		logrus.WithError(err).Warn("RCON service failed to connect")
		return err
	}

	return nil
}

func (c *client) Do(command string, options ...opt) (string, error) {
	req := request{
		command:  command,
		response: make(chan string),
	}
	for _, option := range options {
		option(&req)
	}
	// TODO: Reject/error on commands when RCON is NOT connected yet, or not listening at all.
	c.requests <- req
	res := <-req.response
	if strings.HasPrefix("error: ", res) {
		return "", errors.New(res)
	}
	return res, nil
}

type opt func(req *request)

func UseSearch(query string) opt {
	return func(req *request) {
		req.search = &query
	}
}
