// Copyright 2021 The Lokomotive Authors
// Copyright 2017 CoreOS, Inc
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

package terraform

import (
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

// handler defines signal handler.
type handler struct {
	logger *logrus.Entry
	// signalCh stores signals received from the system.
	signalCh chan os.Signal
}

// signalHandler creates a new Handler.
func signalHandler(logger *logrus.Entry) *handler {
	// Different signals that we are listening to. For the time being putting
	// only os.Interrupt, making a slice so that we can easily add more.
	signals := []os.Signal{os.Interrupt}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, signals...)

	// Goroutine that prints the message to the user in case the --verbose
	// option is not used.
	go func() {
		for range signalCh {
			logger.Warnln("Interrupt received, please wait for Terraform to terminate.")
		}
	}()

	return &handler{
		signalCh: signalCh,
		logger:   logger,
	}
}

func (h *handler) stop() {
	// Stop listening for interrupts on the channel
	signal.Stop(h.signalCh)
	// Close the channel.
	close(h.signalCh)
}
