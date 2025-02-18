package app

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"html/template"
	"os"
	"path/filepath"
	"rulecat/utils"
	"rulecat/utils/email"
	"rulecat/utils/es"
	"rulecat/utils/kafka"
	log2 "rulecat/utils/log"
)

var (
	configFile  []byte
	emailSender *email.EmailConf
	kafkaP      *kafka.DataProducer
	ConfigG     Config
	esSvc       *es.ElasticSearchService
)

type Input struct {
	Kafka struct {
		Enabled bool     `yaml:"enabled"`
		Server  []string `yaml:"server"`
		GroupId string   `yaml:"group_id"`
		User    string   `yaml:"user"`
		Passwd  string   `yaml:"passwd"`
	}
}

type Output struct {
	Es struct {
		Enabled bool     `yaml:"enabled"`
		Server  []string `yaml:"es_host"`
		Version int      `yaml:"version"`
		User    string   `yaml:"user"`
		Passwd  string   `yaml:"passwd"`
	}
	Kafka struct {
		Enabled bool     `yaml:"enabled"`
		Server  []string `yaml:"server"`
		Topic   string   `yaml:"topic"`
		GroupId string   `yaml:"group_id"`
		User    string   `yaml:"user"`
		Passwd  string   `yaml:"passwd"`
	}

	Json struct {
		Enabled bool   `yaml:"enabled"`
		Path    string `yaml:"path"`
		Name    string `yaml:"name"`
	}
	Email struct {
		Enabled       bool   `yaml:"enabled"`
		EmailHost     string `yaml:"email_host"`
		EmailSmtpPort int    `yaml:"email_smtp_port"`
		EmailFrom     string `yaml:"email_from"`
		EmailUserName string `yaml:"email_username"`
		EmailPwd      string `yaml:"email_pwd"`
	}
}

type Config struct {
	Name   string `yaml:"name"`
	Env    string `yaml:"env"`
	InPut  Input  `yaml:"input"`
	OutPut Output `yaml:"output"`
}

func init() {
	var err error
	utils.ServerBanner()
	configFile, err = os.ReadFile(utils.GetCurrentPath() + "/etc/config.yml")
	if err != nil {
		log2.Error.Fatalf("Get yml file err %v ", err)
	}
	err = yaml.Unmarshal(configFile, &ConfigG)
	if err != nil {
		log2.Error.Fatalf("Unmarshal yml file err: %v ", err)
	}

	if ConfigG.OutPut.Email.Enabled {
		emailSender, err = email.New(ConfigG.OutPut.Email.EmailHost, ConfigG.OutPut.Email.EmailSmtpPort,
			ConfigG.OutPut.Email.EmailFrom,
			ConfigG.OutPut.Email.EmailUserName, ConfigG.OutPut.Email.EmailPwd)
		if err != nil {
			log2.Error.Fatalf("Create emailSender err: %v ", err)
		}
	}
	if ConfigG.OutPut.Kafka.Enabled {
		kafkaP = kafka.InitKafkaProducer(ConfigG.OutPut.Kafka.Server,
			ConfigG.OutPut.Kafka.GroupId, ConfigG.OutPut.Kafka.Topic,
			ConfigG.OutPut.Kafka.User, ConfigG.OutPut.Kafka.Passwd)

	}

	if ConfigG.OutPut.Es.Enabled {
		esConf := es.ElasticConfig{Url: ConfigG.OutPut.Es.Server,
			User:   ConfigG.OutPut.Es.User,
			Secret: ConfigG.OutPut.Es.Passwd,
			Sniff:  new(bool)}
		esSvc, err = es.CreateElasticSearchService(esConf, ConfigG.OutPut.Es.Version)
		if err != nil {
			log2.Error.Fatalf("Create elastic search service err: %v ", err)
		}
	}
}

func SendMail(data []byte) {
	if ConfigG.OutPut.Email.Enabled {
		tmp := `<html>
		<head>
			<meta charset="utf-8">
			<style>
				body {
					font-family: Arial, sans-serif;
					margin: 0;
					padding: 20px;
					background-color: #f4f4f9;
					color: #333;
				}
				h3 {
					margin: 10px 0;
					font-weight: bold;
				}
				pre {
					background-color: #f9f9f9;
					border: 1px solid #ddd;
					padding: 10px;
					border-radius: 5px;
					white-space: pre-wrap;
					word-wrap: break-word;
					margin-top: 10px;
				}
				.container {
					max-width: 800px;
					margin: 0 auto;
					background-color: #fff;
					padding: 20px;
					border-radius: 8px;
					box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
				}
				.header {
					text-align: center;
					margin-bottom: 20px;
				}
				.header h2 {
					margin: 0;
					font-size: 24px;
					color: #333;
				}
				.section {
					margin-bottom: 20px;
				}
				.section h3 {
					margin-top: 0;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>Alert Notification</h2>
				</div>
				<div class="section">
					<h3>rule name: {{.Data_.rule_name}}</h3>
					<h3>rule id: {{.Data_.rule_id}}</h3>
					<h3>rule tag: {{.Data_.rule_tag}}</h3>
					<h3>rule type: {{.Data_.rule_type}}</h3>
					<h3>threat level: {{.Data_.threat_level}}</h3>
				</div>
				<div class="section">
					<h4>data:</h4>
					<pre id="out_pre">{{.Json}}</pre>
				</div>
			</div>
		</body>
		</html>`

		type Args struct {
			Data_ map[string]interface{}
			Json  template.HTML
		}

		Data := make(map[string]interface{})
		_ = Json1.Unmarshal(data, &Data)
		Data1 := Args{Data_: Data, Json: template.HTML(utils.FormatJson(Data["data"]))}
		t := template.Must(template.New("mail").Parse(tmp))
		var tpl bytes.Buffer
		err := t.Execute(&tpl, Data1)
		if err != nil {
			log2.Error.Printf("Email template parse err: %v ", err)
			err = nil
		}
		var eAddr []string
		for _, v := range Data["e-mail"].([]interface{}) {
			eAddr = append(eAddr, v.(string))
		}
		to := email.ToSomeBody{To: eAddr, Cc: eAddr}
		err = emailSender.SendEmail(&to, "Alert-"+Data["rule_name"].(string), tpl.String())
		if err != nil {
			log2.Error.Printf("Email send err: %v ", err)
		}
	}

}

func SendKafka(message []byte) {

	if ConfigG.OutPut.Kafka.Enabled {
		err := kafkaP.AddMessage(message)
		if err != nil {
			log2.Error.Printf("Kafka message Delivery err: %v ", err)
		}
	}
}

func SendEs(typeName string, namespace string, sinkData string) {

	if ConfigG.OutPut.Es.Enabled {
		err := esSvc.AddBodyString(typeName, namespace, sinkData)
		if err != nil {
			log2.Error.Printf("Send es message err: %v ", err)
		}
	}
}

func SendJson(message []byte) {
	if ConfigG.OutPut.Json.Enabled {
		filePath := filepath.Join(ConfigG.OutPut.Json.Path, ConfigG.OutPut.Json.Name)

		dir := filepath.Dir(filePath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log2.Error.Printf("Failed to create directory: %v", err)
			return
		}
		err = utils.WriteFile(filePath, string(message)+"\n")
		if err != nil {
			log2.Error.Printf("Failed to write to file %s: %v", filePath, err)
			return
		}

		log2.Info.Printf("Successfully wrote message to file %s", filePath)
	}
}
