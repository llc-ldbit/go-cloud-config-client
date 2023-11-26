package configService

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type ConfigUpdateHandler func(ServiceSetting)

type ConfigServiceManager struct {
	configs        map[string]ServiceSetting
	mu             *sync.Mutex
	updateHandlers map[string]ConfigUpdateHandler
	updateInterval time.Duration
	url            string
	serviceName    string
}

func NewConfigServiceManager(serviceName, url string, updateInterval time.Duration) (*ConfigServiceManager, error) {
	c := &ConfigServiceManager{
		url:            url,
		serviceName:    serviceName,
		updateInterval: updateInterval,
		updateHandlers: map[string]ConfigUpdateHandler{},
		mu:             new(sync.Mutex),
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *ConfigServiceManager) GetParam(key string) (string, bool) {
	c.mu.Lock()
	setting, ok := c.configs[key]
	c.mu.Unlock()
	return setting.Value, ok
}

func (c *ConfigServiceManager) init() error {
	config, err := c.request()
	if err != nil {
		return err
	}
	c.configs = config
	return nil
}

func (c *ConfigServiceManager) SetUpdateHandler(handler ConfigUpdateHandler, keys ...string) {
	c.mu.Lock()
	for _, k := range keys{
		c.updateHandlers[k] = handler
	}
	c.mu.Unlock()
}

func (c *ConfigServiceManager) Updater() {
	defer c.mu.TryLock()
	for {
		time.Sleep(c.updateInterval)
		config, err := c.request()
		if err != nil {
			log.Println("request to config service failed:", err)
			continue
		}
		c.mu.Lock()
		for _, newCfg := range config {
			oldConfig := c.configs[newCfg.Key]
			if newCfg.Updated.After(oldConfig.Updated) && newCfg.Value != oldConfig.Value {
				c.configs[newCfg.Key] = newCfg
				if f, ok := c.updateHandlers[newCfg.Key]; ok {
					f(newCfg)
				}
			}
		}
		c.mu.Unlock()
	}
}

func (c *ConfigServiceManager) FillConfigStruct(config any) error {
	v := reflect.ValueOf(config).Elem()
	s := v.Type()

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < s.NumField(); i++ {
		param := s.Field(i).Tag.Get("config-service")
		fieldType := s.Field(i).Type
		if param != "" {
			if newValueString, ok := c.configs[param]; ok {
				newValue, err := c.formatToStructType(newValueString.Value, fieldType.Kind())
				if err != nil {
					return fmt.Errorf("failed to convert string %s to struct field %s:%s", newValueString.Value, s.Field(i).Name, err)
				}
				v.Field(i).Set(reflect.ValueOf(newValue))
			} else {
				return fmt.Errorf("parameter %s not found", param)
			}
		}
	}
	return nil
}

func (c *ConfigServiceManager) request() (map[string]ServiceSetting, error) {
	req, err := http.NewRequest("GET", c.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to config service:%s", err)
	}
	req.Header.Set("SERVICE_NAME", c.serviceName)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to config service failed:%s", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	config := []ServiceSetting{}
	err = json.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize response from config service:%s", err)
	}
	settings := make(map[string]ServiceSetting, len(config))
	for _, s := range config {
		settings[s.Key] = s
	}
	return settings, nil
}

func (c *ConfigServiceManager) formatToStructType(val string, structType reflect.Kind) (interface{}, error) {
	switch structType {
	case reflect.Int:
		value, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		return value, nil
	case reflect.Bool:
		value, err := strconv.ParseBool(val)
		if err != nil {
			return nil, err
		}
		return value, nil
	case reflect.String:
		return val, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", structType)
	}
}
