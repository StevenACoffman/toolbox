package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

type District struct {
	DistrictID         string      `json:"district_id"`
	Name               string      `json:"name"`
	DistrictProvidedID string      `json:"district_provided_id"`
	NcesID             string      `json:"nces_id"`
	Key                interface{} `json:"key"`
}

func main() {
	var objmap map[string]District

	// note the missing leading and trailing curl braces below are intentional!
	districtTmpl := `"hogwarts_{{.id}}": {
    "district_id": "wizardingworld",
    "name": "Hogwarts School of Witchcraft and Wizardry ({{.id}})",
    "district_provided_id": "hogwarts_{{.id}}",
    "nces_id": "hogwarts_{{.id}}",
    "key": null
  },
  "beauxbatons_{{.id}}": {
    "district_id": "wizardingworld",
    "name": "Beauxbatons Academy of Magic ({{.id}})",
    "district_provided_id": "beauxbatons_{{.id}}",
    "nces_id": "beauxbatons_{{.id}}",
    "key": null
  },
  "mischief_{{.id}}": {
    "district_id": "wizardingworld",
    "name": "Fred and George's School of Mischief ({{.id}})",
    "district_provided_id": "mischief_{{.id}}",
    "nces_id": "mischief_{{.id}}",
    "key": null
  },
  "khanlabschool_{{.id}}": {
    "district_id": "muggleworld",
    "name": "Khan Lab School ({{.id}})",
    "district_provided_id": "khanlabschool_{{.id}}",
    "nces_id": "khanlabschool_{{.id}}",
    "key": null
  }`
	// note the missing trailing "}" above is intentional!

	bracedBytes := amplify(districtTmpl, 3)
	err := json.Unmarshal(bracedBytes, &objmap)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(objmap)
}

// amplify - Given a template of json map values
// repeat the segment by the factor specified
// apply the current iteration to the template
func amplify(segmentTemplate string, factor int) []byte {
	var segments []string
	t := template.Must(template.New("").Parse(segmentTemplate))
	for i := 0; i <= factor; i++ {

		varMap := map[string]string{
			"id": strconv.Itoa(i),
		}

		buf := &bytes.Buffer{}
		if err := t.Execute(buf, varMap); err != nil {
			panic(err)
		}
		segments = append(segments, buf.String())
	}

	joined := strings.Join(segments, ",")
	braced := fmt.Sprintf("{%s}", joined)
	bracedBytes := []byte(braced)
	return bracedBytes
}
