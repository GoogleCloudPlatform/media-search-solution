// Copyright 2024 Google, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Author: rrmcguinness (Ryan McGuinness)
//         kingman (Charlie Wang)

package cloud

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/GoogleCloudPlatform/media-search-solution/pkg/cor"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// PubSubListener is a simple stateful wrapper around a subscription object.
// this allows for the easy configuration of multiple listeners. Since listeners
// life-cycles are outside the command life-cycle they are considered cloud components.
type PubSubListener struct {
	client       *pubsub.Client       // The Pub/Sub client.
	subscription *pubsub.Subscription // The Pub/Sub subscription.
	command      cor.Command          // The command to execute when a message is received.
}

// NewPubSubListener the constructor for PubSubListener
func NewPubSubListener(
	pubsubClient *pubsub.Client, // The Pub/Sub client.
	subscriptionID string, // The ID of the Pub/Sub subscription.
	command cor.Command, // The command to execute when a message is received.
) (cmd *PubSubListener, err error) {

	// Get the subscription from the Pub/Sub client.
	sub := pubsubClient.Subscription(subscriptionID)

	// Create a new PubSubListener.
	cmd = &PubSubListener{
		client:       pubsubClient,
		subscription: sub,
		command:      command,
	}
	return cmd, nil
}

// SetCommand A setter for the underlying handler command.
func (m *PubSubListener) SetCommand(command cor.Command) {
	// Only set the command if it's not already set.
	if m.command == nil {
		m.command = command
	}
}

// Listen starts the async function for listening and should be instantiated
// using the same context of the cloud service but may be configured independently
// for a different recovery life-cycle.
func (m *PubSubListener) Listen(ctx context.Context) {
	log.Printf("listening: %s", m.subscription)

	// Start a goroutine to listen for messages.
	go func() {
		// Create a new tracer.
		tracer := otel.Tracer("message-listener")

		// Receive messages from the subscription.
		err := m.subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
			// Start a new span.
			spanCtx, span := tracer.Start(ctx, "receive-message")
			span.SetName("receive-message")
			msgDataStr := string(msg.Data)
			span.SetAttributes(attribute.String("msg", msgDataStr))

			// Create a new chain context.
			chainCtx := cor.NewBaseContext()
			chainCtx.SetContext(spanCtx)
			chainCtx.Add(cor.CtxIn, msgDataStr)

			// Moving message acknowledgement to here tempurarily as the processing takes more than 600 seconds. which is the maximum time for a message to be acknowledged.
			// If this times out, the resize pipeline don't gets to run to completion, and messages are redelivered so we end up in an infinite loop.
			// TODO: decouple the message receiving from the command execution.
			msg.Ack()

			// Execute the command.
			m.command.Execute(chainCtx)

			// Only acknowledge the message if the command executed successfully.
			if !chainCtx.HasErrors() {
				span.SetStatus(codes.Ok, "success")
			} else {
				span.SetStatus(codes.Error, "failed")
				for _, e := range chainCtx.GetErrors() {
					log.Printf("error executing chain: %v", e)
				}
			}

			// End the span.
			span.End()
		})

		// Log any errors.
		if err != nil {
			log.Printf("error receiving data: %v", err)
		}
	}()
}
