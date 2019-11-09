/*

Copyright (c) 2019, SILVANO ZAMPARDI
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.

*/

package unicode

import (
	"fmt"
	"html"
)

// CodepageIntToEmoji https://gist.github.com/nexus166/8197feef990ac4eb6f7b0366b28ee5a9
func CodepageIntToEmoji(i int) string {
	return html.UnescapeString("&#" + fmt.Sprintf("%d", i) + ";")
}
