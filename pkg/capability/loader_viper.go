package capability

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func WithViperConfigurationLoader(options ...ViperLoaderOption) fx.Option {
	return withViperLoader("configuration", options...)
}

func withViperLoader(name string, options ...ViperLoaderOption) fx.Option {
	provide := func() (Loader, error) {
		loader := newViperLoader(options...)
		if err := loader.Validate(); err != nil {
			return nil, err
		}
		return loader, nil
	}
	return fx.Provide(
		fx.Annotate(
			provide,
			fx.ResultTags(fmt.Sprintf(`name:"%s"`, name)),
		),
	)
}

type ViperLoaderOption func(*viperLoader)

func ViperLoaderWithFS(fs fs.FS) ViperLoaderOption {
	return func(loader *viperLoader) {
		loader.fs = fs
	}
}

func ViperLoaderWithPath(path *string) ViperLoaderOption {
	return func(loader *viperLoader) {
		fs := os.DirFS(*path)
		loader.fs = fs
	}
}

func ViperLoaderWithFileType(fileType string) ViperLoaderOption {
	return func(loader *viperLoader) {
		loader.fileType = fileType
	}
}

func ViperLoaderWithAlias(alias string, target string) ViperLoaderOption {
	return func(loader *viperLoader) {
		if loader.aliases == nil {
			loader.aliases = make(map[string]string)
		}
		loader.aliases[alias] = target
	}
}

type viperLoader struct {
	fs       fs.FS
	fileType string
	aliases  map[string]string
}

func (loader viperLoader) Validate() error {
	return validation.Errors{
		"fs":       validation.Validate(loader.fs, validation.Required),
		"fileType": validation.Validate(loader.fileType, validation.Required),
	}.Filter()
}

func newViperLoader(options ...ViperLoaderOption) *viperLoader {
	loader := &viperLoader{
		aliases: make(map[string]string),
	}
	for _, option := range options {
		option(loader)
	}
	return loader
}

func (loader *viperLoader) Load(name string, data interface{}) error {
	if _, ok := loader.aliases[name]; ok {
		name = loader.aliases[name]
	}
	reader := viper.New()
	reader.SetConfigName(name)
	reader.SetConfigType(loader.fileType)

	if file, err := loader.fs.Open(fmt.Sprintf("%s.%s", name, loader.fileType)); err != nil {
		return err
	} else {
		defer file.Close()
		if err = reader.ReadConfig(file); err != nil {
			return err
		}

		{
			envPrefix := strings.ReplaceAll(strings.ToUpper(name), ".", "_")
			reader.SetEnvPrefix(envPrefix)
			reader.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			reader.AutomaticEnv()
		}

		decoderConfigOption := func(config *mapstructure.DecoderConfig) {
			config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
				FiltersDecodeHook(),
				mapstructure.TextUnmarshallerHookFunc(),
			)
		}
		if err := reader.Unmarshal(data, decoderConfigOption); err != nil {
			return err
		}
		return nil
	}
}
