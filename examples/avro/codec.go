package avro

import (
	"io/ioutil"

	"github.com/linkedin/goavro/v2"
)

func getCodec() *goavro.Codec {
	schema, err := ioutil.ReadFile("user.avsc")
	if err != nil {
		panic(err)
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		panic(err)
	}

	return codec
}
