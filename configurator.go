package configuration

import (
	"errors"
	"reflect"
)

func New(
	cfgPtr interface{},
	providers []Provider,
	loggingEnabled bool,
	failIfCannotSet bool,
) (configurator, error) {
	if len(providers) == 0 {
		return configurator{}, errors.New("providers not found")
	}

	if reflect.TypeOf(cfgPtr).Kind() != reflect.Ptr {
		return configurator{}, errors.New("not a pointer to the struct")
	}

	gLoggingEnabled = loggingEnabled
	gFailIfCannotSet = failIfCannotSet

	return configurator{
		config:    cfgPtr,
		providers: providers,
	}, nil
}

type configurator struct {
	config    interface{}
	providers []Provider
}

func (c configurator) InitValues() {
	c.fillUp(c.config)
}

func (c configurator) fillUp(i interface{}, parentPath ...string) {
	var (
		t = reflect.TypeOf(i)
		v = reflect.ValueOf(i)
	)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		var (
			tField      = t.Field(i)
			vField      = v.Field(i)
			currentPath = append(parentPath, tField.Name)
		)

		if tField.Type.Kind() == reflect.Struct {
			c.fillUp(vField.Addr().Interface(), currentPath...)
			continue
		}

		if tField.Type.Kind() == reflect.Ptr && tField.Type.Elem().Kind() == reflect.Struct {
			vField.Set(reflect.New(tField.Type.Elem()))
			c.fillUp(vField.Interface(), currentPath...)
			continue
		}

		c.applyProviders(tField, vField, currentPath)
	}
}

func (c configurator) applyProviders(field reflect.StructField, v reflect.Value, currentPath []string) {
	logf("configurator: current path: %v", currentPath)

	for _, provider := range c.providers {
		if provider.Provide(field, v, currentPath...) {
			logf("\n")
			return
		}
	}
	logf("configurator: field [%s] with tags [%v] cannot be set!", field.Name, field.Tag)
	fail("configurator: field [%s] with tags [%v] cannot be set!", field.Name, field.Tag)
}
