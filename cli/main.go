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

package main

import (
    "fmt"
    "strings"
    "helloworld/cli/cmd"
    "os"
)

func format(s string) string {
    ss := strings.Replace(s, "  ", " ", -1)
    for ss != s {
        s = ss
        ss = strings.Replace(ss, "  ", " ", -1)
    }
    for s[0] == ' ' {
        s = s[1:]
    }
    for s[len(s) - 1] == ' ' {
        s = s[:len(s) - 1]
    }
    return s
}

func process(args []string) {
    if len(args) == 0 {
        return
    }
    if args[0] != cmd.Root {
        fmt.Println()
        fmt.Println("Command should start with " + cmd.Root)
        fmt.Println()
        return
    }
    if len(args) == 1 {
        cmd.CmdRoot.Help()
        return
    }
    cmd.CmdRoot.ExecuteT(args[1:])
    return
}

func main() {
    args := os.Args
    process(args)
}