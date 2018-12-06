// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
    "github.com/spf13/cobra"
)

var Root = "./drep"

var ptrAccount string
var ptrHeight int64
var ptrTo string
var ptrAmount string
var ptrChainId int64
var ptrDestChain int64
var ptrDataDir string
var ptrKeystore string
var ptrCode string
var ptrInput string
var ptrReadOnly bool

var flagAccount = "accounts"
var flagHeight = "height"
var flagTo = "to"
var flagAmount = "amount"
var flagChainId = "chain-id"
var flagDestChain = "dest-chain"
var flagDataDir = "data-dir"
var flagKeystore = "keystore"
var flagCode = "code"
var flagInput = "input"
var flagReadOnly = "read-only"

var CmdRoot = &cobra.Command{

    Use:   Root,

    Short: `"` + Root + `" is the prefix of all commands.`,

    Long:  `"` + Root + `" is the prefix of all commands. Any command you typed in should start with ` + Root + ".",

    Run: func(cmd *cobra.Command, args []string) {},
}