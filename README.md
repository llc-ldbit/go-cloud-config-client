# Клиент для взаимодействия с сервисом конфигураций 

## Пример использования:
```go
import "github.com/llc-ldbit/go-cloud-config-client"
import "log"

type Config struct {
    A string `config-service:"A"`
    B int `config-service:"B"`
    C bool `config-service:"C"`
}

func main() {
    var cfg Config
    // создаем нового клиента для сервиса service-name с интервалом обновления в 1 минуту
    cfgClient, err := configService.NewConfigServiceManager("service-name", "http://127.0.0.1:7654/api/v1/settings", time.Minute)

    // заполняем обьект конфига согласно тегам. Значения типа int, bool будут конвертированы автоматически
    if err := cfgClient.FillConfigStruct(&cfg); err != nil {
        log.Fatalln("failed to fill config struct:", err)
    }
    
    // Создаем обработчики обновления параметров
    cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		log.Println("A updated to", ss.Value)
	}, "A")

    cfgService.SetUpdateHandler(func(ss configService.ServiceSetting) {
		if ss.Key == "B"{
            log.Println("B updated to", ss.Value)
        } else if ss.Key == "C" {
            log.Println("C updated to", ss.Valude)
        }
	}, "B", "C")

    // запускаем цикл с обновлением
    go cfgService.Updater()
}   
```