package gcp_secretmanager

import (
	"fmt"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/gcp"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
	"os"
)

const secretmanagerKey = "gcp_secretmanager"

type secretmanagerSpec struct {
	SecretName string      `yaml:"gcp_secret_name"`
	Data       interface{} `yaml:"secret_data,omitempty"`
}

type smDriver struct{}

func (s *smDriver) Read(rawData []byte) (name string, data interface{}, err error) {
	secretSpec, err := utils.ReadYAML(rawData)
	if err != nil {
		return
	}
	name, ok := secretSpec["name"].(string)
	if !ok {
		err = fmt.Errorf("gcp_secretmanager: secret must contain string field 'name'")
		return
	}
	sp, ok := secretSpec["spec"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("gcp_secretmanager: secret must contain field 'spec'")
		return
	}
	specRaw, err := yaml.Marshal(sp)
	if err != nil {
		err = fmt.Errorf("gcp_secretmanager: can't parse secret '%v' spec %v", name, err)
		return
	}
	var spec secretmanagerSpec
	err = yaml.Unmarshal(specRaw, &spec)
	if err != nil {
		err = fmt.Errorf("gcp_secretmanager: can't parse secret '%v' spec %v", name, utils.ResolveYamlError(specRaw, err))
		return
	}
	if spec.SecretName == "" {
		err = fmt.Errorf("gcp_secretmanager: can't parse secret '%v', field 'spec.gcp_secret_name' is required", name)
		return
	}
	data, err = gcp.GetSecret(spec.SecretName)
	if err != nil {
		return "", nil, err
	}

	return
}

func (s *smDriver) Key() string {
	return secretmanagerKey
}

func init() {
	err := project.RegisterSecretDriver(&smDriver{}, secretmanagerKey)
	if err != nil {
		log.Fatalf("secrets: secretmanager driver init: %v", err.Error())
	}
}

func (s *smDriver) Edit(sec project.Secret) error {
	runner, err := executor.NewExecutor(config.Global.WorkingDir, &config.Interrupted)
	if err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	command := fmt.Sprintf("%s %s", editor, sec.Filename)
	err = runner.RunWithTty(command)
	if err != nil {
		return err
	}
	return nil
}

func (s *smDriver) Create(files map[string][]byte) error {
	runner, err := executor.NewExecutor(config.Global.WorkingDir, &config.Interrupted)
	if err != nil {
		return fmt.Errorf("create  secret: %v", err.Error())
	}
	if len(files) != 1 {
		return fmt.Errorf("create gcp secret: expected 1 file, received %v", len(files))
	}
	for fn, data := range files {
		filename, err := utils.SaveTmplToFile(fn, data)
		if err != nil {
			return fmt.Errorf("create gcp secret: %v", err.Error())
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		command := fmt.Sprintf("%s %s", editor, filename)
		err = runner.RunWithTty(command)
		if err != nil {
			os.RemoveAll(filename)
			return fmt.Errorf("secrets: create secret: %v", err.Error())
		}
	}
	return nil
}
