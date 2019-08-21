# jira-to-md

## JIRA to MarkDown text format converter
Convert from JIRA text formatting to GitHub Flavored MarkDown.

## Credits
This fun toy was heavily inspired by the J2M project by Fokke Zandbergen (http://j2m.fokkezb.nl/). Major credit to Fokke (and other contributors) for establishing a lot of the fundamental RexExp patterns for this module to work.


## Supported Conversions

* Headers (H1-H6)
* Bold
* Italic
* Bold + Italic
* Un-ordered lists
* Ordered lists
* Programming Language-specific code blocks (with help from herbert-venancio)
* Inline preformatted text spans
* Un-named links
* Named links
* Monospaced Text
* ~~Citations~~ (currently Buggy)
* Strikethroughs
* Inserts
* Superscripts
* Subscripts
* Single-paragraph blockquotes

Not done:
* Tables
* Panels 


## How to Use

### Markdown String

```
**Some bold things**
*Some italic stuff*
## H2
<http://google.com>
```

### Atlassian Wiki MarkUp Syntax (JIRA)

We'll refer to this as the `jira` variable in the examples below.

```
*Some bold things**
_Some italic stuff_
h2. H2
[http://google.com]
```

### Examples

```
cat cmd/j2m/j2m.jira | j2m
```
