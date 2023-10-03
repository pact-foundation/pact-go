package avro

import (
	
	"os"
	"github.com/linkedin/goavro/v2"
)

func getCodec() *goavro.Codec {
	schema, err := os.ReadFile("user.avsc")
	if err != nil {
		panic(err)
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		panic(err)
	}

	return codec
}
