/*

Copyright (c) 2019, SILVANO ZAMPARDI
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

*/

package main

import (
	"os"

	"github.com/nexus166/msg"
)

func main() {
	msg.Debugf("%v\n", os.Args)
}
