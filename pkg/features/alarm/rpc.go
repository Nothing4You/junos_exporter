// SPDX-License-Identifier: MIT

package alarm

import (
	"encoding/xml"
)

type result struct {
	XMLName xml.Name `xml:"alarm-information"`
	Details []details `xml:"alarm-detail"`
}

type details struct {
	Class       string `xml:"alarm-class"`
	Description string `xml:"alarm-description"`
	Type        string `xml:"alarm-type"`
}
