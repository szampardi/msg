// COPYRIGHT (c) 2019-2021 SILVANO ZAMPARDI, ALL RIGHTS RESERVED.
// The license for these sources can be found in the LICENSE file in the root directory of this source tree.

package unicode

import (
	"fmt"
	"html"
)

// CodepageIntToEmoji https://gist.github.com/nexus166/8197feef990ac4eb6f7b0366b28ee5a9
func CodepageIntToEmoji(i int) string {
	return html.UnescapeString("&#" + fmt.Sprintf("%d", i) + ";")
}
