// Copyright 2020 The Lokomotive Authors
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

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/assets"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Print configuration templates",
	Run:   runTemplate,
}

func init() {
	RootCmd.AddCommand(templateCmd)
}

func runTemplate(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl template",
		"args":    args,
	})

	walkList := func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err == nil {
			fmt.Println(fileName)
		}
		return err
	}
	walkPrint := func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err == nil {
			b, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			fmt.Printf("%s", b)
		}
		return err
	}
	if len(args) == 0 {
		if err := assets.Assets.WalkFiles("/examples", walkList); err != nil {
			ctxLogger.Fatalf("failed to walk assets: %s", err)
		}
	} else {
		if err := assets.Assets.WalkFiles("/examples/"+args[0], walkPrint); err != nil {
			ctxLogger.Fatalf("failed to walk assets: %s", err)
		}
	}
}
