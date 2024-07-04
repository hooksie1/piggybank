// Copyright 2023 Cover Whale Insurance Solutions Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/executor"
	"github.com/CoverWhale/logr"
	"github.com/nats-io/nats.go"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type NATSClient struct {
	Subject string
	Servers string
	Options []nats.Option
	Conn    *nats.Conn
	JS      nats.JetStreamContext
	NATSGraph
}

type NATSGraph struct {
	ExecutableSchema graphql.ExecutableSchema
	Exec             *executor.Executor
}

type ClientOpt func(*NATSClient)

func NewNATSClient(subject string, servers []string, opts ...ClientOpt) *NATSClient {
	n := NATSClient{
		Subject: subject,
		Servers: strings.Join(servers, ","),
	}

	for _, v := range opts {
		v(&n)
	}

	return &n
}

func SetServers(s string) ClientOpt {
	return func(n *NATSClient) {
		n.Servers = s
	}
}

func SetOptions(opts ...nats.Option) ClientOpt {
	return func(n *NATSClient) {
		n.Options = opts
	}
}

func SetSubject(s string) ClientOpt {
	return func(n *NATSClient) {
		n.Subject = s
	}
}

func SetGraphQLExecutableSchema(e graphql.ExecutableSchema) ClientOpt {
	ng := NATSGraph{
		ExecutableSchema: e,
		Exec:             executor.New(e),
	}
	return func(n *NATSClient) {
		n.NATSGraph = ng
	}
}

func (n *NATSClient) Connect() error {
	nc, err := nats.Connect(n.Servers, n.Options...)
	if err != nil {
		return err
	}

	n.Conn = nc
	js, err := nc.JetStream()
	if err != nil {
		return err
	}

	n.JS = js

	return nil
}

func (n *NATSClient) Resolve(errChan chan<- error) {
	if n.NATSGraph.ExecutableSchema == nil || n.NATSGraph.Exec == nil {
		errChan <- fmt.Errorf("executable schema must be set")
	}

	n.resolve()
}

func (n *NATSClient) resolve() {
	subject := fmt.Sprintf("%s.graphql", strings.TrimSuffix(n.Subject, ".>"))
	logr.Infof("listening for requests on %s", subject)

	_, err := n.Conn.Subscribe(subject, n.HandleAndLogRequests)
	if err != nil {
		logr.Errorf("Error in subscribing: %s", err)
	}
}

func (n *NATSClient) HandleAndLogRequests(m *nats.Msg) {
	ctx := context.Background()

	defer func() {
		if err := recover(); err != nil {
			err := n.Exec.PresentRecoveredError(ctx, err)
			gqlErr, _ := err.(*gqlerror.Error)
			resp := &graphql.Response{Errors: []*gqlerror.Error{gqlErr}}
			natsResponse(resp)
		}
	}()

	logr.Debugf("on subjeect %s, received request %+v", m.Subject, string(m.Data))
	ctx = graphql.StartOperationTrace(ctx)

	start := time.Now()

	params := &graphql.RawParams{
		ReadTime: graphql.TraceTiming{
			Start: start,
			End:   graphql.Now(),
		},
	}

	bodyReader := io.NopCloser(strings.NewReader(string(m.Data)))
	if err := jsonDecode(bodyReader, &params); err != nil {
		gqlErr := gqlerror.Errorf(
			"json request body could not be decoded: %+v body:%s",
			err,
			string(m.Data),
		)
		resp := n.Exec.DispatchError(ctx, gqlerror.List{gqlErr})
		if err := m.RespondMsg(natsResponse(resp)); err != nil {
			logr.Errorf("error sending message: %s", err)
		}
		return
	}

	rc, Operr := n.Exec.CreateOperationContext(ctx, params)
	if Operr != nil {
		resp := n.Exec.DispatchError(graphql.WithOperationContext(ctx, rc), Operr)
		m.RespondMsg(natsResponse(resp))
		return
	}

	var responses graphql.ResponseHandler
	responses, ctx = n.Exec.DispatchOperation(ctx, rc)
	m.RespondMsg(natsResponse(responses(ctx)))
}

func jsonDecode(r io.Reader, val interface{}) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(val)
}

func natsResponse(resp *graphql.Response) *nats.Msg {
	var data []byte
	var err error
	data, err = json.Marshal(resp)
	if err != nil {
		data = []byte(`{"error": "internal server error"}`)
	}

	return &nats.Msg{
		Data: data,
	}
}
