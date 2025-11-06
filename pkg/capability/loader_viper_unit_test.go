//go:build unit

package capability

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testConfiguration struct {
	Value string `json:"value"`
}

type loaderTestcase struct {
	fs                    fs.FS
	name                  string
	env                   map[string]string
	expectedConfiguration *testConfiguration
	expectingError        bool
}

var loaderTestcases = []loaderTestcase{
	{
		fs:   loader.FS,
		name: "test-1",
		env:  map[string]string{},
		expectedConfiguration: &testConfiguration{
			Value: "value-1",
		},
		expectingError: false,
	},
	{
		fs:   loader.FS,
		name: "test-1",
		env: map[string]string{
			"TEST-1_VALUE": "value-2",
		},
		expectedConfiguration: &testConfiguration{
			Value: "value-2",
		},
		expectingError: false,
	},
	{
		fs:   loader.FS,
		name: "sub.folder/test.1",
		env:  map[string]string{},
		expectedConfiguration: &testConfiguration{
			Value: "value-1",
		},
		expectingError: false,
	},
	{
		fs:   loader.FS,
		name: "sub.folder/test.1",
		env: map[string]string{
			"SUB_FOLDER/TEST_1_VALUE": "value-2",
		},
		expectedConfiguration: &testConfiguration{
			Value: "value-2",
		},
		expectingError: false,
	},
}

func TestLoader(t *testing.T) {
	for _, testcase := range loaderTestcases {
		t.Run(testcase.name, func(t *testing.T) {
			for k, v := range testcase.env {
				os.Setenv(k, v)
			}
			loader := newViperLoader(
				ViperLoaderWithFS(testcase.fs),
				ViperLoaderWithFileType("yaml"),
			)
			configuration := testConfiguration{}
			err := loader.Load(testcase.name, &configuration)
			if testcase.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if testcase.expectedConfiguration != nil {
				assert.Equal(t, *testcase.expectedConfiguration, configuration)
			}
		})
	}
}
