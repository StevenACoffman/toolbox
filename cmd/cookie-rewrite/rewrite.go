package main

import (
	"fmt"
	"regexp"
)

var source_txt = `__Secure-ID=123; Secure; Domain=example.com`

func main() {
	fmt.Printf("Experiment with regular expressions.\n")
	fmt.Printf("source text:\n")
	fmt.Println("--------------------------------")
	fmt.Printf("%s\n", source_txt)
	fmt.Println("--------------------------------")
	var config = make(map[string]string)
	//config["example.com"] = "localhost"
	config["*"] = "yep"

	// a regular expression
	fmt.Println(rewriteCookieProperty(source_txt, config))

}

// https://gist.github.com/elliotchance/d419395aa776d632d897
func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

/* config is:
false (default): disable cookie rewriting
String: new domain, for example cookieDomainRewrite: "new.domain". To remove the domain, use cookieDomainRewrite: "".
map: mapping of domains to new domains, use "*" to match all domains. For example keep one domain unchanged, rewrite one domain and remove other domains:
cookieDomainRewrite: {
  "unchanged.domain": "unchanged.domain",
  "old.domain": "new.domain",
  "*": ""
*  **cookieDomainRewrite**: rewrites domain of `set-cookie` headers. Possible values:
   * `false` (default): disable cookie rewriting
   * String: new domain, for example `cookieDomainRewrite: "new.domain"`. To remove the domain, use `cookieDomainRewrite: ""`.
   * Object: mapping of domains to new domains, use `"*"` to match all domains.
     For example keep one domain unchanged, rewrite one domain and remove other domains:
     ```
     cookieDomainRewrite: {
       "unchanged.domain": "unchanged.domain",
       "old.domain": "new.domain",
       "*": ""
     }
     ```
}
*/
func rewriteCookieProperty(header string, config map[string]string) string {

	re := regexp.MustCompile(`(?i)(\s*; Domain=)([^;]+)`)
	return ReplaceAllStringSubmatchFunc(re, header, func(groups []string) string {
		match, prefix, previousValue := groups[0], groups[1], groups[2]

		var newValue string
		if config[previousValue] != "" {
			newValue = config[previousValue]
		} else if config["*"] != "" {
			newValue = config["*"]
		} else {
			//no match, return previous value
			return match
		}
		if newValue != "" {
			//replace value
			return prefix + newValue
		} else {
			//remove value
			return ""
		}
	})
}