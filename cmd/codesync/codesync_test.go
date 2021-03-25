package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var Lines = []string{"\n", "\n", "a test\n", "more test\n", "\n"}

const ChecksumResult = 1043727889

func lineMerge(lines []string, lineNums []int) string {
	var linesToMerge []string
	for _, num := range lineNums {
		linesToMerge = append(linesToMerge, lines[num])
	}
	return strings.Join(linesToMerge, "")
}

func linePrefixAndSuffix(slices ...[]string) []string {
	var result []string
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

func TestProcess(t *testing.T) {
	type test struct {
		name     string
		lines    []string
		lineNums []int
		tag string
		want     uint32
	}

	tests := []test{
		{
			name: "simple checksum",
			lines: linePrefixAndSuffix(
				[]string{"# sync-start:mytag 0 somefile\n"},
				Lines, []string{"# sync-end:mytag\n"}),
			lineNums: []int{0, 1, 2, 3, 4, 5, 6},
			want:     ChecksumResult,
		},
		{
			name: "with jsx tags",
			lines: linePrefixAndSuffix(
				// sync-start:
				[]string{"{/* sync-start:mytag 0 somefile\n"},
				Lines,
				[]string{"{/* sync-end:mytag\n"}),
			lineNums: []int{0, 1, 2, 3, 4, 5, 6}, want: ChecksumResult,
		},
		{
			name: "with multiple start tags",
			lines: linePrefixAndSuffix([]string{
				"# sync-start:mytag 0 somefile\n",
				"# sync-start:mytag 0 otherfile\n",
				"# sync-start:mytag 0 thirdfile\n",
			},
				Lines,
				[]string{"# sync-end:mytag\n"}),
			lineNums: []int{0, 1, 2, 3, 4, 5, 6, 7, 8}, want: ChecksumResult,
		},
		{
			name: "with nested tags",
			lines: linePrefixAndSuffix(
				[]string{"# sync-start:mytag 0 somefile\n"},
				Lines[0:3],
				[]string{"# sync-start:othertag 0 otherfile\n", Lines[3], "# sync-end:othertag\n"},
				Lines[4:], []string{"# sync-end:mytag\n"}),
			lineNums: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			want:     ChecksumResult,
		},
		{
			name: "Sample",
			lines: []string{`

# sync-start:password-reset-token-constants 1289653574 services/users/passwords/reset_password.go
# Password reset tokens use shared.auth.tokens.  The first item in the tuple is
# a version.  Version 1 is (version, kaid, credential_version, user_nonce).
# TODO(benkraft): Using two nonces (credential_version and user_nonce) is
# unnecessarily complicated.  Make sure we expire the nonce whenever we would
# have expired the credential version, then get rid of the credential version.
_RESET_TOKEN_DATA_TYPE = 'password_reset'
_RESET_TOKEN_VERSION = 1
_RESET_TOKEN_LIFETIME = 24 * 60 * 60  # 1 day, in seconds
_RESET_NONCE_DATA_TYPE = 'pw_reset'
# sync-end:password-reset-token-constants

# When logging a change to a user settings property, the log instantiation
`},
tag: "password-reset-token-constants",
lineNums: []int{0},
			want: uint32(2134132571),
		},

	}

	for i, tc := range tests {
		tag := tc.tag
		if tag == ""{
			tag = "mytag"
		}
		testLines := lineMerge(tc.lines, tc.lineNums)
		fmt.Fprintln(os.Stderr, testLines)
		stringReader := strings.NewReader(testLines)
		idToMarker, err := processFile("<test lines>", stringReader, nil, true)
		fmt.Fprintln(os.Stderr, "idToMarker", idToMarker)
		if err != nil {
			t.Fatalf("%v tc# %v, got error: %v", tc.name, i, err)
		}

		got, ok := idToMarker[tag]
		if !ok {
			for key, value := range idToMarker {
				fmt.Fprint(os.Stderr, "id", key, "value", value.SourceBlockChecksum)
				break
			}
			t.Fatalf("%v tc# %v, missing tag in got %v", tc.name, i, idToMarker)
		}
		if tc.want != got.SourceBlockChecksum {
			fmt.Fprintf(os.Stderr, "<START>%v<END>\n", testLines)
			fmt.Fprintf(os.Stderr, "<STARTBLOCK>%v<ENDBLOCK>\n", got.SourceBlock)
			t.Fatalf("%v tc# %v expected: %v, got: %v from: %+v",
				tc.name, i, tc.want, got.SourceBlockChecksum, got)
		}
	}
}
