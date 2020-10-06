package http

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// MatcherBySchema check if the request matching with the schema file
func MatcherBySchema(imposter Imposter) mux.MatcherFunc {
	return func(req *http.Request, rm *mux.RouteMatch) bool {

		if imposter.Request.SchemaFile == nil {
			return true
		}

		var err error
		switch imposter.Request.SchemaType {
		case "json":
			err = validateJSONSchema(imposter, req)
		case "xml":
			err = validateXMLSchema(imposter, req)
		default:
			err = validateJSONSchema(imposter, req)
		}

		// TODO: inject the logger
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
}

func validateJSONSchema(imposter Imposter, req *http.Request) error {
	var b []byte

	defer func() {
		req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}()

	schemaFile := imposter.CalculateFilePath(*imposter.Request.SchemaFile)
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		return fmt.Errorf("%w: the schema file %s not found", err, schemaFile)
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("%w: impossible read the request body", err)
	}

	contentBody := string(b)
	if contentBody == "" {
		return fmt.Errorf("unexpected empty body request")
	}

	dir, _ := os.Getwd()
	schemaFilePath := "file://" + dir + "/" + schemaFile
	schema := gojsonschema.NewReferenceLoader(schemaFilePath)
	document := gojsonschema.NewStringLoader(string(b))

	res, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return fmt.Errorf("%w: error validating the json schema", err)
	}

	if !res.Valid() {
		for _, desc := range res.Errors() {
			return errors.New(desc.String())
		}
	}

	return nil
}

func validateXMLSchema(imposter Imposter, req *http.Request) error {
	return fmt.Errorf("XML is still under construction")
}
