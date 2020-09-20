package decoder

import (
	"bytes"
	"log"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pgmitche/go-pic/cmd/pkg/copybook"
	"github.com/pgmitche/go-pic/cmd/pkg/template"
)

func TestDecoder_Unmarshal(t *testing.T) {
	tests := []struct {
		name string
		c    *copybook.Copybook
		in   string
	}{
		{
			name: "SuccessfullyParseBasicDummyCopybook",
			c:    copybook.New("dummy", template.CopyBook),
			in: `000600         10  DUMMY-1                PIC X.                  00000167
000610         10  DUMMY-2                PIC X(3).               00000168
000620         10  DUMMY-3                PIC 9(7).               00000169
000630         10  DUMMY-4                PIC 9(4).               00000170
000640         10  DUMMY-5                PIC XX.                 00000171
000650         10  DUMMY-6                PIC 9(7).               00000172
000660         10  DUMMY-7                PIC X(10).              00000173
000670         10  DUMMY-8                PIC X.                  00000174`,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.in), tt.c)
			require.NoError(t, err)

			var b bytes.Buffer
			err = tt.c.WriteToStruct(&b)
			require.NoError(t, err)

			log.Println(b.String())
			// YourCopybook contains a representation of your provided Copybook
			// type YourCopybook struct {
			// 	DUMMY1 string `pic:"1"`
			// 	DUMMY2 string `pic:"3"`
			// 	DUMMY3 int `pic:"7"`
			// 	DUMMY4 int `pic:"4"`
			// 	DUMMY5 string `pic:"2"`
			// 	DUMMY6 int `pic:"7"`
			// 	DUMMY7 string `pic:"10"`
			// 	DUMMY8 string `pic:"1"`
			// }
		})
	}
}

func Test_decoder_findDataRecord(t *testing.T) {
	tests := []struct {
		name string
		line string
		c    *copybook.Copybook
		want *copybook.Record
	}{
		{
			name: "BasicPICStringSingle",
			line: "000600         10  DUMMY-1                PIC X.                  00000167",
			c:    copybook.New("dummy", template.CopyBook),
			want: &copybook.Record{
				Num:     600,
				Level:   10,
				Name:    "DUMMY-1",
				Picture: reflect.String,
				Length:  1,
			},
		}, {
			name: "BasicPICStringParentheses",
			line: "000610         10  DUMMY-2                PIC X(3).               00000168",
			c:    copybook.New("dummy", template.CopyBook),
			want: &copybook.Record{
				Num:     610,
				Level:   10,
				Name:    "DUMMY-2",
				Picture: reflect.String,
				Length:  3,
			},
		}, {
			name: "BasicPICIntParentheses",
			line: "000620         10  DUMMY-3                PIC 9(7).               00000169",
			c:    copybook.New("dummy", template.CopyBook),
			want: &copybook.Record{
				Num:     620,
				Level:   10,
				Name:    "DUMMY-3",
				Picture: reflect.Int,
				Length:  7,
			},
		}, {
			name: "BasicPICStringMulti",
			line: "000640         10  DUMMY-5                PIC XX.                 00000171",
			c:    copybook.New("dummy", template.CopyBook),
			want: &copybook.Record{
				Num:     640,
				Level:   10,
				Name:    "DUMMY-5",
				Picture: reflect.String,
				Length:  2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (&decoder{}).findDataRecord(tt.line, tt.c)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
